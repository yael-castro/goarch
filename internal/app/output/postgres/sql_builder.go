package postgres

import (
	"github.com/yael-castro/goarch/internal/app/business"
	"strconv"
	"strings"
)

func updatePurchaseMessages(messages []business.Message) (string, []any, error) {
	b := strings.Builder{}
	args := make([]interface{}, len(messages))

	b.WriteString(`UPDATE outbox_messages SET updated_at = now(), delivered_at = now() WHERE id IN (`)

	for index, msg := range messages {
		args[index] = msg.ID
		b.WriteString("$" + strconv.Itoa(index+1))

		if index != len(messages)-1 {
			b.WriteString(",")
		}
	}

	b.WriteRune(')')

	return b.String(), args, nil
}

func insertOutboxMessage(message Message) (string, []any, error) {
	const insertOutboxMessage = `
		INSERT INTO outbox_messages(topic, idempotency_key, partition_key, headers, value)
		VALUES ($1, $2, $3, $4, $5)
`

	rawHeaders, err := message.Headers.MarshalBinary()
	if err != nil {
		return "", nil, err
	}

	headers := NullBytes{
		V:     rawHeaders,
		Valid: len(rawHeaders) > 0,
	}

	args := []any{
		message.Topic,
		message.IdempotencyKey,
		message.Key,
		headers,
		message.Value,
	}

	return insertOutboxMessage, args, nil
}
