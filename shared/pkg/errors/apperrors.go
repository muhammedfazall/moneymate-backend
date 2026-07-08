package apperrors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// General Errors
var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("you are unauthorized, please login")
	ErrForbidden         = errors.New("forbidden")
	ErrInternal          = errors.New("internal server error")
	ErrDependencyFailure = errors.New("dependency failure")
	ErrBadRequest        = errors.New("bad request")
)

// User , Auth Specific
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyTaken = errors.New("email already taken")
	ErrPhoneAlreadyTaken = errors.New("phone number already taken")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrOTPExpired        = errors.New("otp expired")
	ErrOTPInvalid        = errors.New("otp invalid")
	ErrOTPTimout         = errors.New("otp max tries reached")
	ErrOAuthFailure      = errors.New("oauth authentication failed")
)

// Financial & Transaction Specific
var (
	ErrInsufficientFunds  = errors.New("insufficient funds for this transaction")
	ErrTransactionLocked  = errors.New("transaction is currently locked or processing")
	ErrDailyLimitReached  = errors.New("daily transaction limit reached")
	ErrIdempotencyKeyUsed = errors.New("this transaction has already been processed")
)

//jwt token errors
var(
	ErrTokenExpired = errors.New("token expired")
    ErrInvalidToken = errors.New("invalid token")
)

// AppError represents a structured HTTP error safely returned to the frontend.
type AppError struct {
	StatusCode int    `json:"-"`       // Used by Fiber, never sent in JSON body
	Code       string `json:"code"`    // e.g., "INSUFFICIENT_FUNDS"
	Message    string `json:"message"` // Safe for the user to read
	Err        error  `json:"-"`       // The raw internal error for your server logs
}

// Error implements the standard Go error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap allows standard errors.Is and errors.As to work with AppError.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError is a helper for creating structured HTTP errors.
func NewAppError(statusCode int, code, message string, err error) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Err:        err,
	}
}

// MapDBErrors translates raw Postgres errors into domain sentinel errors.
func MapDBErrors(err error) error {
	if err == nil {
		return nil
	}

	// Strictly using pgx/v5 NoRows
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			switch pgErr.ConstraintName {
			case "users_email_key":
				return ErrEmailAlreadyTaken
			case "users_phone_number_key":
				return ErrPhoneAlreadyTaken
			case "transactions_idempotency_key":
				return ErrIdempotencyKeyUsed
			default:
				return ErrAlreadyExists
			}
		case "23503":
			return fmt.Errorf("%w: %s", ErrInvalidInput, pgErr.Detail)
		case "23514": // check_violation
			if pgErr.ConstraintName == "wallets_balance_check" {
				return ErrInsufficientFunds
			}
			return ErrInvalidInput
		case "40001":
			return ErrTransactionLocked
		}
	}

	return err
}

// ParseError looks at a domain error and packages it into a safe AppError for the frontend.
func ParseError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr // It's already an AppError, return it directly
	}
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrUserNotFound):
		return NewAppError(http.StatusNotFound, "NOT_FOUND", "The requested resource was not found.", err)

	case errors.Is(err, ErrEmailAlreadyTaken):
		return NewAppError(http.StatusConflict, "EMAIL_TAKEN", "This email is already in use.", err)

	case errors.Is(err, ErrInsufficientFunds):
		return NewAppError(http.StatusPaymentRequired, "INSUFFICIENT_FUNDS", "Your wallet balance is too low for this transaction.", err)

	case errors.Is(err, ErrTransactionLocked):
		return NewAppError(http.StatusConflict, "TRANSACTION_LOCKED", "This account is currently processing another transaction. Please try again in a few seconds.", err)

	case errors.Is(err, ErrUnauthorized):
		return NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "Please log in to continue.", err)

	default:
		return NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong on our end. Please try again later.", err)
	}
}

// how to use

// func (h *AuthHandler) Login(c *fiber.Ctx) error {
// 1. Call your usecase
// token, err := h.usecase.Login(c.Context(), req.Email, req.Password)

// 2. If it fails, parse it and return
// if err != nil {
// appErr := apperrors.ParseError(err)
// You would also log 'appErr.Err' to your internal logging system here
// return c.Status(appErr.StatusCode).JSON(appErr)
// }

// 3. Success
// return c.Status(200).JSON(token)
// }
