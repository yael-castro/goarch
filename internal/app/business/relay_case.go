//go:build relay

package business

import (
	"context"
	"errors"
	"log"
	"time"
)

type MessagesRelayConfig struct {
	Confirmer   MessageDeliveryConfirmer
	Reader      MessagesReader
	Sender      MessageSender
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
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

	if m.InfoLogger == nil {
		return err
	}

	if m.ErrorLogger == nil {
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
		info:      config.InfoLogger,
		error:     config.ErrorLogger,
	}, nil
}

type messagesRelay struct {
	confirmer MessageDeliveryConfirmer
	reader    MessagesReader
	sender    MessageSender
	info      *log.Logger
	error     *log.Logger
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
	m.info.Printf("Relaying %d messages...\n", len(messages))

	err = m.sender.SendMessage(ctx, messages...)
	if err != nil {
		m.error.Printf("Failed to sent messages: %v", err)
		return
	}

	m.info.Printf("Relayed %d messages\n", len(messages))

	err = m.confirmer.ConfirmMessageDelivery(ctx, messages...)
	if err != nil {
		m.error.Printf("Failed to confirm messages: %v", err)
		return
	}

	m.info.Printf("Confirmed %d messages\n", len(messages))
	return
}
