package wallet

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sebastianaldi17/simple-wallet-app/internal/entity"
)

type WalletServiceInterface interface {
	CreateAccount(request entity.CreateAccountRequest) (entity.CreateAccountResponse, error)
	GetBalance(accountID int64) (entity.GetBalanceResponse, error)
	GetTransactionHistory(accountID int64, startDate, endDate string) (entity.TransactionListResponse, error)
}

type Handler struct {
	walletService WalletServiceInterface
}

func NewHandler(walletService WalletServiceInterface) *Handler {
	return &Handler{
		walletService: walletService,
	}
}

func (h *Handler) GetBalance(ctx *gin.Context) {
	accountIDStr := ctx.Param("id")
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}
	balance, err := h.walletService.GetBalance(accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		log.Printf("Error getting balance for account %d: %v", accountID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get balance"})
		return
	}
	ctx.JSON(http.StatusOK, balance)
}

func (h *Handler) GetTransactionHistory(ctx *gin.Context) {
	accountIDStr := ctx.Param("id")
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")
	if startDate != "" {
		_, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format, expected YYYY-MM-DD"})
			return
		}
	}
	if endDate != "" {
		_, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format, expected YYYY-MM-DD"})
			return
		}
	}

	transactionsResponse, err := h.walletService.GetTransactionHistory(accountID, startDate, endDate)
	if err != nil {
		if err == entity.ErrAccountNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		log.Printf("Error getting transaction history for account %d: %v", accountID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transaction history"})
		return
	}
	ctx.JSON(http.StatusOK, transactionsResponse)
}

func (h *Handler) CreateWallet(ctx *gin.Context) {
	var request entity.CreateAccountRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if request.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Account name is required"})
		return
	}
	if len(request.Name) > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Account name must be less than 100 characters"})
		return
	}
	account, err := h.walletService.CreateAccount(request)
	if err != nil {
		log.Printf("Error creating account: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}
	ctx.JSON(http.StatusCreated, account)
}
