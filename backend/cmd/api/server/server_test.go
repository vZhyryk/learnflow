package server_test

import (
	"context"
	"learnflow_backend/cmd/api/app"
	"learnflow_backend/cmd/api/router"
	"learnflow_backend/cmd/api/server"
	"learnflow_backend/internal/shared/testutil"
	"os"
	"syscall"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func newTestServer(t *testing.T) (*server.Server, *app.App) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	a := &app.App{
		Logger: testutil.NewTestLogger(),
		Ctx:    ctx,
		Cancel: cancel,
	}
	a.Config.Port = 0 // OS-assigned free port
	a.Config.Timeouts.IdleTimeout = time.Second
	a.Config.Timeouts.ReadTimeout = time.Second
	a.Config.Timeouts.WriteTimeout = time.Second
	a.Config.Timeouts.ReadHeaderTimeout = time.Second

	handler := testutil.OkHandler()
	routeHandler := &router.RouteHandler{Router: handler, App: a}

	return server.NewServer(routeHandler, a), a
}

func waitForServe(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case serveErr := <-done:
		return serveErr
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return within 5s of SIGTERM — possible shutdown hang/goroutine leak")
		return nil
	}
}

func sendSIGTERM(t *testing.T) {
	t.Helper()
	So(syscall.Kill(os.Getpid(), syscall.SIGTERM), ShouldBeNil)
}

func TestServeGracefulShutdownCancelsContext(t *testing.T) {
	Convey("Serve", t, func() {
		Convey("On SIGTERM, it cancels the app context and shuts down cleanly", func() {
			srv, a := newTestServer(t)

			done := make(chan error, 1)
			go func() {
				done <- srv.Serve()
			}()

			// Give ListenAndServe time to start listening before signaling shutdown.
			time.Sleep(100 * time.Millisecond)
			So(a.Ctx.Err(), ShouldBeNil)

			sendSIGTERM(t)

			So(waitForServe(t, done), ShouldBeNil)
			So(a.Ctx.Err(), ShouldNotBeNil)
		})
	})
}

// TestServeGracefulShutdownWaitsForBackgroundWork proves that Serve waits for
// in-flight background work tracked via App.Wg before returning on SIGTERM.
func TestServeGracefulShutdownWaitsForBackgroundWork(t *testing.T) {
	Convey("Serve", t, func() {
		Convey("On SIGTERM, it waits for in-flight background work (App.Wg) before returning", func() {
			srv, a := newTestServer(t)

			workDone := make(chan struct{})
			a.Wg.Add(1)
			go func() {
				defer a.Wg.Done()
				<-a.Ctx.Done()
				time.Sleep(50 * time.Millisecond) // simulate brief in-flight cleanup work
				close(workDone)
			}()

			done := make(chan error, 1)
			go func() {
				done <- srv.Serve()
			}()

			time.Sleep(100 * time.Millisecond)
			sendSIGTERM(t)

			So(waitForServe(t, done), ShouldBeNil)

			select {
			case <-workDone:
				// background work completed before Serve returned — happens-before via App.Wg.Wait()
			default:
				t.Fatal("Serve returned before in-flight background work (App.Wg) completed")
			}
		})
	})
}
