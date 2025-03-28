package business

import (
	"context"
	"errors"
)

func NewUserCases(store UserStore) (UserCases, error) {
	if store == nil {
		return nil, errors.New("store is nil")
	}

	return userCases{
		store: store,
	}, nil
}

type userCases struct {
	store UserStore
}

func (p userCases) CreateUser(ctx context.Context, user *User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	return p.store.CreateUser(ctx, user)
}

func (p userCases) UpdateUser(ctx context.Context, user *User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	return p.store.UpdateUser(ctx, user)
}

func (p userCases) QueryUser(ctx context.Context, id UserID) (User, error) {
	if err := id.Validate(); err != nil {
		return User{}, err
	}

	return p.store.QueryUser(ctx, id)
}
