package main

import (
	"context"
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

	// Building main command
	c := container.New()

	// Injecting logger
	var logger *slog.Logger

	if err := c.Inject(ctx, &logger); err != nil {
		exitCode = 1
		return
	}

	slog.SetDefault(logger) // Setting a default logger

	// Injecting command
	var cmd func(context.Context, ...string) int

	if err := c.Inject(ctx, &cmd); err != nil {
		exitCode = 1
		slog.Error("failed_command_built", "error", err)
		return
	}

	// Listening for shutdown gracefully
	shutdownCh := make(chan struct{}, 1)

	go func() {
		// Waiting for close gracefully
		<-ctx.Done()

		// Shutting down
		const gracePeriod = 10 * time.Second

		ctx, cancel := context.WithTimeout(context.Background(), gracePeriod)
		defer cancel()

		_ = c.Close(ctx)

		// Confirm shutdown gracefully
		shutdownCh <- struct{}{}
		close(shutdownCh)
	}()

	// Executing message relay
	exitCodeCh := make(chan int, 1)

	go func() {
		defer close(exitCodeCh)

		slog.InfoContext(ctx, "message_relay_running", "version", runtime.GitCommit)
		exitCodeCh <- cmd(ctx)
	}()

	// Waiting for cancellation or exit code
	select {
	case <-ctx.Done():
	case exitCode = <-exitCodeCh:
		slog.InfoContext(ctx, "message_relay_exited", "exit_code", exitCode)
	}

	<-shutdownCh
}
