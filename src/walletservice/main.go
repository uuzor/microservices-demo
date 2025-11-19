package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Wallet represents a user's wallet
type Wallet struct {
	ID              string            `json:"id"`
	UserID          string            `json:"user_id"`
	Balance         float64           `json:"balance"`          // In USDC
	LockedBalance   float64           `json:"locked_balance"`   // In active bets
	DepositAddress  string            `json:"deposit_address"`  // Crypto deposit address
	TotalDeposited  float64           `json:"total_deposited"`
	TotalWithdrawn  float64           `json:"total_withdrawn"`
	TotalEarned     float64           `json:"total_earned"`     // From winning bets
	TotalFeesPaid   float64           `json:"total_fees_paid"`  // Platform fees
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ID              string    `json:"id"`
	WalletID        string    `json:"wallet_id"`
	UserID          string    `json:"user_id"`
	Type            string    `json:"type"` // deposit, withdrawal, bet, win, fee, refund
	Amount          float64   `json:"amount"`
	Fee             float64   `json:"fee"`
	Status          string    `json:"status"` // pending, completed, failed
	Currency        string    `json:"currency"` // USDC, USDT
	TxHash          string    `json:"tx_hash,omitempty"` // Blockchain tx hash
	MarketID        string    `json:"market_id,omitempty"`
	BetID           string    `json:"bet_id,omitempty"`
	Description     string    `json:"description"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

// WalletService manages user wallets
type WalletService struct {
	wallets      map[string]*Wallet      // userID -> Wallet
	transactions map[string]*Transaction // txID -> Transaction
	mu           sync.RWMutex
	kafkaProducer *KafkaProducer
}

// KafkaProducer handles Kafka events
type KafkaProducer struct {
	// Simplified for now - can integrate with actual Kafka later
}

func NewWalletService() *WalletService {
	return &WalletService{
		wallets:      make(map[string]*Wallet),
		transactions: make(map[string]*Transaction),
		kafkaProducer: &KafkaProducer{},
	}
}

// CreateWallet creates a new wallet for a user
func (ws *WalletService) CreateWallet(userID string) (*Wallet, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Check if wallet already exists
	if _, exists := ws.wallets[userID]; exists {
		return nil, fmt.Errorf("wallet already exists for user %s", userID)
	}

	wallet := &Wallet{
		ID:             uuid.New().String(),
		UserID:         userID,
		Balance:        0,
		LockedBalance:  0,
		DepositAddress: ws.generateDepositAddress(),
		TotalDeposited: 0,
		TotalWithdrawn: 0,
		TotalEarned:    0,
		TotalFeesPaid:  0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	ws.wallets[userID] = wallet

	// Publish wallet created event
	ws.publishEvent("wallet.created", wallet)

	return wallet, nil
}

// GetWallet retrieves a user's wallet
func (ws *WalletService) GetWallet(userID string) (*Wallet, error) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	wallet, exists := ws.wallets[userID]
	if !exists {
		return nil, fmt.Errorf("wallet not found for user %s", userID)
	}

	return wallet, nil
}

// Deposit adds funds to a user's wallet
func (ws *WalletService) Deposit(userID string, amount float64, currency string, txHash string) (*Transaction, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	wallet, exists := ws.wallets[userID]
	if !exists {
		return nil, fmt.Errorf("wallet not found")
	}

	// Create transaction
	tx := &Transaction{
		ID:          uuid.New().String(),
		WalletID:    wallet.ID,
		UserID:      userID,
		Type:        "deposit",
		Amount:      amount,
		Fee:         0, // No fee on deposits
		Status:      "completed",
		Currency:    currency,
		TxHash:      txHash,
		Description: fmt.Sprintf("Deposit of %.2f %s", amount, currency),
		CreatedAt:   time.Now(),
	}

	now := time.Now()
	tx.CompletedAt = &now

	// Update wallet balance
	wallet.Balance += amount
	wallet.TotalDeposited += amount
	wallet.UpdatedAt = time.Now()

	// Store transaction
	ws.transactions[tx.ID] = tx

	// Publish events
	ws.publishEvent("deposit.completed", tx)
	ws.publishEvent("wallet.updated", wallet)

	return tx, nil
}

// Withdraw removes funds from a user's wallet
func (ws *WalletService) Withdraw(userID string, amount float64, withdrawAddress string, currency string) (*Transaction, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	wallet, exists := ws.wallets[userID]
	if !exists {
		return nil, fmt.Errorf("wallet not found")
	}

	// Calculate withdrawal fee (1% or minimum $1)
	fee := amount * 0.01
	if fee < 1 {
		fee = 1
	}

	totalAmount := amount + fee

	// Check if user has sufficient balance
	if wallet.Balance < totalAmount {
		return nil, fmt.Errorf("insufficient balance: have %.2f, need %.2f (including %.2f fee)",
			wallet.Balance, totalAmount, fee)
	}

	// Create transaction
	tx := &Transaction{
		ID:          uuid.New().String(),
		WalletID:    wallet.ID,
		UserID:      userID,
		Type:        "withdrawal",
		Amount:      amount,
		Fee:         fee,
		Status:      "pending", // Will be completed when blockchain confirms
		Currency:    currency,
		Description: fmt.Sprintf("Withdrawal of %.2f %s to %s", amount, currency, withdrawAddress),
		Metadata: map[string]interface{}{
			"withdraw_address": withdrawAddress,
		},
		CreatedAt: time.Now(),
	}

	// Deduct from balance immediately
	wallet.Balance -= totalAmount
	wallet.TotalWithdrawn += amount
	wallet.TotalFeesPaid += fee
	wallet.UpdatedAt = time.Now()

	// Store transaction
	ws.transactions[tx.ID] = tx

	// Publish events
	ws.publishEvent("withdrawal.pending", tx)
	ws.publishEvent("wallet.updated", wallet)

	// In production, this would trigger actual blockchain withdrawal
	go ws.processWithdrawal(tx)

	return tx, nil
}

// LockFunds locks funds for an active bet
func (ws *WalletService) LockFunds(userID string, amount float64, betID string, marketID string) (*Transaction, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	wallet, exists := ws.wallets[userID]
	if !exists {
		return nil, fmt.Errorf("wallet not found")
	}

	// Calculate betting fee (2.5%)
	fee := amount * 0.025

	totalAmount := amount + fee

	// Check if user has sufficient balance
	if wallet.Balance < totalAmount {
		return nil, fmt.Errorf("insufficient balance: have %.2f, need %.2f (including %.2f fee)",
			wallet.Balance, totalAmount, fee)
	}

	// Create transaction for bet
	tx := &Transaction{
		ID:          uuid.New().String(),
		WalletID:    wallet.ID,
		UserID:      userID,
		Type:        "bet",
		Amount:      amount,
		Fee:         fee,
		Status:      "completed",
		Currency:    "USDC",
		BetID:       betID,
		MarketID:    marketID,
		Description: fmt.Sprintf("Bet of %.2f USDC on market %s", amount, marketID),
		CreatedAt:   time.Now(),
	}

	now := time.Now()
	tx.CompletedAt = &now

	// Move funds from balance to locked
	wallet.Balance -= totalAmount
	wallet.LockedBalance += amount
	wallet.TotalFeesPaid += fee
	wallet.UpdatedAt = time.Now()

	// Store transaction
	ws.transactions[tx.ID] = tx

	// Publish events
	ws.publishEvent("funds.locked", tx)
	ws.publishEvent("wallet.updated", wallet)

	return tx, nil
}

// UnlockFundsWin unlocks funds and adds winnings when bet wins
func (ws *WalletService) UnlockFundsWin(userID string, lockedAmount float64, winnings float64, betID string, marketID string) (*Transaction, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	wallet, exists := ws.wallets[userID]
	if !exists {
		return nil, fmt.Errorf("wallet not found")
	}

	// Create transaction for win
	tx := &Transaction{
		ID:          uuid.New().String(),
		WalletID:    wallet.ID,
		UserID:      userID,
		Type:        "win",
		Amount:      winnings,
		Fee:         0,
		Status:      "completed",
		Currency:    "USDC",
		BetID:       betID,
		MarketID:    marketID,
		Description: fmt.Sprintf("Winnings of %.2f USDC from market %s", winnings, marketID),
		CreatedAt:   time.Now(),
	}

	now := time.Now()
	tx.CompletedAt = &now

	// Unlock original bet and add winnings
	wallet.LockedBalance -= lockedAmount
	wallet.Balance += winnings
	wallet.TotalEarned += (winnings - lockedAmount) // Net profit
	wallet.UpdatedAt = time.Now()

	// Store transaction
	ws.transactions[tx.ID] = tx

	// Publish events
	ws.publishEvent("bet.won", tx)
	ws.publishEvent("wallet.updated", wallet)

	return tx, nil
}

// UnlockFundsLoss unlocks funds when bet loses (funds already lost)
func (ws *WalletService) UnlockFundsLoss(userID string, lockedAmount float64, betID string, marketID string) (*Transaction, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	wallet, exists := ws.wallets[userID]
	if !exists {
		return nil, fmt.Errorf("wallet not found")
	}

	// Create transaction for loss
	tx := &Transaction{
		ID:          uuid.New().String(),
		WalletID:    wallet.ID,
		UserID:      userID,
		Type:        "loss",
		Amount:      lockedAmount,
		Fee:         0,
		Status:      "completed",
		Currency:    "USDC",
		BetID:       betID,
		MarketID:    marketID,
		Description: fmt.Sprintf("Lost bet of %.2f USDC on market %s", lockedAmount, marketID),
		CreatedAt:   time.Now(),
	}

	now := time.Now()
	tx.CompletedAt = &now

	// Just unlock - funds already gone
	wallet.LockedBalance -= lockedAmount
	wallet.UpdatedAt = time.Now()

	// Store transaction
	ws.transactions[tx.ID] = tx

	// Publish events
	ws.publishEvent("bet.lost", tx)
	ws.publishEvent("wallet.updated", wallet)

	return tx, nil
}

// RefundBet refunds a bet (e.g., market cancelled)
func (ws *WalletService) RefundBet(userID string, amount float64, betID string, marketID string) (*Transaction, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	wallet, exists := ws.wallets[userID]
	if !exists {
		return nil, fmt.Errorf("wallet not found")
	}

	// Create refund transaction
	tx := &Transaction{
		ID:          uuid.New().String(),
		WalletID:    wallet.ID,
		UserID:      userID,
		Type:        "refund",
		Amount:      amount,
		Fee:         0,
		Status:      "completed",
		Currency:    "USDC",
		BetID:       betID,
		MarketID:    marketID,
		Description: fmt.Sprintf("Refund of %.2f USDC for cancelled market %s", amount, marketID),
		CreatedAt:   time.Now(),
	}

	now := time.Now()
	tx.CompletedAt = &now

	// Unlock and return to balance
	wallet.LockedBalance -= amount
	wallet.Balance += amount
	wallet.UpdatedAt = time.Now()

	// Store transaction
	ws.transactions[tx.ID] = tx

	// Publish events
	ws.publishEvent("bet.refunded", tx)
	ws.publishEvent("wallet.updated", wallet)

	return tx, nil
}

// GetTransactionHistory gets transaction history for a user
func (ws *WalletService) GetTransactionHistory(userID string, limit int) []*Transaction {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	var userTxs []*Transaction
	for _, tx := range ws.transactions {
		if tx.UserID == userID {
			userTxs = append(userTxs, tx)
		}
	}

	// Sort by created_at descending
	for i := 0; i < len(userTxs)-1; i++ {
		for j := i + 1; j < len(userTxs); j++ {
			if userTxs[j].CreatedAt.After(userTxs[i].CreatedAt) {
				userTxs[i], userTxs[j] = userTxs[j], userTxs[i]
			}
		}
	}

	// Limit results
	if limit > 0 && len(userTxs) > limit {
		userTxs = userTxs[:limit]
	}

	return userTxs
}

// GetPlatformRevenue calculates total platform revenue from fees
func (ws *WalletService) GetPlatformRevenue() float64 {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	totalRevenue := 0.0
	for _, wallet := range ws.wallets {
		totalRevenue += wallet.TotalFeesPaid
	}

	return totalRevenue
}

// Helper functions

func (ws *WalletService) generateDepositAddress() string {
	// In production, generate real crypto address
	// For now, use mock address
	return fmt.Sprintf("0x%s", uuid.New().String()[:40])
}

func (ws *WalletService) processWithdrawal(tx *Transaction) {
	// Simulate blockchain withdrawal processing
	time.Sleep(2 * time.Second)

	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Update transaction status
	tx.Status = "completed"
	tx.TxHash = fmt.Sprintf("0x%s", uuid.New().String())
	now := time.Now()
	tx.CompletedAt = &now

	ws.publishEvent("withdrawal.completed", tx)
}

func (ws *WalletService) publishEvent(eventType string, data interface{}) {
	// Publish to Kafka
	eventData, _ := json.Marshal(data)
	log.Printf("Event: %s | Data: %s", eventType, string(eventData))
	// In production: ws.kafkaProducer.Publish(eventType, eventData)
}

// HTTP Handlers

func (ws *WalletService) createWalletHandler(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wallet, err := ws.CreateWallet(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, wallet)
}

func (ws *WalletService) getWalletHandler(c *gin.Context) {
	userID := c.Param("user_id")

	wallet, err := ws.GetWallet(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

func (ws *WalletService) depositHandler(c *gin.Context) {
	var req struct {
		UserID   string  `json:"user_id" binding:"required"`
		Amount   float64 `json:"amount" binding:"required,gt=0"`
		Currency string  `json:"currency" binding:"required"`
		TxHash   string  `json:"tx_hash" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := ws.Deposit(req.UserID, req.Amount, req.Currency, req.TxHash)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

func (ws *WalletService) withdrawHandler(c *gin.Context) {
	var req struct {
		UserID          string  `json:"user_id" binding:"required"`
		Amount          float64 `json:"amount" binding:"required,gt=0"`
		WithdrawAddress string  `json:"withdraw_address" binding:"required"`
		Currency        string  `json:"currency" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := ws.Withdraw(req.UserID, req.Amount, req.WithdrawAddress, req.Currency)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

func (ws *WalletService) getTransactionsHandler(c *gin.Context) {
	userID := c.Param("user_id")
	limit := 50

	transactions := ws.GetTransactionHistory(userID, limit)

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"total":        len(transactions),
	})
}

func (ws *WalletService) getPlatformRevenueHandler(c *gin.Context) {
	revenue := ws.GetPlatformRevenue()

	c.JSON(http.StatusOK, gin.H{
		"total_revenue": revenue,
		"currency":      "USDC",
	})
}

func main() {
	walletService := NewWalletService()

	// Create some demo wallets for testing
	walletService.CreateWallet("user123")
	walletService.CreateWallet("user456")
	walletService.CreateWallet("user789")

	// Add some demo deposits
	walletService.Deposit("user123", 1000, "USDC", "0xdemo123")
	walletService.Deposit("user456", 500, "USDC", "0xdemo456")
	walletService.Deposit("user789", 2000, "USDC", "0xdemo789")

	router := gin.Default()

	// Wallet endpoints
	router.POST("/api/wallets", walletService.createWalletHandler)
	router.GET("/api/wallets/:user_id", walletService.getWalletHandler)

	// Transaction endpoints
	router.POST("/api/deposit", walletService.depositHandler)
	router.POST("/api/withdraw", walletService.withdrawHandler)
	router.GET("/api/transactions/:user_id", walletService.getTransactionsHandler)

	// Admin endpoints
	router.GET("/api/admin/revenue", walletService.getPlatformRevenueHandler)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	log.Println("Wallet Service running on :8085")
	router.Run(":8085")
}
