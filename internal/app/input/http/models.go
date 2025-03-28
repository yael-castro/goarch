package http

import (
	"github.com/yael-castro/goarch/internal/app/business"
	"github.com/yael-castro/goarch/pkg/jsont"
)

func NewUser(u *business.User) *User {
	return &User{
		ID:    int64(u.ID),
		Age:   uint8(u.Age),
		Name:  u.Name.String(),
		Email: u.Email.String(),
	}
}

type User jsont.User

func (u *User) ToBusiness() *business.User {
	return &business.User{
		ID:    business.UserID(u.ID),
		Age:   business.Age(u.Age),
		Name:  business.Name(u.Name),
		Email: business.Email(u.Email),
	}
}
