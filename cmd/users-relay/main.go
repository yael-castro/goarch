package main

import (
	"context"
	"github.com/yael-castro/goarch/internal/container"
	"github.com/yael-castro/goarch/internal/runtime"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Building main context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer stop()

	// Declaration of dependencies
	var cmd func(context.Context, ...string) int

	// Injecting dependencies
	c := container.New()

	err := c.Inject(ctx, &cmd)
	if err != nil {
		log.Println(err)
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

		log.Printf("Message relay version '%s' is running", runtime.GitCommit)
		exitCodeCh <- cmd(ctx)
	}()

	// Waiting for cancellation or exit code
	select {
	case <-ctx.Done():
		stop()
		<-shutdownCh

	case exitCode := <-exitCodeCh:
		stop()
		<-shutdownCh

		os.Exit(exitCode)
	}
}
