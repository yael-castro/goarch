package userapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yael-castro/goarch/pkg/jsont"
	"net/http"
	"net/url"
	"strconv"
)

func New(address string) (UserAPI, error) {
	parsed, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	return userAPI{
		client:  new(http.Client),
		address: parsed.Scheme + "://" + parsed.Host,
	}, nil
}

type UserAPI interface {
	Ping(context.Context) error
	GetUser(context.Context, uint64) (User, error)
}

type userAPI struct {
	address string
	client  *http.Client
}

func (u userAPI) GetUser(ctx context.Context, userID uint64) (User, error) {
	// Building request
	const endpointPath = "/v1/users/"

	req, err := http.NewRequest(http.MethodGet, u.address+endpointPath+strconv.FormatUint(userID, 10), nil)
	if err != nil {
		return User{}, err
	}

	// Doing request
	resp, err := u.client.Do(req.WithContext(ctx))
	if err != nil {
		return User{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Validating http status code
	if resp.StatusCode != http.StatusOK {
		return User{}, fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	// Decoding response
	var user User

	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return User{}, err
	}

	return user, err
}

func (u userAPI) Ping(ctx context.Context) error {
	// Building request
	const healthPath = "/v1/health"

	req, err := http.NewRequest(http.MethodGet, u.address+healthPath, nil)
	if err != nil {
		return err
	}

	response, err := u.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %d", response.StatusCode)
	}

	return nil // Success!
}

// User alias for jsont.User
type User = jsont.User
