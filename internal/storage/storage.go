package storage

import "errors"

var (
	ErrUserExist           = errors.New("user already exist")
	ErrUserNotExist        = errors.New("user not exist")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrBalanceLessZero     = errors.New("balance is less than zero")
	ErrIdempotencyKeyExist = errors.New("idempotency key exist")

	ErrWithdrawalNotExist = errors.New("withdrawal not exist")
)
