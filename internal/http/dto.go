package http

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrWithdrawalIDEmpty = errors.New("withdrawal id is empty")
	ErrBadIDFormat       = errors.New("incorrect id")
	ErrInvalidToken      = errors.New("invalid auth token")
	ErrTokenEmpty        = errors.New("token is empty")

	ErrUserIDEmpty      = errors.New("incorrect user_id")
	ErrIncorrectBalance = errors.New("incorrect balance")

	ErrIncorrectAmount     = errors.New("incorrect amount")
	ErrCurrencyEmpty       = errors.New("currency is empty")
	ErrDestinationEmpty    = errors.New("destination is empty")
	ErrIdempotencyKeyEmpty = errors.New("idempotency_key is empty")
)

type CreateUserReqDTO struct {
	UserID  int   `json:"user_id"`
	Balance int64 `json:"balance"`
}
type CreateUserResDTO struct {
	UserID int `json:"user_id"`
}

func (d *CreateUserReqDTO) Validate() error {
	if d.UserID == 0 {
		return ErrUserIDEmpty
	}

	if d.Balance < 0 {
		return ErrIncorrectBalance
	}

	return nil
}

type WithdrawalReqDTO struct {
	UserID         int    `json:"user_id"`
	Amount         int    `json:"amount"`
	Currency       string `json:"currency"`
	Destination    string `json:"destination"`
	IdempotencyKey string `json:"idempotency_key"`
}

func (d *WithdrawalReqDTO) Validate() error {
	if d.UserID <= 0 {
		return ErrUserIDEmpty
	}

	if d.Amount <= 0 {
		return ErrIncorrectAmount
	}

	if d.Currency == "" {
		return ErrCurrencyEmpty
	}
	if d.Destination == "" {
		return ErrDestinationEmpty
	}
	if d.IdempotencyKey == "" {
		return ErrIdempotencyKeyEmpty
	}

	return nil
}

type GetWithdrawalRes struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Destination string `json:"destination"`
}

type ReqDTO struct {
	UserID  int   `json:"user_id"`
	Balance int64 `json:"balance"`
}

type ErrorDTO struct {
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

func NewErrorDTO(err error) ErrorDTO {
	root := rootError(err)

	return ErrorDTO{
		Message: root.Error(),
		Time:    time.Now(),
	}
}

func (e ErrorDTO) String() string {
	b, _ := json.MarshalIndent(e, "", "    ")

	return string(b)
}

func rootError(err error) error {
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}
