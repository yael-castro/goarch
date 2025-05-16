//go:build http

package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/yael-castro/goarch/internal/container"
	"github.com/yael-castro/goarch/internal/runtime"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Setting exit code
	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()

	// Building main context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer stop()

	// Building DI container
	c := container.New()

	// Injecting logger
	var logger *slog.Logger

	if err := c.Inject(ctx, &logger); err != nil {
		exitCode = 1
		return
	}

	slog.SetDefault(logger) // Setting default logger

	// Injecting dependencies
	var e *echo.Echo

	if err := c.Inject(ctx, &e); err != nil {
		slog.ErrorContext(ctx, "failed_server_built", "error", err)
		return
	}

	// Getting http port
	port := os.Getenv("PORT")
	if len(port) == 0 {
		const defaultPort = "8080"
		port = defaultPort
	}

	// Listening for shutdown gracefully
	shutdownCh := make(chan struct{}, 1)

	go func() {
		defer close(shutdownCh)

		<-ctx.Done()
		shutdown(c, e)

		shutdownCh <- struct{}{}
	}()

	// Running http server
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		slog.InfoContext(ctx, "http_server_is_running", "version", runtime.GitCommit, "port", port)
		errCh <- e.Start(":" + port)
	}()

	// Waiting for cancellation or error
	select {
	case <-ctx.Done():
		<-shutdownCh

	case err := <-errCh:
		stop()
		<-shutdownCh

		exitCode = 1

		slog.Error("", err)
	}
}

func shutdown(c container.Container, e *echo.Echo) {
	const gracePeriod = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), gracePeriod)
	defer cancel()

	// Closing http server
	err := e.Shutdown(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed_server_shutdown", "error", err)
		return
	}

	// Closing DI container
	err = c.Close(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed_container_shutdown", "error", err)
		return
	}

	slog.InfoContext(ctx, "success_shutdown")
}
