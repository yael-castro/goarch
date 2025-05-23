package container

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/yael-castro/goarch/pkg/env"
	"log/slog"
	"os"
	"sync"
)

type Container interface {
	Inject(context.Context, any) error
	Close(context.Context) error
}

type container struct {
	sync.Mutex
	db     *sql.DB
	logger *slog.Logger
}

func (c *container) Inject(ctx context.Context, a any) error {
	switch a := a.(type) {
	case **sql.DB:
		return c.injectDB(ctx, a)
	case **slog.Logger:
		return c.injectLogger(ctx, a)
	}

	return fmt.Errorf("type \"%T\" is not supported", a)
}

func (c *container) Close(_ context.Context) error {
	c.Lock()
	defer c.Unlock()

	if c.db != nil {
		return c.db.Close()
	}

	return nil
}

func (c *container) injectLogger(ctx context.Context, logger **slog.Logger) (err error) {
	err = c.initLogger(ctx)
	if err != nil {
		return
	}

	*logger = c.logger
	return
}

func (c *container) initLogger(context.Context) (err error) {
	c.Lock()
	defer c.Unlock()

	if c.logger != nil {
		return
	}

	c.logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return
}

func (c *container) injectDB(ctx context.Context, db **sql.DB) (err error) {
	err = c.initDB(ctx)
	if err != nil {
		return
	}

	*db = c.db
	return err
}

func (c *container) initDB(ctx context.Context) (err error) {
	c.Lock()
	defer c.Unlock()

	if c.db != nil {
		return
	}

	dsn, err := env.Get("SQL_DSN")
	if err != nil {
		return
	}

	var newDB *sql.DB

	const driverName = "postgres"
	newDB, err = sql.Open(driverName, dsn)
	if err != nil {
		return
	}

	err = newDB.PingContext(ctx)
	if err != nil {
		return
	}

	c.db = new(sql.DB)
	c.db = newDB
	return
}
