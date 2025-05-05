package business

import (
	"context"
	"io"
)

// Ports for drive adapters
type (
	// UserCases defines business cases related to user operations
	UserCases interface {
		CreateUser(context.Context, *User) error
		UpdateUser(context.Context, *User) error
		QueryUser(context.Context, UserID) (User, error)
	}

	// MessagesRelay defines a way to relay Message(s)
	MessagesRelay interface {
		RelayMessages(context.Context) error
	}
)

// Ports for driven adapters
type (
	// UserStore defines business cases related to user operations
	UserStore interface {
		CreateUser(context.Context, *User) error
		UpdateUser(context.Context, *User) error
		QueryUser(context.Context, UserID) (User, error)
	}

	// MessagesReader defines a way to read the pending Message(s)
	MessagesReader interface {
		io.Closer
		ReadMessages(context.Context, int) ([]Message, error)
	}

	// MessageSender defines a way to send a Message
	MessageSender interface {
		SendMessage(context.Context, ...Message) error
	}

	// MessageDeliveryConfirmer defines a way to confirm the delivery of a Message
	MessageDeliveryConfirmer interface {
		ConfirmMessageDelivery(context.Context, ...Message) error
	}
)
