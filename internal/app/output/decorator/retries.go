package decorator

import (
	"context"
	"errors"
	"github.com/sony/gobreaker/v2"
	"github.com/yael-castro/goarch/internal/app/business"
)

func NewSenderRetryer(sender business.MessageSender) (business.MessageSender, error) {
	if sender == nil {
		return nil, errors.New("sender is nil")
	}

	return senderRetryer{
		MessageSender: sender,
	}, nil
}

type senderRetryer struct {
	business.MessageSender
}

func (r senderRetryer) SendMessage(ctx context.Context, messages ...business.Message) (err error) {
	for {
		err = r.MessageSender.SendMessage(ctx, messages...)
		if !errors.Is(err, gobreaker.ErrOpenState) {
			return
		}
	}
}
