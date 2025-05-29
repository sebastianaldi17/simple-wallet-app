package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type TransactionDetail struct {
	// Transaction fields
	TransactionID   int       `json:"transaction_id" db:"transaction_id"`
	TransactionDate time.Time `json:"transaction_date" db:"transaction_date"`
	Description     string    `json:"description" db:"description"`

	// Ledger fields
	LedgerID  int             `json:"ledger_id" db:"ledger_id"`
	AccountID int             `json:"account_id" db:"account_id"`
	Amount    decimal.Decimal `json:"amount" db:"amount"`
	IsCredit  bool            `json:"is_credit" db:"is_credit"`
}
