package command

import (
	"context"
	"errors"
	"github.com/yael-castro/goarch/internal/app/business"
	"log"
)

const (
	successExitCode = 0
	fatalExitCode   = 1
)

// Relay builds the command for message relay
func Relay(relay business.MessagesRelay, errLogger *log.Logger) (func(context.Context, ...string) int, error) {
	if relay == nil || errLogger == nil {
		return nil, errors.New("some dependencies are nil")
	}

	return func(ctx context.Context, _ ...string) int {
		err := relay.RelayMessages(ctx)
		if err != nil {
			errLogger.Println(err)
			return fatalExitCode
		}

		return successExitCode
	}, nil
}
