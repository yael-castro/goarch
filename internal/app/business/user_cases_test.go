package business_test

import (
	"context"
	"errors"
	"github.com/yael-castro/goarch/internal/app/business"
	"github.com/yael-castro/goarch/internal/app/business/mock"
	"reflect"
	"strconv"
	"testing"
)

func TestUserCases_CreateUser(t *testing.T) {
	cases := [...]struct {
		ctx         context.Context
		expectedErr error
		user        *business.User
	}{
		// Test case: name must be at least 4 characters
		{
			ctx:         context.Background(),
			user:        &business.User{},
			expectedErr: business.ErrInvalidUserName,
		},
		// Test case: invalid name by numbers
		{
			ctx: context.Background(),
			user: &business.User{
				Name: "1234",
			},
			expectedErr: business.ErrInvalidUserName,
		},
		// Test case: user can't be a baby
		{
			ctx: context.Background(),
			user: &business.User{
				Name: "Yael",
			},
			expectedErr: business.ErrInvalidUserAge,
		},
		// Test case: user is not alive
		{
			ctx: context.Background(),
			user: &business.User{
				Name: "Yael",
				Age:  120,
			},
			expectedErr: business.ErrInvalidUserAge,
		},
		// Test case: missing email
		{
			ctx: context.Background(),
			user: &business.User{
				Name: "Yael",
				Age:  23,
			},
			expectedErr: business.ErrInvalidUserEmail,
		},
		// Test case: Success!
		{
			ctx: context.Background(),
			user: &business.User{
				Name:  "Yael",
				Age:   23,
				Email: "contacto@yael.mx",
			},
		},
	}

	logic, err := business.NewUserCases(mock.UserStore{})
	if err != nil {
		t.Fatal(err)
	}

	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := logic.CreateUser(c.ctx, c.user)
			if !errors.Is(err, c.expectedErr) {
				t.Fatal(err)
			}

			if err != nil {
				t.Log(err)
				return
			}

			t.Logf("%+v", c.user)
		})
	}
}

func TestUserCases_UpdateUser(t *testing.T) {
	cases := [...]struct {
		ctx         context.Context
		expectedErr error
		user        *business.User
	}{
		// Test case: name must be at least 4 characters
		{
			ctx:         context.Background(),
			user:        &business.User{},
			expectedErr: business.ErrInvalidUserName,
		},
		// Test case: invalid name by numbers
		{
			ctx: context.Background(),
			user: &business.User{
				Name: "1234",
			},
			expectedErr: business.ErrInvalidUserName,
		},
		// Test case: user can't be a baby
		{
			ctx: context.Background(),
			user: &business.User{
				Name: "Yael",
			},
			expectedErr: business.ErrInvalidUserAge,
		},
		// Test case: user is not alive
		{
			ctx: context.Background(),
			user: &business.User{
				Name: "Yael",
				Age:  120,
			},
			expectedErr: business.ErrInvalidUserAge,
		},
		// Test case: missing email
		{
			ctx: context.Background(),
			user: &business.User{
				Name: "Yael",
				Age:  23,
			},
			expectedErr: business.ErrInvalidUserEmail,
		},
		// Test case: Success!
		{
			ctx: context.Background(),
			user: &business.User{
				Name:  "Yael",
				Age:   23,
				Email: "contacto@yael.mx",
			},
		},
	}

	logic, err := business.NewUserCases(mock.UserStore{})
	if err != nil {
		t.Fatal(err)
	}

	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := logic.UpdateUser(c.ctx, c.user)
			if !errors.Is(err, c.expectedErr) {
				t.Fatal(err)
			}

			if err != nil {
				t.Log(err)
				return
			}

			t.Logf("%+v", c.user)
		})
	}
}

func TestUserCases_QueryUser(t *testing.T) {
	cases := [...]struct {
		ctx          context.Context
		userID       business.UserID
		expectedUser business.User
		expectedErr  error
	}{
		// Test case: invalid user id
		{
			ctx:         context.Background(),
			userID:      0,
			expectedErr: business.ErrInvalidUserID,
		},
		// Test case: success
		{
			ctx:    context.Background(),
			userID: 1,
		},
	}

	logic, err := business.NewUserCases(mock.UserStore{})
	if err != nil {
		t.Fatal(err)
	}

	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			user, err := logic.QueryUser(c.ctx, c.userID)
			if !errors.Is(err, c.expectedErr) {
				t.Fatal(err)
			}

			if err != nil {
				t.Log(err)
				return
			}

			if !reflect.DeepEqual(user, c.expectedUser) {
				t.Fatalf("expected '%+v' got '%+v'", c.expectedUser, user)
			}

			t.Logf("%+v", user)
		})
	}
}
