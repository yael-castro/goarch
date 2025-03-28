package kafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/yael-castro/goarch/internal/app/business"
	"log"
	"sync"
	"time"
)

type MessageSenderConfig struct {
	Producer *kafka.Producer
	Info     *log.Logger
	Error    *log.Logger
}

func NewMessageSender(config MessageSenderConfig) business.MessageSender {
	return &messageSender{
		producer: config.Producer,
		error:    config.Error,
		info:     config.Info,
	}
}

type messageSender struct {
	sync.Mutex
	producer *kafka.Producer
	info     *log.Logger
	error    *log.Logger
}

func (p *messageSender) SendMessage(ctx context.Context, message *business.Message) error {
	const maxWaitTime = 1_500 * time.Millisecond

	ctx, cancel := context.WithTimeout(ctx, maxWaitTime)
	defer cancel()

	return p.sendMessage(ctx, message)
}

func (p *messageSender) sendMessage(ctx context.Context, msg *business.Message) (err error) {
	message, err := NewMessage(msg)
	if err != nil {
		return
	}

	// Locking until message was sent
	p.Lock()
	defer p.Unlock()

	// Trying to send Kafka's message
	deliveryChan := make(chan kafka.Event, 1)

	err = p.producer.Produce(message, deliveryChan)
	if err != nil {
		return
	}

	// Waiting for message delivery
	var evt kafka.Event

	select {
	case <-ctx.Done():
		err := p.producer.Purge(kafka.PurgeQueue)
		if err != nil {
			p.error.Println("FAILED PURGE:", err)
		}

		return ctx.Err()
	case evt = <-deliveryChan:
		close(deliveryChan)
	}

	// Evaluating received event
	switch evt := evt.(type) {
	case *kafka.Message:
		if evt.TopicPartition.Error != nil {
			return evt.TopicPartition.Error
		}

		return
	case kafka.Error:
		return fmt.Errorf("%w: communication issues '%v'", business.ErrMessageDeliveryFailed, evt)
	default:
		p.error.Printf("Unknown event: %[1]T (%[1]T)\n", evt)
		return fmt.Errorf("it seems that the message %d could not be sent", msg.ID)
	}
}
