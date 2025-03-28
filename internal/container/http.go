//go:build http

package container

import (
	"context"
	"database/sql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yael-castro/goarch/internal/app/business"
	http2 "github.com/yael-castro/goarch/internal/app/input/http"
	"github.com/yael-castro/goarch/internal/app/output/postgres"
	"github.com/yael-castro/goarch/pkg/env"
	"log"
	"os"
)

func New() Container {
	return new(handler)
}

type handler struct {
	container
}

func (h *handler) Inject(ctx context.Context, a any) error {
	switch a := a.(type) {
	case **echo.Echo:
		return h.injectHandler(ctx, a)
	}

	return h.container.Inject(ctx, a)
}

func (h *handler) injectHandler(ctx context.Context, e **echo.Echo) (err error) {
	// Getting environment variables
	createUserTopic, err := env.Get("CREATE_USER_TOPIC")
	if err != nil {
		return err
	}

	updateUserTopic, err := env.Get("UPDATE_USER_TOPIC")
	if err != nil {
		return err
	}

	// External dependencies
	var db *sql.DB

	if err = h.Inject(ctx, &db); err != nil {
		return err
	}

	// infoLogger := log.New(os.Stdout, "[INFO] ", log.LstdFlags)
	errLogger := log.New(os.Stderr, "[ERROR] ", log.LstdFlags)

	// Secondary adapters
	userStore := postgres.NewUserStore(postgres.UserStoreConfig{
		CreateUserTopic: createUserTopic,
		UpdateUserTopic: updateUserTopic,
		ErrLogger:       errLogger,
		DB:              db,
	})

	// Business logic
	userCases, err := business.NewUserCases(userStore)
	if err != nil {
		return err
	}

	// Primary adapters
	userHandler, err := http2.NewUserHandler(userCases)
	if err != nil {
		return err
	}

	n := echo.New()

	// Setting error handler
	n.HTTPErrorHandler = http2.ErrorHandler(n.HTTPErrorHandler)

	// Setting middlewares
	n.Use(middleware.Recover(), middleware.Logger())

	// Setting health checks
	dbCheck := func(ctx context.Context) error {
		return h.db.PingContext(ctx)
	}

	// Setting http routes
	http2.SetRoutes(n, userHandler, dbCheck)

	*e = n
	return
}
