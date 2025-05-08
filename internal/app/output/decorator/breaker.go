package decorator

import (
	"context"
	"errors"
	"github.com/sony/gobreaker/v2"
	"github.com/yael-castro/goarch/internal/app/business"
)

func NewSenderBreaker(sender business.MessageSender, breaker *gobreaker.CircuitBreaker[struct{}]) (business.MessageSender, error) {
	if sender == nil || breaker == nil {
		return nil, errors.New("sender or breaker is nil")
	}

	return senderBreaker{
		breaker: breaker,
		sender:  sender,
	}, nil
}

type senderBreaker struct {
	breaker *gobreaker.CircuitBreaker[struct{}]
	sender  business.MessageSender
}

func (m senderBreaker) SendMessage(ctx context.Context, messages ...business.Message) (err error) {
	_, err = m.breaker.Execute(func() (struct{}, error) {
		return struct{}{}, m.sender.SendMessage(ctx, messages...)
	})

	return
}
