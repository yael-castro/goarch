package business

import "strconv"

// Supported values for Error
//
// WARNING: Do not change the order of error declarations.
const (
	_ Error = iota
	ErrInvalidUserID
	ErrInvalidUserAge
	ErrInvalidUserName
	ErrInvalidUserEmail
	ErrDuplicateUserEmail
	ErrUserNotFound
	ErrMessageDeliveryFailed
	ErrUnableToDeliverMessages
)

type Error uint8

func (e Error) Error() string {
	const errorPrefix = "E"
	return errorPrefix + strconv.FormatUint(uint64(e), 10)
}
