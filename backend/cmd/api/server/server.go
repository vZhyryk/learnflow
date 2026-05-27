// Package server starts and gracefully shuts down the HTTP server.
package server

import (
	"context"
	"errors"
	"fmt"
	"learnflow_backend/cmd/api/app"
	"learnflow_backend/cmd/api/router"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server wraps the HTTP server and application dependencies needed to serve and shut down.
type Server struct {
	Router *router.RouteHandler
	App    *app.App
}

// NewServer creates a Server with the given router and application container.
func NewServer(r *router.RouteHandler, a *app.App) *Server {
	return &Server{Router: r, App: a}
}

// Serve starts the HTTP server and blocks until a SIGINT or SIGTERM is received,
// then performs a graceful shutdown waiting up to 20 seconds for in-flight requests.
func (server *Server) Serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", server.App.Config.Port),
		Handler:      server.Router.Router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		// http.Server.ErrorLog requires *log.Logger; use the custom logger as its Writer
		ErrorLog:       log.New(server.App.Logger, "", 0),
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		server.App.Cancel()

		server.App.Logger.Info("caught signal", map[string]any{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
			return
		}

		server.App.Logger.Info("completing background tasks", map[string]any{
			"addr": srv.Addr,
		})

		server.App.Wg.Wait()
		shutdownError <- nil
	}()

	server.App.Logger.Info("starting server",
		map[string]any{
			"addr": srv.Addr,
			"env":  server.App.Config.Env,
		})

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server: listen: %w", err)
	}

	err = <-shutdownError
	if err != nil {
		return fmt.Errorf("server: shutdown: %w", err)
	}

	server.App.Logger.Info("stopped server", map[string]any{
		"addr": srv.Addr,
	})

	return nil
}
