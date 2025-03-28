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
	"log"
	"os"
	"time"
)

func New() Container {
	logger := log.New(os.Stdout, "[CONTAINER] ", log.LstdFlags)

	return &usersRelay{
		logger: logger,
	}
}

type usersRelay struct {
	container
	logger   *log.Logger
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
	errLogger := log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile)
	infoLogger := log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lshortfile)

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

	// Secondary adapters
	confirmer := postgres.NewMessageDeliveryConfirmer(db)

	reader := postgres.NewMessagesReader(db, infoLogger)

	sender := userskafka.NewMessageSender(userskafka.MessageSenderConfig{
		Info:     infoLogger,
		Error:    errLogger,
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
		Reader:      reader,
		Sender:      sender,
		Confirmer:   confirmer,
		InfoLogger:  infoLogger,
		ErrorLogger: errLogger,
	})
	if err != nil {
		return
	}

	// Primary adapters
	cmdRelay, err := command.Relay(messagesRelay, errLogger)
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
	r.mux.Lock()
	defer r.mux.Unlock()

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

func (r *usersRelay) injectCircuitBreaker(_ context.Context, breaker **gobreaker.CircuitBreaker[struct{}]) (err error) {
	const (
		maxHalfRequests        = 1
		maxConsecutiveFailures = 3
		openStateTimeout       = 10 * time.Second
		resetCounterInterval   = 10 * time.Second
	)

	logger := log.New(os.Stdout, "[CIRCUIT_BREAKER] ", log.LstdFlags)

	*breaker = gobreaker.NewCircuitBreaker[struct{}](gobreaker.Settings{
		Name:        "MessageSender",
		Timeout:     openStateTimeout,
		Interval:    resetCounterInterval,
		MaxRequests: maxHalfRequests,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > maxConsecutiveFailures
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Printf("%s: %s -> %s", name, from, to)
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
		r.logger.Println("Kafka producer is closed")
	}

	err = r.container.Close(ctx)
	if err != nil {
		r.logger.Println("Error trying to close container", err)
		return
	}

	r.logger.Println("Container is closed")
	return
}
