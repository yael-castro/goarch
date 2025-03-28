package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/yael-castro/goarch/internal/app/business"
)

func NewMessage(message *business.Message) (msg *Message, err error) {
	msg = &Message{
		TopicPartition: TopicPartition{
			Topic: &message.Topic,
		},
		Key:   message.Key,
		Value: message.Value,
	}

	if len(message.Headers) > 0 {
		return
	}

	msg.Headers = make([]Header, len(message.Headers))

	for i, h := range message.Headers {
		msg.Headers[i] = (Header)(h) // TODO: avoid convert directly
	}

	return
}

// Aliases
type (
	Header         = kafka.Header
	Message        = kafka.Message
	TopicPartition = kafka.TopicPartition
)
