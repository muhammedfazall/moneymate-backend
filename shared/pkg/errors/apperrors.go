package apperrors

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	//general errors

	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("you are unauthorized, please login")
	ErrForbidden         = errors.New("forbidden")
	ErrInternal          = errors.New("internal server error")
	ErrDependencyFailure = errors.New("dependency failure")
	ErrBadRequest        = errors.New("bad request")

	// User,Auth specific

	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyTaken = errors.New("email already taken")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrOTPExpired        = errors.New("otp expired")
	ErrOTPInvalid        = errors.New("otp invalid")
	ErrOTPTimout         = errors.New("otp max tries reached")
	ErrOAuthFailure      = errors.New("oauth authentication failed")

	// Trading specific

	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidOrder      = errors.New("invalid order")
	ErrOrderNotFound     = errors.New("order not found")
	ErrMarketClosed      = errors.New("market is closed")
	ErrLimitExceeded     = errors.New("trading limit exceeded")
)

// custom error types
type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

func NewNotFoundError(resource string) error {
	return &NotFoundError{Resource: resource}
}

func MappDBErrors(err error) error {

	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			switch pgErr.ConstraintName {
			case "users_email_key":
				return ErrEmailAlreadyTaken
			case "orders_unique_id_key":
				return ErrInvalidOrder
			default:
				return ErrAlreadyExists
			}
		case "23503": // foreign key violation
			return fmt.Errorf("%w: %s", ErrInvalidInput, pgErr.Detail)
		}
	}
	return err
}
