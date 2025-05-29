package wallet

import (
	"github.com/jmoiron/sqlx"
	"github.com/sebastianaldi17/simple-wallet-app/internal/entity"
	"github.com/shopspring/decimal"
)

type RepositoryInterface interface {
	Begin() (*sqlx.Tx, error)
	Commit(tx *sqlx.Tx) error
	Rollback(tx *sqlx.Tx) error
	GetBalance(accountID int64) (decimal.Decimal, error)
	CreateAccount(trx *sqlx.Tx, accountName string) (int64, error)
	CheckAccountExists(accountID int64) (bool, error)
	GetTransactionHistory(accountID int64, startDate, endDate string) ([]entity.TransactionDetail, error)
}

type Service struct {
	repository RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{
		repository: repo,
	}
}

func (s *Service) GetBalance(accountID int64) (entity.GetBalanceResponse, error) {
	balance, err := s.repository.GetBalance(accountID)
	if err != nil {
		return entity.GetBalanceResponse{}, err
	}
	return entity.GetBalanceResponse{
		AccountID: accountID,
		Balance:   balance,
	}, nil
}

func (s *Service) CreateAccount(request entity.CreateAccountRequest) (entity.CreateAccountResponse, error) {
	tx, err := s.repository.Begin()
	if err != nil {
		return entity.CreateAccountResponse{}, err
	}
	defer s.repository.Rollback(tx)

	accountID, err := s.repository.CreateAccount(tx, request.Name)
	if err != nil {
		return entity.CreateAccountResponse{}, err
	}

	err = s.repository.Commit(tx)
	if err != nil {
		return entity.CreateAccountResponse{}, err
	}
	return entity.CreateAccountResponse{
		AccountID:   accountID,
		AccountName: request.Name,
	}, nil
}

func (s *Service) GetTransactionHistory(accountID int64, startDate, endDate string) (entity.TransactionListResponse, error) {
	exists, err := s.repository.CheckAccountExists(accountID) // Ensure the account exists before fetching transaction history
	if err != nil {
		return entity.TransactionListResponse{}, err
	}
	if !exists {
		return entity.TransactionListResponse{}, entity.ErrAccountNotFound
	}
	transactions, err := s.repository.GetTransactionHistory(accountID, startDate, endDate)
	if err != nil {
		return entity.TransactionListResponse{}, err
	}
	return entity.TransactionListResponse{
		AccountID:    accountID,
		Transactions: transactions,
		StartDate:    startDate,
		EndDate:      endDate,
	}, nil
}
