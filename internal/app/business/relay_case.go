//go:build relay

package business

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

type MessagesRelayConfig struct {
	Confirmer MessageDeliveryConfirmer
	Reader    MessagesReader
	Sender    MessageSender
	Logger    *slog.Logger
}

func (m MessagesRelayConfig) Validate() error {
	err := errors.New("some config is nil")

	if m.Confirmer == nil {
		return err
	}

	if m.Reader == nil {
		return err
	}

	if m.Sender == nil {
		return err
	}

	if m.Logger == nil {
		return err
	}

	return nil
}

func NewMessagesRelay(config MessagesRelayConfig) (MessagesRelay, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &messagesRelay{
		confirmer: config.Confirmer,
		reader:    config.Reader,
		sender:    config.Sender,
		logger:    config.Logger,
	}, nil
}

type messagesRelay struct {
	confirmer MessageDeliveryConfirmer
	reader    MessagesReader
	sender    MessageSender
	logger    *slog.Logger
}

func (m *messagesRelay) RelayMessages(ctx context.Context) (err error) {
	const messageLimit = 100

	length := 0
	messages := make([]Message, messageLimit)

	for {
		length, err = m.reader.ReadMessages(ctx, messages)
		if err != nil {
			return
		}

		// Waiting to poll more messages
		if length <= 0 {
			const retryDelay = 100 * time.Millisecond

			select {
			case <-ctx.Done():
				return errors.Join(err, m.reader.Close())
			case <-time.After(retryDelay): // TODO: explore alternative solutions to wait
				continue
			}
		}

		// Relaying messages...
		err = m.relayMessages(ctx, messages[:length])
		if err != nil && !errors.Is(err, ErrUnableToDeliverMessages) {
			return
		}
	}
}

func (m *messagesRelay) relayMessages(ctx context.Context, messages []Message) (err error) {
	m.logger.InfoContext(ctx, "relaying_messages", "messages", len(messages))

	err = m.sender.SendMessage(ctx, messages...)
	if err != nil {
		m.logger.InfoContext(ctx, "failed_sent_messages", "error", err)
		return
	}

	m.logger.InfoContext(ctx, "relayed_messages", "messages", len(messages))

	err = m.confirmer.ConfirmMessageDelivery(ctx, messages...)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed_confirmations", "error", err)
		return
	}

	m.logger.InfoContext(ctx, "confirmed_messages", "messages", len(messages))
	return
}
