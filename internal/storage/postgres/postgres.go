package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"test-assignment/internal/models"
	"test-assignment/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type Storage struct {
	db *pgx.Conn
}

func New(ctx context.Context, connURL string) (*Storage, error) {
	const op = "storage.postgres.New"

	conn, err := pgx.Connect(ctx, connURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		db: conn,
	}, nil
}

func (s *Storage) Close(ctx context.Context) error {
	return s.db.Close(ctx)
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *Storage) GetWithdrawal(
	ctx context.Context,
	id int,
) (models.Withdrawal, error) {
	const op = "storage.postgres.GetWithdrawal"

	sqlQuery := `SELECT id, user_id, amount, currency, destination, status FROM withdrawals WHERE id = @id`
	args := pgx.NamedArgs{
		"id": id,
	}

	var wd models.Withdrawal

	err := s.db.QueryRow(ctx, sqlQuery, args).Scan(
		&wd.ID,
		&wd.UserID,
		&wd.Amount,
		&wd.Currency,
		&wd.Destination,
		&wd.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Withdrawal{}, fmt.Errorf("%s: %w", op, storage.ErrWithdrawalNotExist)
		}

		return models.Withdrawal{}, fmt.Errorf("%s: %w", op, err)
	}

	return wd, nil
}

func (s *Storage) Withdrawal(
	ctx context.Context,
	userID int,
	withdrawal models.Withdrawal,
	idempotencyKey string,
) (models.Withdrawal, error) {
	const op = "storage.postgres.Withdrawal"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.Withdrawal{}, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	var balanceStr string

	sqlQuery := `SELECT amount FROM balances WHERE user_id = @user_id FOR UPDATE`
	args := pgx.NamedArgs{
		"user_id": userID,
	}

	if err := tx.QueryRow(ctx, sqlQuery, args).Scan(&balanceStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Withdrawal{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotExist)
		}
		return models.Withdrawal{}, fmt.Errorf("%s: %w", op, err)
	}

	balance, err := decimal.NewFromString(balanceStr)
	if err != nil {
		return models.Withdrawal{}, fmt.Errorf("%s: %w", op, err)
	}

	amount, err := decimal.NewFromString(withdrawal.Amount)
	if err != nil {
		return models.Withdrawal{}, fmt.Errorf("%s: %w", op, err)
	}

	if amount.LessThan(decimal.NewFromInt(0)) {
		return models.Withdrawal{}, fmt.Errorf("%s: %w", op, storage.ErrBalanceLessZero)
	}

	if balance.LessThan(amount) {
		return models.Withdrawal{}, fmt.Errorf("%s: %w", op, storage.ErrInsufficientBalance)
	}

	var wd models.Withdrawal

	sqlQuery = `SELECT id, user_id, amount, currency, destination, status FROM withdrawals WHERE idempotency_key = @key`
	args = pgx.NamedArgs{
		"key": idempotencyKey,
	}

	err = tx.QueryRow(ctx, sqlQuery, args).Scan(
		&wd.ID,
		&wd.UserID,
		&wd.Amount,
		&wd.Currency,
		&wd.Destination,
		&wd.Status,
	)
	if err == nil {
		tx.Commit(ctx)
		return wd, nil
	}

	sqlQuery = `
		INSERT INTO withdrawals(user_id, amount, currency, destination, idempotency_key)
		VALUES(@user_id, @amount, @currency, @destination, @key)
		RETURNING id
	`
	args = pgx.NamedArgs{
		"user_id":     userID,
		"amount":      amount,
		"currency":    withdrawal.Currency,
		"destination": withdrawal.Destination,
		"key":         idempotencyKey,
	}

	var newID int

	err = tx.QueryRow(ctx, sqlQuery, args).Scan(&newID)
	if err != nil {
		return models.Withdrawal{}, fmt.Errorf("%s: %w", op, err)
	}

	// _, err = tx.Exec(ctx, sqlQuery, args)
	// if err != nil {
	// 	return models.Withdrawal{}, fmt.Errorf("%s: %w", op, err)
	// }

	withdrawal.ID = newID

	tx.Commit(ctx)
	return withdrawal, nil
}

func (s *Storage) CreateUser(
	ctx context.Context,
	userID int,
	balance decimal.Decimal,
) (int, error) {
	const op = "storage.postgres.CreateUser"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	var usrID int

	sqlQuery := `SELECT user_id FROM balances WHERE user_id = @user_id`
	args := pgx.NamedArgs{
		"user_id": userID,
	}

	if err := tx.QueryRow(ctx, sqlQuery, args).Scan(&usrID); err == nil {
		return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExist)
	}

	sqlQuery = `INSERT INTO balances(user_id, amount) VALUES(@user_id, @amount)`
	args = pgx.NamedArgs{
		"user_id": userID,
		"amount":  balance.String(),
	}

	_, err = tx.Exec(ctx, sqlQuery, args)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userID, tx.Commit(ctx)
}

// Just plus the user balance
func (s *Storage) Deposit(
	ctx context.Context,
	userID int,
	amount decimal.Decimal,
) error {
	const op = "storage.postgres.Deposit"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	var newBalance string

	sql := `
		UPDATE balances
		SET amount = amount + @deposit::numeric
		WHERE user_id = @user_id
		RETURNING amount
	`
	args := pgx.NamedArgs{
		"deposit": amount.String(),
		"user_id": userID,
	}

	if err = tx.QueryRow(ctx, sql, args).Scan(&newBalance); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return tx.Commit(ctx)
}

func (s *Storage) GetBalance(
	ctx context.Context,
	userID int,
) (decimal.Decimal, error) {
	const op = "storage.postgres.GetBalance"

	sqlQuery := `SELECT amount FROM balances WHERE user_id = @user_id`
	args := pgx.NamedArgs{
		"user_id": userID,
	}

	var m decimal.Decimal

	if err := s.db.QueryRow(ctx, sqlQuery, args).Scan(&m); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return decimal.Decimal{}, storage.ErrUserNotExist
		}

		fmt.Println("ERR", err)
		return decimal.Decimal{}, err
	}

	return m, nil
}
