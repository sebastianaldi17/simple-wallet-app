package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/sebastianaldi17/simple-wallet-app/internal/entity"
	"github.com/shopspring/decimal"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetDB() *sqlx.DB {
	return r.db
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) Begin() (*sqlx.Tx, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (r *Repository) Commit(tx *sqlx.Tx) error {
	if tx == nil {
		return nil
	}
	return tx.Commit()
}

func (r *Repository) Rollback(tx *sqlx.Tx) error {
	if tx == nil {
		return nil
	}
	return tx.Rollback()
}

func (r *Repository) GetBalance(accountID int64) (decimal.Decimal, error) {
	var balance decimal.Decimal
	query := "SELECT balance FROM denormalized_balances WHERE account_id = $1"
	err := r.db.Get(&balance, query, accountID)
	if err != nil {
		return balance, err
	}
	return balance, nil
}

func (r *Repository) GetBalanceWithLock(trx *sqlx.Tx, accountID int64) (decimal.Decimal, error) {
	var balance decimal.Decimal
	query := "SELECT balance FROM denormalized_balances WHERE account_id = $1 FOR UPDATE"
	err := trx.Get(&balance, query, accountID)
	if err != nil {
		return balance, err
	}
	return balance, nil
}

func (r *Repository) CreateAccount(trx *sqlx.Tx, accountName string) (int64, error) {
	createAccountQuery := "INSERT INTO accounts (name) VALUES ($1) RETURNING id"
	var accountID int64
	err := trx.QueryRow(createAccountQuery, accountName).Scan(&accountID)
	if err != nil {
		return 0, err
	}

	initBalanceQuery := "INSERT INTO denormalized_balances (account_id, balance) VALUES ($1, $2)"
	_, err = trx.Exec(initBalanceQuery, accountID, decimal.NewFromInt(0))
	if err != nil {
		return 0, err
	}

	return accountID, nil
}

func (r *Repository) CheckAccountExists(accountID int64) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1)"
	err := r.db.Get(&exists, query, accountID)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) CreateTransaction(trx *sqlx.Tx, accountID int64, amount decimal.Decimal, description string, isCredit bool) error {
	createTransactionQuery := "INSERT INTO transactions (description) VALUES ($1) RETURNING id"
	var transactionID int64
	err := trx.QueryRow(createTransactionQuery, description).Scan(&transactionID)
	if err != nil {
		return err
	}

	createLedgerQuery := "INSERT INTO ledgers (transaction_id, account_id, amount, is_credit) VALUES ($1, $2, $3, $4)"
	_, err = trx.Exec(createLedgerQuery, transactionID, accountID, amount, isCredit)
	if err != nil {
		return err
	}

	if isCredit {
		updateBalanceQuery := "UPDATE denormalized_balances SET balance = balance - $1 WHERE account_id = $2"
		_, err = trx.Exec(updateBalanceQuery, amount, accountID)
		if err != nil {
			return err
		}
	} else {
		updateBalanceQuery := "UPDATE denormalized_balances SET balance = balance + $1 WHERE account_id = $2"
		_, err = trx.Exec(updateBalanceQuery, amount, accountID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) CreateTransfer(trx *sqlx.Tx, fromAccountID, toAccountID int64, amount decimal.Decimal, description string) error {
	createTransactionQuery := "INSERT INTO transactions (description) VALUES ($1) RETURNING id"
	var transactionID int64
	err := trx.QueryRow(createTransactionQuery, description).Scan(&transactionID)
	if err != nil {
		return err
	}

	createLedgerQuery := "INSERT INTO ledgers (transaction_id, account_id, amount, is_credit) VALUES ($1, $2, $3, $4)"
	_, err = trx.Exec(createLedgerQuery, transactionID, fromAccountID, amount, true)
	if err != nil {
		return err
	}

	_, err = trx.Exec(createLedgerQuery, transactionID, toAccountID, amount, false)
	if err != nil {
		return err
	}

	updateFromBalanceQuery := "UPDATE denormalized_balances SET balance = balance - $1 WHERE account_id = $2"
	_, err = trx.Exec(updateFromBalanceQuery, amount, fromAccountID)
	if err != nil {
		return err
	}

	updateToBalanceQuery := "UPDATE denormalized_balances SET balance = balance + $1 WHERE account_id = $2"
	_, err = trx.Exec(updateToBalanceQuery, amount, toAccountID)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetTransactionHistory(accountID int64, startDate, endDate string) ([]entity.TransactionDetail, error) {
	var query string
	var args []interface{}

	if startDate != "" && endDate != "" {
		query = `
            SELECT t.id AS transaction_id, t.transaction_date, t.description, 
                   l.id AS ledger_id, l.account_id, l.amount, l.is_credit
            FROM transactions t
            JOIN ledgers l ON t.id = l.transaction_id
            WHERE l.account_id = $1 AND DATE(t.transaction_date) >= $2 AND DATE(t.transaction_date) <= $3
            ORDER BY t.transaction_date DESC`
		args = []interface{}{accountID, startDate, endDate}
	} else if startDate != "" {
		query = `
            SELECT t.id AS transaction_id, t.transaction_date, t.description,
                   l.id AS ledger_id, l.account_id, l.amount, l.is_credit
            FROM transactions t
            JOIN ledgers l ON t.id = l.transaction_id
            WHERE l.account_id = $1 AND DATE(t.transaction_date) >= $2
            ORDER BY t.transaction_date DESC`
		args = []interface{}{accountID, startDate}
	} else if endDate != "" {
		query = `
            SELECT t.id AS transaction_id, t.transaction_date, t.description,
                   l.id AS ledger_id, l.account_id, l.amount, l.is_credit
            FROM transactions t
            JOIN ledgers l ON t.id = l.transaction_id
            WHERE l.account_id = $1 AND DATE(t.transaction_date) <= $2
            ORDER BY t.transaction_date DESC`
		args = []interface{}{accountID, endDate}
	} else {
		query = `
            SELECT t.id AS transaction_id, t.transaction_date, t.description,
                   l.id AS ledger_id, l.account_id, l.amount, l.is_credit
            FROM transactions t
            JOIN ledgers l ON t.id = l.transaction_id
            WHERE l.account_id = $1
            ORDER BY t.transaction_date DESC`
		args = []interface{}{accountID}
	}

	transactions := make([]entity.TransactionDetail, 0)
	err := r.db.Select(&transactions, query, args...)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}
