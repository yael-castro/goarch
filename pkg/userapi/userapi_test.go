package userapi

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestUserAPI_Ping(t *testing.T) {
	cases := [...]struct {
		ctx         context.Context
		address     string
		expectedErr error
	}{
		{
			ctx:     context.Background(),
			address: "http://localhost:8080/v1/users",
		},
	}

	for _, c := range cases {
		t.Run(c.address, func(t *testing.T) {
			t.Log("Ping")

			client, err := New(c.address)
			if err != nil {
				t.Fatal(err)
			}

			err = client.Ping(c.ctx)
			if !errors.Is(err, c.expectedErr) {
				t.Fatalf("expected error: %v, got: %v", c.expectedErr, err)
			}

			t.Log("Pong!")
		})
	}
}

func TestUserAPI_GetUser(t *testing.T) {
	cases := [...]struct {
		ctx          context.Context
		userID       uint64
		address      string
		expectedErr  error
		expectedUser User
	}{
		{
			ctx:     context.Background(),
			address: "http://localhost:8080/v1/users",
			userID:  1,
		},
	}

	for _, c := range cases {
		t.Run(c.address, func(t *testing.T) {
			t.Log("Ping")

			client, err := New(c.address)
			if err != nil {
				t.Fatal(err)
			}

			user, err := client.GetUser(c.ctx, c.userID)
			if !errors.Is(err, c.expectedErr) {
				t.Fatalf("expected error: %v, got: %v", c.expectedErr, err)
			}

			if !reflect.DeepEqual(user, c.expectedUser) {
				t.Fatalf("expected user: %+v, got: %+v", c.expectedUser, user)
			}

			t.Log("Pong!")
		})
	}
}
