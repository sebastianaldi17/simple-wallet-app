package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	transactionHandler "github.com/sebastianaldi17/simple-wallet-app/internal/handler/transaction"
	walletHandler "github.com/sebastianaldi17/simple-wallet-app/internal/handler/wallet"
	"github.com/sebastianaldi17/simple-wallet-app/internal/repository"
	transactionService "github.com/sebastianaldi17/simple-wallet-app/internal/service/transaction"
	walletService "github.com/sebastianaldi17/simple-wallet-app/internal/service/wallet"
)

func main() {
	connectionString := os.Getenv("DATABASE_URL")
	if connectionString == "" {
		connectionString = "postgres://postgres:postgres@127.0.0.1:5432/wallet-app?sslmode=disable"
	}
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Initialize repository
	repository := repository.NewRepository(db)

	// Initialize services
	transactionService := transactionService.NewService(repository)
	walletService := walletService.NewService(repository)

	// Initialize handlers
	transactionHandler := transactionHandler.NewHandler(transactionService)
	walletHandler := walletHandler.NewHandler(walletService)

	// Register routes
	r := gin.Default()

	r.POST("/wallets", walletHandler.CreateWallet)
	r.GET("/wallets/:id", walletHandler.GetBalance)
	r.GET("/wallets/:id/transactions", walletHandler.GetTransactionHistory)
	r.POST("/wallets/:id/transactions", transactionHandler.HandleNewTransaction)

	r.POST("/transfers", transactionHandler.HandleTransfer)

	r.Run(":8080")
}
