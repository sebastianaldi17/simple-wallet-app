package transaction

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sebastianaldi17/simple-wallet-app/internal/entity"
	"github.com/shopspring/decimal"
)

type TransactionServiceInterface interface {
	HandleDeposit(accountID int64, amount decimal.Decimal, description string) error
	HandleWithdraw(accountID int64, amount decimal.Decimal, description string) error
	HandleTransfer(fromAccountID, toAccountID int64, amount decimal.Decimal, description string) error
}

type Handler struct {
	transactionService TransactionServiceInterface
}

func NewHandler(transactionService TransactionServiceInterface) *Handler {
	return &Handler{
		transactionService: transactionService,
	}
}

// HandleNewTransaction processes a new transaction request (deposit or withdrawal)
// Transfers are handled separately in HandleTransfer
func (h *Handler) HandleNewTransaction(ctx *gin.Context) {
	var request entity.CreateTransactionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if request.Amount.LessThanOrEqual(decimal.Zero) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be greater than zero"})
		return
	}

	if request.Description == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
		return
	}

	if len(request.Description) > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Description must be less than 100 characters"})
		return
	}

	accountIDStr := ctx.Param("id")
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	switch request.TransactionType {
	case entity.TransactionTypeDeposit:
		err = h.transactionService.HandleDeposit(accountID, request.Amount, request.Description)
	case entity.TransactionTypeWithdrawal:
		err = h.transactionService.HandleWithdraw(accountID, request.Amount, request.Description)
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction type"})
		return
	}
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		if err == entity.ErrInsufficientFunds {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds for withdrawal"})
			return
		}
		log.Printf("Error processing transaction for account %d: %v", accountID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process transaction"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "New transaction successful"})
}

func (h *Handler) HandleTransfer(ctx *gin.Context) {
	var request entity.CreateTransferRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if request.Amount.LessThanOrEqual(decimal.Zero) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be greater than zero"})
		return
	}

	if request.FromAccountID == request.ToAccountID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot transfer to the same account"})
		return
	}

	if request.Description == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Description is required"})
		return
	}

	if len(request.Description) > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Description must be less than 100 characters"})
		return
	}

	err := h.transactionService.HandleTransfer(request.FromAccountID, request.ToAccountID, request.Amount, request.Description)
	if err != nil {
		if err == entity.ErrAccountNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "One or both accounts not found"})
			return
		}
		if err == entity.ErrInsufficientFunds {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds for transfer"})
			return
		}
		log.Printf("Error processing transfer from account %d to %d: %v", request.FromAccountID, request.ToAccountID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process transfer"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
}
