package worker

import (
	"bytes"
	"context"
	"fmt"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/infrastructure/sanitizer"
	"learnflow_backend/internal/shared/testutil"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// panicNTimesWorker panics on its first n calls to Run, then blocks until ctx is
// done — simulating a real worker whose internal loop (e.g. Redis BLPop) runs
// forever once it stops erroring. Every call is reported non-blockingly on
// calledCh so tests can synchronize on "Run was invoked" without polling.
type panicNTimesWorker struct {
	n        int32
	calls    atomic.Int32
	calledCh chan int32
}

func (w *panicNTimesWorker) Run(ctx context.Context) {
	c := w.calls.Add(1)
	select {
	case w.calledCh <- c:
	default:
	}
	if c <= w.n {
		panic(fmt.Sprintf("induced panic #%d", c))
	}
	<-ctx.Done()
}

// ctxAwareWorker returns from Run as soon as ctx is done, simulating a blocking
// call (e.g. BLPop) that woke up because of shutdown rather than an error.
type ctxAwareWorker struct {
	calls atomic.Int32
}

func (w *ctxAwareWorker) Run(ctx context.Context) {
	w.calls.Add(1)
	<-ctx.Done()
}

// signalingWriter is an io.Writer that notifies sig (non-blockingly) after every
// Write, so a test can wait for "a log line arrived" via a channel instead of
// polling the buffer on a timer.
type signalingWriter struct {
	mu  sync.Mutex
	buf bytes.Buffer
	sig chan struct{}
}

func newSignalingWriter() *signalingWriter {
	return &signalingWriter{sig: make(chan struct{}, 1)}
}

func (w *signalingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	n, err := w.buf.Write(p)
	w.mu.Unlock()

	select {
	case w.sig <- struct{}{}:
	default:
	}
	return n, err
}

func (w *signalingWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

// waitForLogSubstring blocks until sw's buffer contains substr or the deadline
// elapses, failing the test in the latter case.
func waitForLogSubstring(t *testing.T, sw *signalingWriter, substr string, timeout time.Duration) {
	t.Helper()
	deadline := time.After(timeout)
	for {
		select {
		case <-sw.sig:
			if strings.Contains(sw.String(), substr) {
				return
			}
		case <-deadline:
			t.Fatalf("timed out waiting for log to contain %q", substr)
		}
	}
}

func TestRunWithRecoveryRestartsAfterPanic(t *testing.T) {
	Convey("Given a worker that panics twice before blocking on ctx.Done", t, func() {
		const n = 2
		w := &panicNTimesWorker{n: n, calledCh: make(chan int32, n+1)}
		log := testutil.NewTestLogger()
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		done := make(chan struct{})
		go func() {
			RunWithRecovery(ctx, log, w)
			close(done)
		}()

		Convey("Run is invoked n+1 times, surviving each panic via restart", func() {
			for i := int32(1); i <= n+1; i++ {
				select {
				case call := <-w.calledCh:
					So(call, ShouldEqual, i)
				case <-time.After(5 * time.Second):
					t.Fatalf("timed out waiting for call #%d", i)
				}
			}
			So(w.calls.Load(), ShouldEqual, n+1)

			cancel()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("RunWithRecovery did not return after ctx cancellation")
			}
		})
	})
}

func TestRunWithRecoveryShutdownDoesNotTriggerExtraRestart(t *testing.T) {
	Convey("Given ctx already cancelled before RunWithRecovery starts", t, func() {
		w := &ctxAwareWorker{}
		log := testutil.NewTestLogger()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		start := time.Now()
		RunWithRecovery(ctx, log, w)
		elapsed := time.Since(start)

		Convey("Run is called exactly once and no restartBackoff delay is paid", func() {
			So(w.calls.Load(), ShouldEqual, int32(1))
			So(elapsed, ShouldBeLessThan, restartBackoff)
		})
	})
}

func TestRunWithRecoveryLogsPanicWithStackTrace(t *testing.T) {
	Convey("Given a worker that panics exactly once", t, func() {
		log, buf := testutil.NewBufferLogger(logger.LevelInfo)
		w := &panicNTimesWorker{n: 1, calledCh: make(chan int32, 2)}
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		done := make(chan struct{})
		go func() {
			RunWithRecovery(ctx, log, w)
			close(done)
		}()

		Convey("the recovered panic and its stack trace are logged", func() {
			<-w.calledCh // first call: panics
			<-w.calledCh // second call: blocks on ctx.Done, proving the restart happened
			cancel()

			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("RunWithRecovery did not return after ctx cancellation")
			}

			logged := buf.String()
			So(logged, ShouldContainSubstring, "worker panic:")
			So(logged, ShouldContainSubstring, "induced panic #1")
			So(logged, ShouldContainSubstring, "goroutine")
		})
	})
}

// TestRunWithRecoveryProtectsOutboxPollerFromNilDBPanic wraps a real production
// worker (OutboxPoller) instead of a fake Worker stub. Wiring it with a nil
// db.QueryRunner makes its first poll() tick dereference a nil interface and
// panic — proving the supervisor actually shields this package's workers, not
// just the Worker interface contract in the abstract.
func TestRunWithRecoveryProtectsOutboxPollerFromNilDBPanic(t *testing.T) {
	Convey("Given an OutboxPoller wired with a nil db.QueryRunner", t, func() {
		sw := newSignalingWriter()
		log := logger.New(sw, sanitizer.NewSanitizer("***", 2000, nil), logger.LevelInfo)
		poller := NewOutboxPoller(nil, &mockPublisher{}, log, testutil.NoopTransactor{})

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		done := make(chan struct{})
		go func() {
			RunWithRecovery(ctx, log, poller)
			close(done)
		}()

		Convey("the nil-pointer panic inside Run is recovered and logged, not crashing the process", func() {
			waitForLogSubstring(t, sw, "worker panic:", 15*time.Second)

			cancel()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("RunWithRecovery did not return after ctx cancellation")
			}
		})
	})
}
