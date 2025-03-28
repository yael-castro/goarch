//go:build http

package main

import (
	"context"
	"github.com/labstack/echo/v4"
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
	var e *echo.Echo

	// Injecting dependencies
	c := container.New()

	if err := c.Inject(ctx, &e); err != nil {
		log.Println(err)
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

		log.Printf("Server http version '%s' is running on port '%s'\n", runtime.GitCommit, port)
		errCh <- e.Start(":" + port)
	}()

	// Waiting for cancellation or error
	select {
	case <-ctx.Done():
		stop()
		<-shutdownCh

	case err := <-errCh:
		stop()
		<-shutdownCh

		log.Fatal(err)
	}
}

func shutdown(c container.Container, e *echo.Echo) {
	const gracePeriod = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), gracePeriod)
	defer cancel()

	// Closing http server
	err := e.Shutdown(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Server shutdown gracefully")

	// Closing DI container
	err = c.Close(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("DI container gracefully closed")
}
