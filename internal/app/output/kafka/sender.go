package kafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/yael-castro/goarch/internal/app/business"
	"log/slog"
	"reflect"
	"sync"
	"time"
)

type MessageSenderConfig struct {
	Producer *kafka.Producer
	Logger   *slog.Logger
}

func NewMessageSender(config MessageSenderConfig) business.MessageSender {
	return &messageSender{
		producer: config.Producer,
		logger:   config.Logger,
	}
}

type messageSender struct {
	sync.Mutex
	producer *kafka.Producer
	logger   *slog.Logger
}

func (p *messageSender) SendMessage(ctx context.Context, messages ...business.Message) error {
	const maxWaitTime = 2 * time.Second

	ctx, cancel := context.WithTimeout(ctx, maxWaitTime)
	defer cancel()

	return p.sendMessage(ctx, messages...)
}

func (p *messageSender) sendMessage(ctx context.Context, messages ...business.Message) error {
	p.Lock()
	defer p.Unlock()

	// Trying to send Kafka's message
	deliveryChan := make(chan kafka.Event, len(messages))

	// Purge messages
	defer func() {
		err := p.producer.Purge(kafka.PurgeQueue)
		if err != nil {
			p.logger.WarnContext(ctx, "kafka_purge_queue", "error", err)
		}

		// TODO: find a way to confirm messages that are sent from the batch to prevent them from being sent twice in subsequent retries
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

		err := p.evaluateEvt(ctx, evt)
		if err != nil {
			return err
		}

		remaining--
	}

	return nil
}

// evaluateEvt evaluates the received event to know if there is an error
func (p *messageSender) evaluateEvt(ctx context.Context, evt kafka.Event) error {
	switch evt := evt.(type) {
	case *kafka.Message:
		if evt.TopicPartition.Error != nil {
			return evt.TopicPartition.Error
		}

		p.logger.InfoContext(ctx, "sent_kafka_message", "topic", *evt.TopicPartition.Topic, "partition", evt.TopicPartition.Partition, "offset", evt.TopicPartition.Offset)
		return nil // No error
	case kafka.Error:
		return fmt.Errorf("%w: communication issues '%v'", business.ErrMessageDeliveryFailed, evt)
	}

	p.logger.ErrorContext(ctx, "unknown_kafka_event", "event", evt, "event_type", reflect.TypeOf(evt).String())
	return fmt.Errorf("it seems that the message '%s' could not be sent", evt.String())
}
