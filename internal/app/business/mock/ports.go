package mock

import (
	"context"
	"github.com/yael-castro/goarch/internal/app/business"
)

type MessageSender struct{}

func (MessageSender) SendMessage(context.Context, *business.Message) error {
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
