package mock

import (
	"context"
	"github.com/yael-castro/goarch/internal/app/business"
	"log"
)

type MessageSender struct{}

func (MessageSender) SendMessage(ctx context.Context, message *business.Message) error {
	log.Printf("MESSAGE: %+v", message)
	return nil
}

type UserStore struct{}

func (UserStore) CreateUser(context.Context, *business.User) error {
	return nil
}

func (UserStore) UpdateUser(context.Context, *business.User) error {
	return nil
}

func (UserStore) QueryUser(context.Context, business.UserID) (business.User, error) {
	return business.User{}, nil
}
