package http

import (
	"test-assignment/internal/models"
)

func sameWithdrawal(a models.Withdrawal, b models.Withdrawal) bool {
	return a.Amount == b.Amount &&
		a.Currency == b.Currency &&
		a.Destination == b.Destination &&
		a.UserID == b.UserID
}
