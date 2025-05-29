package transaction

import (
	"github.com/jmoiron/sqlx"
	"github.com/sebastianaldi17/simple-wallet-app/internal/entity"
	"github.com/shopspring/decimal"
)

type RepositoryInterface interface {
	Begin() (*sqlx.Tx, error)
	Commit(tx *sqlx.Tx) error
	Rollback(tx *sqlx.Tx) error
	GetBalanceWithLock(trx *sqlx.Tx, accountID int64) (decimal.Decimal, error)
	CreateTransaction(trx *sqlx.Tx, accountID int64, amount decimal.Decimal, description string, isCredit bool) error
	CreateTransfer(trx *sqlx.Tx, fromAccountID, toAccountID int64, amount decimal.Decimal, description string) error
	CheckAccountExists(accountID int64) (bool, error)
}

type Service struct {
	repository RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{
		repository: repo,
	}
}

func (s *Service) HandleDeposit(accountID int64, amount decimal.Decimal, description string) error {
	tx, err := s.repository.Begin()
	if err != nil {
		return err
	}
	defer s.repository.Rollback(tx)

	_, err = s.repository.GetBalanceWithLock(tx, accountID)
	if err != nil {
		return err
	}

	err = s.repository.CreateTransaction(tx, accountID, amount, description, false)
	if err != nil {
		return err
	}

	return s.repository.Commit(tx)
}

func (s *Service) HandleWithdraw(accountID int64, amount decimal.Decimal, description string) error {
	tx, err := s.repository.Begin()
	if err != nil {
		return err
	}
	defer s.repository.Rollback(tx)

	balance, err := s.repository.GetBalanceWithLock(tx, accountID)
	if err != nil {
		return err
	}

	if balance.LessThan(amount) {
		return entity.ErrInsufficientFunds
	}

	err = s.repository.CreateTransaction(tx, accountID, amount, description, true)
	if err != nil {
		return err
	}

	return s.repository.Commit(tx)
}

func (s *Service) HandleTransfer(fromAccountID, toAccountID int64, amount decimal.Decimal, description string) error {
	tx, err := s.repository.Begin()
	if err != nil {
		return err
	}
	defer s.repository.Rollback(tx)

	// Verify that both accounts exist
	fromExists, err := s.repository.CheckAccountExists(fromAccountID)
	if err != nil {
		return err
	}
	if !fromExists {
		return entity.ErrAccountNotFound
	}
	toExists, err := s.repository.CheckAccountExists(toAccountID)
	if err != nil {
		return err
	}
	if !toExists {
		return entity.ErrAccountNotFound
	}

	// Lock accounts in consistent order (ascending by ID) to prevent deadlocks
	firstLockID := min(fromAccountID, toAccountID)
	secondLockID := max(fromAccountID, toAccountID)

	firstBalance, err := s.repository.GetBalanceWithLock(tx, firstLockID)
	if err != nil {
		return err
	}

	secondBalance, err := s.repository.GetBalanceWithLock(tx, secondLockID)
	if err != nil {
		return err
	}

	var fromBalance decimal.Decimal
	if fromAccountID == firstLockID {
		fromBalance = firstBalance
	} else {
		fromBalance = secondBalance
	}

	if fromBalance.LessThan(amount) {
		return entity.ErrInsufficientFunds
	}

	err = s.repository.CreateTransfer(tx, fromAccountID, toAccountID, amount, description)
	if err != nil {
		return err
	}

	return s.repository.Commit(tx)
}
