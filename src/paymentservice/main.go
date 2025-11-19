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

// PaymentMethod represents supported payment methods
type PaymentMethod string

const (
	PaymentMethodUSDC    PaymentMethod = "USDC"
	PaymentMethodUSDT    PaymentMethod = "USDT"
	PaymentMethodETH     PaymentMethod = "ETH"
	PaymentMethodPolygon PaymentMethod = "MATIC"
)

// PaymentStatus represents payment status
type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusConfirming PaymentStatus = "confirming"
	StatusCompleted PaymentStatus = "completed"
	StatusFailed    PaymentStatus = "failed"
)

// DepositRequest represents a crypto deposit
type DepositRequest struct {
	ID              string        `json:"id"`
	UserID          string        `json:"user_id"`
	Amount          float64       `json:"amount"`
	Currency        PaymentMethod `json:"currency"`
	DepositAddress  string        `json:"deposit_address"`
	FromAddress     string        `json:"from_address,omitempty"`
	TxHash          string        `json:"tx_hash,omitempty"`
	Confirmations   int           `json:"confirmations"`
	RequiredConfs   int           `json:"required_confirmations"`
	Status          PaymentStatus `json:"status"`
	CreatedAt       time.Time     `json:"created_at"`
	CompletedAt     *time.Time    `json:"completed_at,omitempty"`
}

// WithdrawalRequest represents a crypto withdrawal
type WithdrawalRequest struct {
	ID              string        `json:"id"`
	UserID          string        `json:"user_id"`
	Amount          float64       `json:"amount"`
	Currency        PaymentMethod `json:"currency"`
	ToAddress       string        `json:"to_address"`
	TxHash          string        `json:"tx_hash,omitempty"`
	Fee             float64       `json:"fee"`
	NetworkFee      float64       `json:"network_fee"`
	Status          PaymentStatus `json:"status"`
	CreatedAt       time.Time     `json:"created_at"`
	ProcessedAt     *time.Time    `json:"processed_at,omitempty"`
	CompletedAt     *time.Time    `json:"completed_at,omitempty"`
	FailureReason   string        `json:"failure_reason,omitempty"`
}

// BlockchainConfig holds blockchain network configuration
type BlockchainConfig struct {
	Network          string
	RPCEndpoint      string
	USDCContract     string
	USDTContract     string
	ConfirmationsReq int
}

// PaymentService handles crypto payments
type PaymentService struct {
	deposits          map[string]*DepositRequest    // depositID -> Deposit
	withdrawals       map[string]*WithdrawalRequest // withdrawalID -> Withdrawal
	userDeposits      map[string][]string           // userID -> depositIDs
	userWithdrawals   map[string][]string           // userID -> withdrawalIDs
	addressToUser     map[string]string             // depositAddress -> userID
	mu                sync.RWMutex
	blockchainConfig  BlockchainConfig
	walletServiceURL  string
}

func NewPaymentService() *PaymentService {
	return &PaymentService{
		deposits:         make(map[string]*DepositRequest),
		withdrawals:      make(map[string]*WithdrawalRequest),
		userDeposits:     make(map[string][]string),
		userWithdrawals:  make(map[string][]string),
		addressToUser:    make(map[string]string),
		walletServiceURL: "http://localhost:8085",
		blockchainConfig: BlockchainConfig{
			Network:          "polygon-mainnet", // Using Polygon for low fees
			RPCEndpoint:      "https://polygon-rpc.com",
			USDCContract:     "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174", // USDC on Polygon
			USDTContract:     "0xc2132D05D31c914a87C6611C10748AEb04B58e8F", // USDT on Polygon
			ConfirmationsReq: 10,
		},
	}
}

// GenerateDepositAddress generates a unique deposit address for a user
func (ps *PaymentService) GenerateDepositAddress(userID string, currency PaymentMethod) (string, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// In production, generate real blockchain address
	// For now, generate mock address
	address := fmt.Sprintf("0x%s%s", userID[:8], uuid.New().String()[:32])

	// Map address to user
	ps.addressToUser[address] = userID

	log.Printf("Generated deposit address %s for user %s (currency: %s)", address, userID, currency)

	return address, nil
}

// CreateDepositRequest creates a new deposit request
func (ps *PaymentService) CreateDepositRequest(userID string, currency PaymentMethod) (*DepositRequest, error) {
	depositAddress, err := ps.GenerateDepositAddress(userID, currency)
	if err != nil {
		return nil, err
	}

	deposit := &DepositRequest{
		ID:              uuid.New().String(),
		UserID:          userID,
		Amount:          0, // Will be updated when funds arrive
		Currency:        currency,
		DepositAddress:  depositAddress,
		Confirmations:   0,
		RequiredConfs:   ps.blockchainConfig.ConfirmationsReq,
		Status:          StatusPending,
		CreatedAt:       time.Now(),
	}

	ps.mu.Lock()
	ps.deposits[deposit.ID] = deposit
	ps.userDeposits[userID] = append(ps.userDeposits[userID], deposit.ID)
	ps.mu.Unlock()

	// Start monitoring for deposits to this address
	go ps.monitorDepositAddress(deposit)

	return deposit, nil
}

// monitorDepositAddress monitors blockchain for deposits
func (ps *PaymentService) monitorDepositAddress(deposit *DepositRequest) {
	// In production, this would:
	// 1. Subscribe to blockchain events
	// 2. Watch for transactions to deposit address
	// 3. Verify transaction confirmations
	// 4. Update deposit status
	// 5. Notify wallet service

	// For demo, simulate deposit after 5 seconds
	time.Sleep(5 * time.Second)

	// Simulate receiving deposit (in production, read from blockchain)
	ps.processIncomingDeposit(deposit.DepositAddress, 100.0, "USDC", "0xsimulated_tx_hash_123")
}

// processIncomingDeposit processes an incoming blockchain deposit
func (ps *PaymentService) processIncomingDeposit(toAddress string, amount float64, currency string, txHash string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Find user by deposit address
	userID, exists := ps.addressToUser[toAddress]
	if !exists {
		log.Printf("Received deposit to unknown address: %s", toAddress)
		return
	}

	// Find pending deposit for this user
	var deposit *DepositRequest
	for _, depID := range ps.userDeposits[userID] {
		dep := ps.deposits[depID]
		if dep.DepositAddress == toAddress && dep.Status == StatusPending {
			deposit = dep
			break
		}
	}

	if deposit == nil {
		log.Printf("No pending deposit found for address: %s", toAddress)
		return
	}

	// Update deposit
	deposit.Amount = amount
	deposit.TxHash = txHash
	deposit.Status = StatusConfirming
	deposit.Confirmations = 1

	log.Printf("Deposit detected: %s sent %.2f %s (tx: %s)", userID, amount, currency, txHash)

	// Start confirmation monitoring
	go ps.monitorConfirmations(deposit)
}

// monitorConfirmations monitors transaction confirmations
func (ps *PaymentService) monitorConfirmations(deposit *DepositRequest) {
	for deposit.Confirmations < deposit.RequiredConfs {
		time.Sleep(2 * time.Second) // In production, poll blockchain

		ps.mu.Lock()
		deposit.Confirmations++
		log.Printf("Deposit %s: %d/%d confirmations", deposit.ID, deposit.Confirmations, deposit.RequiredConfs)
		ps.mu.Unlock()
	}

	// Deposit fully confirmed
	ps.completeDeposit(deposit)
}

// completeDeposit marks deposit as completed and credits user wallet
func (ps *PaymentService) completeDeposit(deposit *DepositRequest) {
	ps.mu.Lock()
	deposit.Status = StatusCompleted
	now := time.Now()
	deposit.CompletedAt = &now
	ps.mu.Unlock()

	log.Printf("Deposit completed: User %s received %.2f %s", deposit.UserID, deposit.Amount, deposit.Currency)

	// Credit user wallet via wallet service
	ps.creditWallet(deposit.UserID, deposit.Amount, string(deposit.Currency), deposit.TxHash)
}

// creditWallet calls wallet service to credit user balance
func (ps *PaymentService) creditWallet(userID string, amount float64, currency string, txHash string) {
	// In production, make HTTP request to wallet service
	log.Printf("Crediting wallet: User %s + %.2f %s", userID, amount, currency)

	// Mock HTTP request
	requestData := map[string]interface{}{
		"user_id":  userID,
		"amount":   amount,
		"currency": currency,
		"tx_hash":  txHash,
	}

	requestJSON, _ := json.Marshal(requestData)
	log.Printf("POST %s/api/deposit: %s", ps.walletServiceURL, string(requestJSON))

	// In production:
	// resp, err := http.Post(ps.walletServiceURL+"/api/deposit", "application/json", bytes.NewBuffer(requestJSON))
}

// CreateWithdrawalRequest creates a new withdrawal request
func (ps *PaymentService) CreateWithdrawalRequest(userID string, amount float64, toAddress string, currency PaymentMethod) (*WithdrawalRequest, error) {
	// Validate address format (basic check)
	if len(toAddress) != 42 || toAddress[:2] != "0x" {
		return nil, fmt.Errorf("invalid withdrawal address format")
	}

	// Calculate fees
	platformFee := amount * 0.01 // 1% platform fee
	if platformFee < 1 {
		platformFee = 1
	}

	networkFee := ps.estimateNetworkFee(currency)

	withdrawal := &WithdrawalRequest{
		ID:         uuid.New().String(),
		UserID:     userID,
		Amount:     amount,
		Currency:   currency,
		ToAddress:  toAddress,
		Fee:        platformFee,
		NetworkFee: networkFee,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
	}

	ps.mu.Lock()
	ps.withdrawals[withdrawal.ID] = withdrawal
	ps.userWithdrawals[userID] = append(ps.userWithdrawals[userID], withdrawal.ID)
	ps.mu.Unlock()

	// Process withdrawal asynchronously
	go ps.processWithdrawal(withdrawal)

	return withdrawal, nil
}

// processWithdrawal processes a withdrawal request
func (ps *PaymentService) processWithdrawal(withdrawal *WithdrawalRequest) {
	// Step 1: Verify wallet service has locked funds
	time.Sleep(1 * time.Second)

	ps.mu.Lock()
	now := time.Now()
	withdrawal.ProcessedAt = &now
	ps.mu.Unlock()

	// Step 2: Send blockchain transaction
	txHash, err := ps.sendBlockchainTransaction(withdrawal)
	if err != nil {
		ps.failWithdrawal(withdrawal, err.Error())
		return
	}

	ps.mu.Lock()
	withdrawal.TxHash = txHash
	withdrawal.Status = StatusConfirming
	ps.mu.Unlock()

	log.Printf("Withdrawal initiated: %s sending %.2f %s to %s (tx: %s)",
		withdrawal.UserID, withdrawal.Amount, withdrawal.Currency, withdrawal.ToAddress, txHash)

	// Step 3: Monitor transaction confirmation
	time.Sleep(5 * time.Second) // Simulate blockchain confirmation

	ps.completeWithdrawal(withdrawal)
}

// sendBlockchainTransaction sends funds on blockchain
func (ps *PaymentService) sendBlockchainTransaction(withdrawal *WithdrawalRequest) (string, error) {
	// In production:
	// 1. Connect to blockchain RPC
	// 2. Sign transaction with hot wallet private key
	// 3. Broadcast transaction
	// 4. Return transaction hash

	// For demo, generate mock tx hash
	txHash := fmt.Sprintf("0x%s", uuid.New().String())

	log.Printf("Broadcasting transaction: %s %.2f %s to %s",
		withdrawal.Currency, withdrawal.Amount, withdrawal.Currency, withdrawal.ToAddress)

	return txHash, nil
}

// completeWithdrawal marks withdrawal as completed
func (ps *PaymentService) completeWithdrawal(withdrawal *WithdrawalRequest) {
	ps.mu.Lock()
	withdrawal.Status = StatusCompleted
	now := time.Now()
	withdrawal.CompletedAt = &now
	ps.mu.Unlock()

	log.Printf("Withdrawal completed: User %s withdrew %.2f %s (tx: %s)",
		withdrawal.UserID, withdrawal.Amount, withdrawal.Currency, withdrawal.TxHash)
}

// failWithdrawal marks withdrawal as failed
func (ps *PaymentService) failWithdrawal(withdrawal *WithdrawalRequest, reason string) {
	ps.mu.Lock()
	withdrawal.Status = StatusFailed
	withdrawal.FailureReason = reason
	ps.mu.Unlock()

	log.Printf("Withdrawal failed: User %s - %s", withdrawal.UserID, reason)

	// Refund to wallet service
	// In production: call wallet service to unlock funds
}

// estimateNetworkFee estimates blockchain network fee
func (ps *PaymentService) estimateNetworkFee(currency PaymentMethod) float64 {
	// In production, query blockchain for current gas prices
	// Polygon has very low fees (~$0.01-0.10)
	switch currency {
	case PaymentMethodUSDC, PaymentMethodUSDT:
		return 0.05 // $0.05 for Polygon
	case PaymentMethodETH:
		return 5.0 // $5 for Ethereum mainnet (high!)
	default:
		return 0.10
	}
}

// GetDeposit retrieves a deposit by ID
func (ps *PaymentService) GetDeposit(depositID string) (*DepositRequest, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	deposit, exists := ps.deposits[depositID]
	if !exists {
		return nil, fmt.Errorf("deposit not found")
	}

	return deposit, nil
}

// GetWithdrawal retrieves a withdrawal by ID
func (ps *PaymentService) GetWithdrawal(withdrawalID string) (*WithdrawalRequest, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	withdrawal, exists := ps.withdrawals[withdrawalID]
	if !exists {
		return nil, fmt.Errorf("withdrawal not found")
	}

	return withdrawal, nil
}

// GetUserDeposits gets all deposits for a user
func (ps *PaymentService) GetUserDeposits(userID string) []*DepositRequest {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var deposits []*DepositRequest
	for _, depID := range ps.userDeposits[userID] {
		if dep, exists := ps.deposits[depID]; exists {
			deposits = append(deposits, dep)
		}
	}

	return deposits
}

// GetUserWithdrawals gets all withdrawals for a user
func (ps *PaymentService) GetUserWithdrawals(userID string) []*WithdrawalRequest {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var withdrawals []*WithdrawalRequest
	for _, wdID := range ps.userWithdrawals[userID] {
		if wd, exists := ps.withdrawals[wdID]; exists {
			withdrawals = append(withdrawals, wd)
		}
	}

	return withdrawals
}

// HTTP Handlers

func (ps *PaymentService) createDepositHandler(c *gin.Context) {
	var req struct {
		UserID   string        `json:"user_id" binding:"required"`
		Currency PaymentMethod `json:"currency" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deposit, err := ps.CreateDepositRequest(req.UserID, req.Currency)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"deposit_id":      deposit.ID,
		"deposit_address": deposit.DepositAddress,
		"currency":        deposit.Currency,
		"status":          deposit.Status,
		"instructions":    fmt.Sprintf("Send %s to address: %s", deposit.Currency, deposit.DepositAddress),
		"required_confirmations": deposit.RequiredConfs,
	})
}

func (ps *PaymentService) createWithdrawalHandler(c *gin.Context) {
	var req struct {
		UserID    string        `json:"user_id" binding:"required"`
		Amount    float64       `json:"amount" binding:"required,gt=0"`
		ToAddress string        `json:"to_address" binding:"required"`
		Currency  PaymentMethod `json:"currency" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	withdrawal, err := ps.CreateWithdrawalRequest(req.UserID, req.Amount, req.ToAddress, req.Currency)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"withdrawal_id": withdrawal.ID,
		"amount":        withdrawal.Amount,
		"to_address":    withdrawal.ToAddress,
		"platform_fee":  withdrawal.Fee,
		"network_fee":   withdrawal.NetworkFee,
		"total_cost":    withdrawal.Amount + withdrawal.Fee + withdrawal.NetworkFee,
		"status":        withdrawal.Status,
		"currency":      withdrawal.Currency,
	})
}

func (ps *PaymentService) getDepositHandler(c *gin.Context) {
	depositID := c.Param("id")

	deposit, err := ps.GetDeposit(depositID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deposit)
}

func (ps *PaymentService) getWithdrawalHandler(c *gin.Context) {
	withdrawalID := c.Param("id")

	withdrawal, err := ps.GetWithdrawal(withdrawalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, withdrawal)
}

func (ps *PaymentService) getUserDepositsHandler(c *gin.Context) {
	userID := c.Param("user_id")
	deposits := ps.GetUserDeposits(userID)

	c.JSON(http.StatusOK, gin.H{
		"deposits": deposits,
		"total":    len(deposits),
	})
}

func (ps *PaymentService) getUserWithdrawalsHandler(c *gin.Context) {
	userID := c.Param("user_id")
	withdrawals := ps.GetUserWithdrawals(userID)

	c.JSON(http.StatusOK, gin.H{
		"withdrawals": withdrawals,
		"total":       len(withdrawals),
	})
}

func (ps *PaymentService) getSupportedCurrenciesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"currencies": []gin.H{
			{
				"code":        "USDC",
				"name":        "USD Coin",
				"network":     "Polygon",
				"decimals":    6,
				"min_deposit": 10.0,
				"min_withdraw": 20.0,
			},
			{
				"code":        "USDT",
				"name":        "Tether",
				"network":     "Polygon",
				"decimals":    6,
				"min_deposit": 10.0,
				"min_withdraw": 20.0,
			},
		},
	})
}

func main() {
	paymentService := NewPaymentService()

	router := gin.Default()

	// Deposit endpoints
	router.POST("/api/deposits", paymentService.createDepositHandler)
	router.GET("/api/deposits/:id", paymentService.getDepositHandler)
	router.GET("/api/users/:user_id/deposits", paymentService.getUserDepositsHandler)

	// Withdrawal endpoints
	router.POST("/api/withdrawals", paymentService.createWithdrawalHandler)
	router.GET("/api/withdrawals/:id", paymentService.getWithdrawalHandler)
	router.GET("/api/users/:user_id/withdrawals", paymentService.getUserWithdrawalsHandler)

	// Info endpoints
	router.GET("/api/currencies", paymentService.getSupportedCurrenciesHandler)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	log.Println("Payment Service running on :8086")
	log.Println("Supported networks: Polygon (USDC, USDT)")
	log.Println("Blockchain RPC:", paymentService.blockchainConfig.RPCEndpoint)
	router.Run(":8086")
}
