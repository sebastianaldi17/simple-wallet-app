package entity

import "github.com/shopspring/decimal"

// CreateAccountRequest represents the request to create a new account
type CreateAccountRequest struct {
	Name string `json:"account_name" binding:"required"`
}

// CreateAccountResponse represents the response after creating a new account
type CreateAccountResponse struct {
	AccountID   int64  `json:"account_id"`
	AccountName string `json:"account_name"`
}

// GetBalanceResponse represents the response for balance queries
type GetBalanceResponse struct {
	AccountID int64           `json:"account_id"`
	Balance   decimal.Decimal `json:"balance"`
}

// TransactionListResponse represents the response for transaction history queries
type TransactionListResponse struct {
	AccountID    int64               `json:"account_id"`
	StartDate    string              `json:"start_date,omitempty"`
	EndDate      string              `json:"end_date,omitempty"`
	Transactions []TransactionDetail `json:"transactions"`
}

// CreateTransactionRequest represents the request to create a transaction
type CreateTransactionRequest struct {
	Amount          decimal.Decimal `json:"amount" binding:"required"`
	Description     string          `json:"description" binding:"required"`
	TransactionType TransactionType `json:"transaction_type" binding:"required"`
}

// CreateTransferRequest represents the request to create a transfer transaction
type CreateTransferRequest struct {
	FromAccountID int64           `json:"from_account_id" binding:"required"`
	ToAccountID   int64           `json:"to_account_id" binding:"required"`
	Amount        decimal.Decimal `json:"amount" binding:"required"`
	Description   string          `json:"description" binding:"required"`
}
