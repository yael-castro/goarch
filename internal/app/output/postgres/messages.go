//go:build relay || tests

package postgres

import (
	"context"
	"database/sql"
	"github.com/yael-castro/goarch/internal/app/business"
	"log/slog"
)

func NewMessagesReader(db *sql.DB, logger *slog.Logger) business.MessagesReader {
	return messageReader{
		db:     db,
		logger: logger,
	}
}

type messageReader struct {
	db     *sql.DB
	logger *slog.Logger
}

func (p messageReader) ReadMessages(ctx context.Context, messages []business.Message) (int, error) {
	rows, err := p.db.QueryContext(ctx, selectPurchaseMessages, len(messages))
	if err != nil {
		return -1, err
	}
	defer func() {
		_ = rows.Close()
	}()

	index := -1

	for rows.Next() {
		index++

		message := Message{}

		var rawHeaders []byte

		err = rows.Scan(
			&message.ID,
			&message.Topic,
			&message.Key,
			&rawHeaders,
			&message.Value,
		)
		if err != nil {
			return -1, err
		}

		err = message.Headers.UnmarshalBinary(rawHeaders)
		if err != nil {
			return -1, err
		}

		messages[index] = *message.ToBusiness()
	}

	length := index + 1
	return length, nil
}

func (messageReader) Close() error {
	return nil
}

func NewMessageDeliveryConfirmer(db *sql.DB) business.MessageDeliveryConfirmer {
	return messageDeliveryConfirmer{
		db: db,
	}
}

type messageDeliveryConfirmer struct {
	db *sql.DB
}

func (m messageDeliveryConfirmer) ConfirmMessageDelivery(ctx context.Context, messages ...business.Message) (err error) {
	stmt, args, err := updatePurchaseMessages(messages)
	if err != nil {
		return
	}

	_, err = m.db.ExecContext(ctx, stmt, args...)
	return
}
