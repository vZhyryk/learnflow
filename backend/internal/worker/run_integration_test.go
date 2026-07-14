//go:build integration

package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	. "github.com/smartystreets/goconvey/convey"
)

var runTestSeq atomic.Int64

// uniqueEventType returns a collision-free Redis list key per test
func uniqueEventType(prefix string) string {
	return fmt.Sprintf("%s:%d:%d", prefix, time.Now().UnixNano(), runTestSeq.Add(1))
}

// newRunIntegrationWorker wires an EmailWorker[map[string]string] to the real docker-compose
// Redis (localhost:6379) and Postgres (via testutil.NewTestPool) so Run's full
// BLPop -> validate -> idempotency -> process -> retry -> DLQ loop can be exercised
// end-to-end, unlike newRealRedisEmailWorker, which only drives handleMessage in isolation.
func newRunIntegrationWorker(t *testing.T, eventType string, process func(map[string]string) error) (*EmailWorker[map[string]string], *pgxpool.Pool) {
	t.Helper()

	pool := testutil.NewTestPool(t)
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	t.Cleanup(func() {
		ctx := context.Background()
		//nolint:errcheck // best-effort cleanup of throwaway test data; errors intentionally ignored
		redisClient.Del(ctx, eventType)
		//nolint:errcheck // best-effort cleanup of throwaway test data; errors intentionally ignored
		pool.Exec(ctx, "DELETE FROM failed_jobs WHERE event_type = $1", eventType)
		//nolint:errcheck // Close's error is never actionable in test cleanup
		redisClient.Close()
	})

	w := &EmailWorker[map[string]string]{
		redisClient: redisClient,
		logger:      testutil.NewTestLogger(),
		dlq:         NewDLQ(pool, testutil.NewTestLogger()),
		cfg: Config[map[string]string]{
			EventType:      eventType,
			Validate:       func(_ map[string]string) error { return nil },
			IdempotencyKey: func(p map[string]string) string { return fmt.Sprintf("%s:%s", eventType, p["id"]) },
			Process: func(p map[string]string, _ string, _ Mailer) error {
				return process(p)
			},
		},
	}

	return w, pool
}

// pushMessage marshals payload and LPushes it onto eventType's list, the way
// a real producer would.
func pushMessage(t *testing.T, w *EmailWorker[map[string]string], eventType string, payload map[string]string) {
	t.Helper()

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("pushMessage: marshal: %v", err)
	}
	if err := w.redisClient.LPush(context.Background(), eventType, raw).Err(); err != nil {
		t.Fatalf("pushMessage: lpush: %v", err)
	}
}

// startRun launches w.Run in a goroutine and returns a cancel func plus a
// channel closed once Run has returned.
func startRun(w *EmailWorker[map[string]string]) (cancel context.CancelFunc, done <-chan struct{}) {
	ctx, cancelFn := context.WithCancel(context.Background())
	doneCh := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(doneCh)
	}()
	return cancelFn, doneCh
}

// stopRunAndWait cancels Run's context and fails the test if Run doesn't
// return afterward. The production Redis client (see redis.InitRedis) never
// sets ContextTimeoutEnabled, so an in-flight BLPop does not abort the moment
// ctx is cancelled — it only notices at its next loop iteration, after its own
// 5s BLPop timeout elapses. The wait here must exceed that 5s, or every
// cancellation would spuriously time out.
func stopRunAndWait(t *testing.T, cancel context.CancelFunc, done <-chan struct{}) {
	t.Helper()

	cancel()
	select {
	case <-done:
	case <-time.After(8 * time.Second):
		t.Fatal("Run did not return after ctx cancellation")
	}
}

// waitForDLQRow polls failed_jobs for eventType's row, returning its
// attempt_count and error_message once it appears. retry.Do's exponential
// backoff means the DLQ write can take a while to land — for attempts=3
// that's roughly 2s+4s+8s=14s of retry sleeps before Run calls dlq.Write.
func waitForDLQRow(t *testing.T, pool *pgxpool.Pool, eventType string) (attempts int, errMsg string) {
	t.Helper()

	deadline := time.After(180 * time.Second)
	for {
		select {
		case <-time.After(500 * time.Millisecond):
			row := pool.QueryRow(context.Background(),
				"SELECT attempt_count, error_message FROM failed_jobs WHERE event_type = $1", eventType)
			if err := row.Scan(&attempts, &errMsg); err == nil {
				return attempts, errMsg
			}
		case <-deadline:
			t.Fatal("timed out waiting for the DLQ row to appear")
			return 0, ""
		}
	}
}

func TestRun_ProcessesMessageSuccessfully_Integration(t *testing.T) {
	Convey("Run", t, func() {
		Convey("When BLPop receives a valid, not-yet-processed message, Process runs exactly once", func() {
			eventType := uniqueEventType("run-success")
			var callCount atomic.Int32
			processedCh := make(chan map[string]string, 1)

			w, _ := newRunIntegrationWorker(t, eventType, func(p map[string]string) error {
				callCount.Add(1)
				processedCh <- p
				return nil
			})

			payload := map[string]string{"id": "msg-1", "value": "hello"}
			t.Cleanup(func() {
				w.redisClient.Del(context.Background(), w.cfg.IdempotencyKey(payload))
			})
			pushMessage(t, w, eventType, payload)

			cancel, done := startRun(w)

			select {
			case got := <-processedCh:
				So(got, ShouldResemble, payload)
			case <-time.After(10 * time.Second):
				t.Fatal("timed out waiting for Process to be called")
			}
			So(callCount.Load(), ShouldEqual, int32(1))

			stopRunAndWait(t, cancel, done)
		})
	})
}

func TestRun_SkipsAlreadyProcessedMessage_Integration(t *testing.T) {
	Convey("Run", t, func() {
		Convey("When the idempotency key is already set, Process is never called", func() {
			eventType := uniqueEventType("run-duplicate")
			var callCount atomic.Int32

			w, _ := newRunIntegrationWorker(t, eventType, func(_ map[string]string) error {
				callCount.Add(1)
				return nil
			})

			payload := map[string]string{"id": "msg-dup", "value": "hello"}
			key := w.cfg.IdempotencyKey(payload)
			t.Cleanup(func() { w.redisClient.Del(context.Background(), key) })
			So(w.redisClient.SetNX(context.Background(), key, 1, time.Hour).Err(), ShouldBeNil)

			pushMessage(t, w, eventType, payload)
			cancel, done := startRun(w)

			// Run's BLPop timeout is 5s; sleeping past a full cycle proves the
			// message was drained from the list without Process ever firing.
			time.Sleep(6 * time.Second)
			So(callCount.Load(), ShouldEqual, int32(0))

			stopRunAndWait(t, cancel, done)
		})
	})
}

func TestRun_WritesToDLQAfterRetriesExhausted_Integration(t *testing.T) {
	Convey("Run", t, func() {
		Convey("When Process always fails, Run retries 3x then writes to the DLQ and clears the idempotency key", func() {
			eventType := uniqueEventType("run-dlq")
			var callCount atomic.Int32
			processErr := errors.New("process failed")

			w, pool := newRunIntegrationWorker(t, eventType, func(_ map[string]string) error {
				callCount.Add(1)
				return processErr
			})

			payload := map[string]string{"id": "msg-dlq", "value": "hello"}
			key := w.cfg.IdempotencyKey(payload)
			t.Cleanup(func() { w.redisClient.Del(context.Background(), key) })
			pushMessage(t, w, eventType, payload)

			cancel, done := startRun(w)

			attempts, errMsg := waitForDLQRow(t, pool, eventType)
			So(attempts, ShouldEqual, 3)
			So(errMsg, ShouldEqual, "retry: all 3 attempts failed: "+processErr.Error())
			So(callCount.Load(), ShouldEqual, int32(3))

			exists, err := w.redisClient.Exists(context.Background(), key).Result()
			So(err, ShouldBeNil)
			So(exists, ShouldEqual, int64(0))

			stopRunAndWait(t, cancel, done)
		})
	})
}
