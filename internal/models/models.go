package models

type Withdrawal struct {
	ID     int
	UserID int

	Amount      string
	Currency    string
	Destination string

	Status string
}
