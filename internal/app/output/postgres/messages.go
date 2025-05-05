//go:build relay || tests

package postgres

import (
	"context"
	"database/sql"
	"github.com/yael-castro/goarch/internal/app/business"
	"log"
)

func NewMessagesReader(db *sql.DB, info *log.Logger) business.MessagesReader {
	return messageReader{
		db:   db,
		info: info,
	}
}

type messageReader struct {
	db   *sql.DB
	info *log.Logger
}

func (p messageReader) ReadMessages(ctx context.Context, limit int) ([]business.Message, error) {
	rows, err := p.db.QueryContext(ctx, selectPurchaseMessages, limit)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	const expectedMessages = 100
	messages := make([]business.Message, 0, expectedMessages)

	for rows.Next() {
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
			return nil, err
		}

		err = message.Headers.UnmarshalBinary(rawHeaders)
		if err != nil {
			return nil, err
		}

		messages = append(messages, *message.ToBusiness())
	}

	return messages, nil
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
