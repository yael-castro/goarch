//go:build relay

package container

import (
	"context"
	"database/sql"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sony/gobreaker/v2"
	"github.com/yael-castro/goarch/internal/app/business"
	"github.com/yael-castro/goarch/internal/app/input/command"
	"github.com/yael-castro/goarch/internal/app/output/decorator"
	userskafka "github.com/yael-castro/goarch/internal/app/output/kafka"
	"github.com/yael-castro/goarch/internal/app/output/postgres"
	"github.com/yael-castro/goarch/pkg/env"
	"log/slog"
	"time"
)

func New() Container {
	return &usersRelay{
		logger: slog.Default(),
	}
}

type usersRelay struct {
	container
	logger   *slog.Logger
	producer *kafka.Producer
}

func (r *usersRelay) Inject(ctx context.Context, a any) (err error) {
	switch a := a.(type) {
	case *func(context.Context, ...string) int:
		return r.injectCommand(ctx, a)
	case **kafka.Producer:
		return r.injectProducer(ctx, a)
	case **gobreaker.CircuitBreaker[struct{}]:
		return r.injectCircuitBreaker(ctx, a)
	}

	return r.container.Inject(ctx, a)
}

func (r *usersRelay) injectCommand(ctx context.Context, cmd *func(context.Context, ...string) int) (err error) {
	// External dependencies
	var db *sql.DB
	if err = r.Inject(ctx, &db); err != nil {
		return
	}

	var producer *kafka.Producer
	if err = r.Inject(ctx, &producer); err != nil {
		return
	}

	var breaker *gobreaker.CircuitBreaker[struct{}]
	if err = r.Inject(ctx, &breaker); err != nil {
		return
	}

	var logger *slog.Logger
	if err = r.Inject(ctx, &logger); err != nil {
		return
	}

	// Secondary adapters
	confirmer := postgres.NewMessageDeliveryConfirmer(db)

	reader := postgres.NewMessagesReader(db, logger)

	sender := userskafka.NewMessageSender(userskafka.MessageSenderConfig{
		Logger:   logger,
		Producer: producer,
	})

	// Decorating secondary adapters
	sender, err = decorator.NewSenderBreaker(sender, breaker)
	if err != nil {
		return
	}

	sender, err = decorator.NewSenderRetryer(sender)
	if err != nil {
		return
	}

	// Business logic
	messagesRelay, err := business.NewMessagesRelay(business.MessagesRelayConfig{
		Reader:    reader,
		Sender:    sender,
		Confirmer: confirmer,
		Logger:    logger,
	})
	if err != nil {
		return
	}

	// Primary adapters
	cmdRelay, err := command.Relay(messagesRelay, logger)
	if err != nil {
		return
	}

	*cmd = cmdRelay
	return
}

func (r *usersRelay) injectProducer(ctx context.Context, producer **kafka.Producer) error {
	if err := r.initProducer(ctx); err != nil {
		return err
	}

	*producer = r.producer
	return nil
}

func (r *usersRelay) initProducer(_ context.Context) (err error) {
	r.Lock()
	defer r.Unlock()

	kafkaServers, err := env.Get("KAFKA_SERVERS")
	if err != nil {
		return err
	}

	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaServers,
		"acks":              "all",
	})
	if err != nil {
		return
	}

	r.producer = kafkaProducer
	return
}

func (r *usersRelay) injectCircuitBreaker(ctx context.Context, breaker **gobreaker.CircuitBreaker[struct{}]) (err error) {
	var logger *slog.Logger
	if err = r.Inject(ctx, &logger); err != nil {
		return err
	}

	const (
		maxHalfRequests        = 1
		maxConsecutiveFailures = 3
		openStateTimeout       = 10 * time.Second
		resetCounterInterval   = 10 * time.Second
	)

	*breaker = gobreaker.NewCircuitBreaker[struct{}](gobreaker.Settings{
		Name:        "MessageSender",
		Timeout:     openStateTimeout,
		Interval:    resetCounterInterval,
		MaxRequests: maxHalfRequests,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > maxConsecutiveFailures
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Info("circuit_status_change", "name", name, "old", from, "new", to)
		},
		IsSuccessful: func(err error) bool {
			return err == nil
		},
	})

	return
}

func (r *usersRelay) Close(ctx context.Context) (err error) {
	if r.producer != nil {
		r.producer.Close()
		r.logger.InfoContext(ctx, "kafka_producer_closed")
	}

	err = r.container.Close(ctx)
	if err != nil {
		r.logger.InfoContext(ctx, "container_is_not_close", "error", err)
		return
	}

	r.logger.InfoContext(ctx, "container_is_closed")
	return
}
