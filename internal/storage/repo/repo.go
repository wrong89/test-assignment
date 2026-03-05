package repo

import (
	"context"
	"test-assignment/internal/models"

	"github.com/shopspring/decimal"
)

// type WithdrawalRepository interface {
// 	Create(ctx context.Context, tx Tx, w *domain.Withdrawal) error
// 	GetByID(ctx context.Context, id uuid.UUID) (*domain.Withdrawal, error)
// 	GetByIdempotencyKey(ctx context.Context, userID uuid.UUID, key string) (*domain.Withdrawal, error)
// }

type Storage interface {
	Withdrawal(
		ctx context.Context,
		userID int,
		withdrawal models.Withdrawal,
		idempotencyKey string,
	) (models.Withdrawal, error)
	GetWithdrawal(
		ctx context.Context,
		id int,
	) (models.Withdrawal, error)
	CreateUser(
		ctx context.Context,
		userID int,
		balance decimal.Decimal,
	) (int, error)
}
