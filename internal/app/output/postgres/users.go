//go:build http

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/yael-castro/goarch/internal/app/business"
	"log/slog"
	"reflect"
)

type UserStoreConfig struct {
	CreateUserTopic string
	UpdateUserTopic string
	Logger          *slog.Logger
	DB              *sql.DB
}

func NewUserStore(config UserStoreConfig) business.UserStore {
	return userStore{
		createUserTopic: config.CreateUserTopic,
		updateUserTopic: config.UpdateUserTopic,
		logger:          config.Logger,
		db:              config.DB,
	}
}

type userStore struct {
	createUserTopic string
	updateUserTopic string
	logger          *slog.Logger
	db              *sql.DB
}

func (s userStore) CreateUser(ctx context.Context, user *business.User) (err error) {
	// BEGIN
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		// ROLLBACK
		_ = tx.Rollback()
	}()

	// Building SQL record for User entity
	userSQL := NewUser(user)

	// Inserting purchase record
	err = s.createUser(ctx, tx, userSQL)
	if err != nil {
		return
	}

	// Inserting outbox message
	err = s.insertCreateUserMsg(ctx, tx, userSQL, s.createUserTopic)
	if err != nil {
		return
	}

	// Setting inserted user ID
	user.ID = business.UserID(userSQL.ID.Int64)

	// COMMIT
	return tx.Commit()
}

func (s userStore) createUser(ctx context.Context, tx *sql.Tx, userSQL *User) (err error) {
	const violateUniqueConstraint = "23505"

	// Inserting SQL record
	err = tx.QueryRowContext(
		ctx,
		insertUser,
		userSQL.Name,
		userSQL.Age,
		userSQL.Email,
	).Scan(&userSQL.ID)
	if err != nil {
		s.logger.InfoContext(ctx, "failed_user_insert", "error", err, "error_type", reflect.TypeOf(err), "user_id", userSQL.ID)

		// Error handling for postgres errors
		var pqErr *pq.Error

		if errors.As(err, &pqErr) && pqErr.Code == violateUniqueConstraint {
			err = fmt.Errorf("%w: email '%s' already exists", business.ErrDuplicateUserEmail, userSQL.Email.String)
			return
		}

		return
	}

	return
}

func (s userStore) insertCreateUserMsg(ctx context.Context, tx *sql.Tx, user *User, topic string) (err error) {
	value, err := user.MarshalBinary()
	if err != nil {
		return
	}

	msg := business.Message{
		Value: value,
		Topic: topic,
	}

	err = msg.Idempotent()
	if err != nil {
		return
	}

	stmt, args, err := insertOutboxMessage(NewMessage(msg))
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, stmt, args...)
	return
}

func (s userStore) UpdateUser(ctx context.Context, user *business.User) error {
	// BEGIN
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		// ROLLBACK
		_ = tx.Rollback()
	}()

	// Building SQL record for User entity
	userSQL := NewUser(user)

	err = s.updateUser(ctx, tx, userSQL)
	if err != nil {
		return err
	}

	// Inserting outbox message
	err = s.insertCreateUserMsg(ctx, tx, userSQL, s.updateUserTopic)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s userStore) updateUser(ctx context.Context, tx *sql.Tx, userSQL *User) (err error) {
	result, err := tx.ExecContext(
		ctx,
		updateUser,
		userSQL.Name,
		userSQL.Age,
		userSQL.Email,
		userSQL.ID,
	)
	if err != nil {
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return
	}

	if affected != 1 {
		err = fmt.Errorf("%w: unable to update user %d", business.ErrUserNotFound, userSQL.ID.Int64)
		return
	}

	return
}

func (s userStore) QueryUser(ctx context.Context, id business.UserID) (business.User, error) {
	var userSQL User

	err := s.db.QueryRowContext(
		ctx,
		selectUser,
		id,
	).Scan(
		&userSQL.ID,
		&userSQL.Name,
		&userSQL.Age,
		&userSQL.Email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("%w: user %d not found", business.ErrUserNotFound, id)
		}

		return business.User{}, err
	}

	return *userSQL.ToBusiness(), nil
}
