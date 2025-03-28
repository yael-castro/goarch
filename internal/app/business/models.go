package business

import (
	"fmt"
	"github.com/google/uuid"
	"unicode"
)

type User struct {
	ID    UserID
	Name  Name
	Email Email
	Age   Age
}

func (p User) Validate() error {
	if err := p.Name.Validate(); err != nil {
		return err
	}

	if err := p.Age.Validate(); err != nil {
		return err
	}

	return p.Email.Validate()
}

type Age uint8

func (a Age) Validate() error {
	const minAge, maxAge = 1, 100

	if a < minAge {
		return fmt.Errorf("%w: user can't be a baby", ErrInvalidUserAge)
	}

	if a > maxAge {
		return fmt.Errorf("%w: user is not alive", ErrInvalidUserAge)
	}

	return nil
}

type UserID uint64

func (u UserID) Validate() error {
	if u == 0 {
		return fmt.Errorf("%w: %d is not a valid user id", ErrInvalidUserID, u)
	}

	return nil
}

var _ fmt.Stringer = Email("")

type Email string

func (e Email) Validate() error {
	const minEmailLength = 3

	if len(e) < minEmailLength {
		return fmt.Errorf("%w: '%s' is not a valid email address", ErrInvalidUserEmail, e)
	}

	return nil
}

func (e Email) String() string {
	return string(e)
}

var _ fmt.Stringer = Name("")

type Name string

func (n Name) Validate() error {
	const minNameLength = 4

	if len(n) < minNameLength {
		return fmt.Errorf("%w: name must be at least %d characters", ErrInvalidUserName, minNameLength)
	}

	for _, char := range n {
		if !unicode.IsLetter(char) && !unicode.IsSpace(char) {
			return fmt.Errorf("%w: '%s' is not a letter", ErrInvalidUserName, string(char))
		}
	}

	return nil
}

func (n Name) String() string {
	return string(n)
}

type Header struct {
	Key   string
	Value []byte
}

type Headers = []Header

type Message struct {
	ID             uint64
	Topic          string
	Key            []byte
	Value          []byte
	IdempotencyKey []byte
	Headers        Headers
}

func (m *Message) Idempotent() (err error) {
	idempotencyKey, err := uuid.NewV7()
	if err != nil {
		return
	}

	m.IdempotencyKey, err = idempotencyKey.MarshalBinary()
	if err != nil {
		return
	}

	// Setting idempotency key in header
	m.Headers = append(m.Headers, Header{
		Key:   "idempotency_key",
		Value: m.IdempotencyKey,
	})

	return
}
