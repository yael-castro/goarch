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

func (p *messageSender) SendMessage(ctx context.Context, messages ...business.Message) error {
	const maxWaitTime = 1_500 * time.Millisecond

	ctx, cancel := context.WithTimeout(ctx, maxWaitTime)
	defer cancel()

	return p.sendMessage(ctx, messages...)
}

func (p *messageSender) sendMessage(ctx context.Context, messages ...business.Message) error {
	// Locking until message was sent
	p.Lock()
	defer p.Unlock()

	// Trying to send Kafka's message
	deliveryChan := make(chan kafka.Event, len(messages))

	// Purge messages
	defer func() {
		err := p.producer.Purge(kafka.PurgeQueue)
		if err != nil {
			p.error.Println("PURGE:", err)
		}

		close(deliveryChan)
	}()

	for i := range messages {
		message, err := NewMessage(&messages[i])
		if err != nil {
			return err
		}

		err = p.producer.Produce(message, deliveryChan)
		if err != nil {
			return err
		}
	}

	// Waiting for message delivery
	remaining := len(messages)

	for remaining > 0 {
		var evt kafka.Event

		select {
		case <-ctx.Done():
			return ctx.Err()
		case evt = <-deliveryChan:
		}

		err := p.evaluateEvt(evt)
		if err != nil {
			return err
		}

		remaining--
	}

	return nil
}

// evaluateEvt evaluates the received event to know if there is an error
func (p *messageSender) evaluateEvt(evt kafka.Event) error {
	switch evt := evt.(type) {
	case *kafka.Message:
		if evt.TopicPartition.Error != nil {
			return evt.TopicPartition.Error
		}

		p.info.Printf("Message in topic %s[%d] at offset %d", *evt.TopicPartition.Topic, evt.TopicPartition.Partition, evt.TopicPartition.Offset)
		return nil // No error
	case kafka.Error:
		return fmt.Errorf("%w: communication issues '%v'", business.ErrMessageDeliveryFailed, evt)
	}

	p.error.Printf("Unknown event: %[1]T (%[1]T)\n", evt)
	return fmt.Errorf("it seems that the message '%s' could not be sent", evt.String())
}
