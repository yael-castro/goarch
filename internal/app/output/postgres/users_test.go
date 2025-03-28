//go:build tests && http

package postgres_test

import (
	"context"
	"database/sql"
	"github.com/yael-castro/goarch/internal/app/business"
	"github.com/yael-castro/goarch/internal/container"
	"log"
	"os"
	"strconv"
	"testing"
)

func TestUserStore_CreateUser(t *testing.T) {
	cases := [...]struct {
		ctx  context.Context
		user *business.User
	}{
		{
			ctx: context.Background(),
			user: &business.User{
				ID: 1_000,
			},
		},
	}

	ctx := context.Background()

	// Establishing connection with DB
	var db *sql.DB

	di := container.New()

	err := di.Inject(ctx, &db)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = di.Close(context.Background())
	})

	errLogger := log.New(os.Stderr, "[ERROR] ", log.LstdFlags)

	store := NewUserStore(UserStoreConfig{
		DB:        db,
		UserTopic: "",
		ErrLogger: errLogger,
	})

	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ctx := c.ctx

			err := store.CreateUser(ctx, c.user)
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() {
				// TODO
			})

			t.Logf("User: %+v", c.user)
		})
	}
}
