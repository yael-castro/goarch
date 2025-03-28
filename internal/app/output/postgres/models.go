package postgres

import (
	"database/sql"
	"encoding/json"
	"github.com/yael-castro/goarch/internal/app/business"
	"github.com/yael-castro/goarch/pkg/jsont"
)

type NullBytes = sql.Null[[]byte]

func NewUser(u *business.User) *User {
	return &User{
		ID: sql.NullInt64{
			Int64: int64(u.ID),
			Valid: u.ID > 0,
		},
		Name: sql.NullString{
			String: u.Name.String(),
			Valid:  len(u.Name) > 0,
		},
		Email: sql.NullString{
			String: u.Email.String(),
			Valid:  len(u.Email) > 0,
		},
		Age: sql.NullInt64{
			Int64: int64(u.Age),
			Valid: u.Age > 0,
		},
	}
}

type User struct {
	ID    sql.NullInt64
	Name  sql.NullString
	Email sql.NullString
	Age   sql.NullInt64
}

func (u *User) ToBusiness() *business.User {
	return &business.User{
		ID:    business.UserID(u.ID.Int64),
		Age:   business.Age(u.Age.Int64),
		Name:  business.Name(u.Name.String),
		Email: business.Email(u.Email.String),
	}
}

func (u *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(jsont.User{
		ID:    u.ID.Int64,
		Name:  u.Name.String,
		Email: u.Email.String,
		Age:   uint8(u.Age.Int64),
	})
}

func NewHeader(header business.Header) Header {
	return (Header)(header)
}

type Header struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

func (h Header) ToBusiness() business.Header {
	return (business.Header)(h)
}

type Headers []Header

func (h *Headers) MarshalBinary() (data []byte, err error) {
	if len(*h) < 1 {
		return
	}

	return json.Marshal(h)
}

func (h *Headers) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return nil
	}

	return json.Unmarshal(data, h)
}

func NewMessage(m business.Message) (message Message) {
	message = Message{
		ID: sql.NullInt64{
			Int64: int64(m.ID),
			Valid: m.ID > 0,
		},
		Topic: sql.NullString{
			String: m.Topic,
			Valid:  len(m.Topic) > 0,
		},
		Key: NullBytes{
			V:     m.Key,
			Valid: len(m.Key) > 0,
		},
		Value: NullBytes{
			V:     m.Value,
			Valid: len(m.Value) > 0,
		},
		IdempotencyKey: NullBytes{
			V:     m.IdempotencyKey,
			Valid: len(m.IdempotencyKey) > 0,
		},
	}

	if len(m.Headers) < 1 {
		return
	}

	message.Headers = make(Headers, len(m.Headers))

	for i, header := range m.Headers {
		message.Headers[i] = NewHeader(header)
	}

	return
}

type Message struct {
	ID             sql.NullInt64
	Topic          sql.NullString
	Headers        Headers
	Key            NullBytes
	Value          NullBytes
	IdempotencyKey NullBytes
}

func (m *Message) ToBusiness() (message *business.Message) {
	message = &business.Message{
		ID:             uint64(m.ID.Int64),
		IdempotencyKey: m.IdempotencyKey.V,
		Topic:          m.Topic.String,
		Key:            m.Key.V,
		Value:          m.Value.V,
	}

	if len(m.Headers) < 1 {
		return
	}

	message.Headers = make([]business.Header, len(m.Headers))

	for i, header := range m.Headers {
		message.Headers[i] = header.ToBusiness()
	}

	return
}
