//go:build tests && relay

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/yael-castro/goarch/internal/container"
	"strconv"
	"testing"
)

func TestMessageReader_ReadMessages(t *testing.T) {
	cases := [...]struct {
		ctx         context.Context
		expectedErr error
	}{
		{
			ctx: context.Background(),
		},
	}

	var db *sql.DB
	c := container.New()
	ctx := context.Background()

	err := c.Inject(ctx, &db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = c.Close(ctx)
	})

	reader := NewMessagesReader(db)

	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			messages, err := reader.ReadMessages(c.ctx)
			if !errors.Is(err, c.expectedErr) {
				t.Fatalf("expected error '%v' unexpected error '%v'", err, c.expectedErr)
			}

			if err != nil {
				t.Skip(err)
			}

			// TODO: compare results
			for _, msg := range messages {
				t.Logf("Message: %+v", string(msg.Value))
			}
		})
	}

}
