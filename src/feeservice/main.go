package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// FeeType represents different types of fees
type FeeType string

const (
	FeeTradingStandard FeeType = "trading_standard"
	FeeTradingVIP      FeeType = "trading_vip"
	FeeWithdrawal      FeeType = "withdrawal"
	FeeMarketCreation  FeeType = "market_creation"
)

// FeeTier represents user fee tier based on volume
type FeeTier string

const (
	TierStandard FeeTier = "standard" // 0-10K monthly volume
	TierBronze   FeeTier = "bronze"   // 10K-50K monthly volume
	TierSilver   FeeTier = "silver"   // 50K-100K monthly volume
	TierGold     FeeTier = "gold"     // 100K+ monthly volume
)

// FeeConfig holds fee configuration
type FeeConfig struct {
	TradingFees map[FeeTier]float64 // Percentage fees by tier
	WithdrawalFeePercent float64
	WithdrawalFeeMin     float64
	NetworkFees          map[string]float64 // Network fees by currency
}

// UserVolume tracks user trading volume for fee tiers
type UserVolume struct {
	UserID         string
	MonthlyVolume  float64
	Tier           FeeTier
	LastUpdated    time.Time
}

// FeeCalculation represents a calculated fee
type FeeCalculation struct {
	Amount          float64 `json:"amount"`
	FeeType         string  `json:"fee_type"`
	FeePercent      float64 `json:"fee_percent"`
	FeeAmount       float64 `json:"fee_amount"`
	NetworkFee      float64 `json:"network_fee,omitempty"`
	TotalFee        float64 `json:"total_fee"`
	UserTier        string  `json:"user_tier"`
	Breakdown       map[string]float64 `json:"breakdown"`
}

// RevenueStats tracks platform revenue
type RevenueStats struct {
	TotalRevenue        float64            `json:"total_revenue"`
	RevenueByType       map[string]float64 `json:"revenue_by_type"`
	RevenueByTier       map[string]float64 `json:"revenue_by_tier"`
	TotalVolume         float64            `json:"total_volume"`
	TransactionCount    int                `json:"transaction_count"`
	AverageRevenuePerTx float64            `json:"average_revenue_per_tx"`
	LastUpdated         time.Time          `json:"last_updated"`
}

// FeeService manages all fee calculations and revenue tracking
type FeeService struct {
	config       FeeConfig
	userVolumes  map[string]*UserVolume
	revenue      *RevenueStats
	mu           sync.RWMutex
}

func NewFeeService() *FeeService {
	return &FeeService{
		config: FeeConfig{
			TradingFees: map[FeeTier]float64{
				TierStandard: 0.025, // 2.5%
				TierBronze:   0.020, // 2.0%
				TierSilver:   0.015, // 1.5%
				TierGold:     0.010, // 1.0%
			},
			WithdrawalFeePercent: 0.01,  // 1%
			WithdrawalFeeMin:     1.0,   // $1 minimum
			NetworkFees: map[string]float64{
				"USDC": 0.05,  // Polygon network fee
				"USDT": 0.05,
				"ETH":  5.0,   // Ethereum mainnet (high!)
			},
		},
		userVolumes: make(map[string]*UserVolume),
		revenue: &RevenueStats{
			TotalRevenue:     0,
			RevenueByType:    make(map[string]float64),
			RevenueByTier:    make(map[string]float64),
			TotalVolume:      0,
			TransactionCount: 0,
			LastUpdated:      time.Now(),
		},
	}
}

// GetUserTier determines user's fee tier based on monthly volume
func (fs *FeeService) GetUserTier(userID string) FeeTier {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	volume, exists := fs.userVolumes[userID]
	if !exists {
		return TierStandard
	}

	// Check if volume data is current month
	now := time.Now()
	if volume.LastUpdated.Month() != now.Month() || volume.LastUpdated.Year() != now.Year() {
		// Reset for new month
		return TierStandard
	}

	return volume.Tier
}

// UpdateUserVolume updates user's monthly volume and recalculates tier
func (fs *FeeService) UpdateUserVolume(userID string, amount float64) FeeTier {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	volume, exists := fs.userVolumes[userID]
	if !exists {
		volume = &UserVolume{
			UserID:        userID,
			MonthlyVolume: 0,
			Tier:          TierStandard,
			LastUpdated:   time.Now(),
		}
		fs.userVolumes[userID] = volume
	}

	// Reset if new month
	now := time.Now()
	if volume.LastUpdated.Month() != now.Month() || volume.LastUpdated.Year() != now.Year() {
		volume.MonthlyVolume = 0
	}

	// Add to volume
	volume.MonthlyVolume += amount
	volume.LastUpdated = now

	// Update tier
	volume.Tier = fs.calculateTier(volume.MonthlyVolume)

	log.Printf("User %s volume: $%.2f -> Tier: %s", userID, volume.MonthlyVolume, volume.Tier)

	return volume.Tier
}

// calculateTier determines tier based on volume
func (fs *FeeService) calculateTier(volume float64) FeeTier {
	if volume >= 100000 {
		return TierGold
	} else if volume >= 50000 {
		return TierSilver
	} else if volume >= 10000 {
		return TierBronze
	}
	return TierStandard
}

// CalculateTradingFee calculates trading fee for a bet
func (fs *FeeService) CalculateTradingFee(userID string, amount float64) *FeeCalculation {
	tier := fs.GetUserTier(userID)
	feePercent := fs.config.TradingFees[tier]
	feeAmount := amount * feePercent

	calc := &FeeCalculation{
		Amount:     amount,
		FeeType:    "trading",
		FeePercent: feePercent * 100, // Convert to percentage
		FeeAmount:  feeAmount,
		TotalFee:   feeAmount,
		UserTier:   string(tier),
		Breakdown: map[string]float64{
			"trading_fee": feeAmount,
		},
	}

	return calc
}

// CalculateWithdrawalFee calculates withdrawal fee
func (fs *FeeService) CalculateWithdrawalFee(amount float64, currency string) *FeeCalculation {
	// Platform fee
	platformFee := amount * fs.config.WithdrawalFeePercent
	if platformFee < fs.config.WithdrawalFeeMin {
		platformFee = fs.config.WithdrawalFeeMin
	}

	// Network fee
	networkFee := fs.config.NetworkFees[currency]
	if networkFee == 0 {
		networkFee = 0.10 // Default network fee
	}

	totalFee := platformFee + networkFee

	calc := &FeeCalculation{
		Amount:     amount,
		FeeType:    "withdrawal",
		FeePercent: fs.config.WithdrawalFeePercent * 100,
		FeeAmount:  platformFee,
		NetworkFee: networkFee,
		TotalFee:   totalFee,
		UserTier:   "standard",
		Breakdown: map[string]float64{
			"platform_fee": platformFee,
			"network_fee":  networkFee,
		},
	}

	return calc
}

// RecordFee records a fee transaction for revenue tracking
func (fs *FeeService) RecordFee(userID string, feeType string, amount float64, feeAmount float64) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	tier := string(fs.GetUserTier(userID))

	// Update revenue stats
	fs.revenue.TotalRevenue += feeAmount
	fs.revenue.RevenueByType[feeType] += feeAmount
	fs.revenue.RevenueByTier[tier] += feeAmount
	fs.revenue.TotalVolume += amount
	fs.revenue.TransactionCount++
	fs.revenue.AverageRevenuePerTx = fs.revenue.TotalRevenue / float64(fs.revenue.TransactionCount)
	fs.revenue.LastUpdated = time.Now()

	log.Printf("Fee recorded: %s | Amount: $%.2f | Fee: $%.2f | Tier: %s",
		feeType, amount, feeAmount, tier)
}

// GetRevenueStats returns current revenue statistics
func (fs *FeeService) GetRevenueStats() *RevenueStats {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := &RevenueStats{
		TotalRevenue:        fs.revenue.TotalRevenue,
		RevenueByType:       make(map[string]float64),
		RevenueByTier:       make(map[string]float64),
		TotalVolume:         fs.revenue.TotalVolume,
		TransactionCount:    fs.revenue.TransactionCount,
		AverageRevenuePerTx: fs.revenue.AverageRevenuePerTx,
		LastUpdated:         fs.revenue.LastUpdated,
	}

	for k, v := range fs.revenue.RevenueByType {
		stats.RevenueByType[k] = v
	}
	for k, v := range fs.revenue.RevenueByTier {
		stats.RevenueByTier[k] = v
	}

	return stats
}

// GetUserStats returns stats for a specific user
func (fs *FeeService) GetUserStats(userID string) map[string]interface{} {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	volume, exists := fs.userVolumes[userID]
	if !exists {
		return map[string]interface{}{
			"user_id":        userID,
			"tier":           TierStandard,
			"monthly_volume": 0,
			"trading_fee":    fs.config.TradingFees[TierStandard] * 100,
		}
	}

	return map[string]interface{}{
		"user_id":        userID,
		"tier":           volume.Tier,
		"monthly_volume": volume.MonthlyVolume,
		"trading_fee":    fs.config.TradingFees[volume.Tier] * 100,
		"next_tier":      fs.getNextTier(volume.Tier),
		"volume_to_next": fs.getVolumeToNextTier(volume.MonthlyVolume, volume.Tier),
	}
}

func (fs *FeeService) getNextTier(current FeeTier) string {
	switch current {
	case TierStandard:
		return string(TierBronze)
	case TierBronze:
		return string(TierSilver)
	case TierSilver:
		return string(TierGold)
	case TierGold:
		return "max"
	default:
		return ""
	}
}

func (fs *FeeService) getVolumeToNextTier(volume float64, tier FeeTier) float64 {
	switch tier {
	case TierStandard:
		return 10000 - volume
	case TierBronze:
		return 50000 - volume
	case TierSilver:
		return 100000 - volume
	case TierGold:
		return 0
	default:
		return 0
	}
}

// HTTP Handlers

func (fs *FeeService) calculateTradingFeeHandler(c *gin.Context) {
	var req struct {
		UserID string  `json:"user_id" binding:"required"`
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	calc := fs.CalculateTradingFee(req.UserID, req.Amount)
	c.JSON(http.StatusOK, calc)
}

func (fs *FeeService) calculateWithdrawalFeeHandler(c *gin.Context) {
	var req struct {
		Amount   float64 `json:"amount" binding:"required,gt=0"`
		Currency string  `json:"currency" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	calc := fs.CalculateWithdrawalFee(req.Amount, req.Currency)
	c.JSON(http.StatusOK, calc)
}

func (fs *FeeService) recordFeeHandler(c *gin.Context) {
	var req struct {
		UserID    string  `json:"user_id" binding:"required"`
		FeeType   string  `json:"fee_type" binding:"required"`
		Amount    float64 `json:"amount" binding:"required,gt=0"`
		FeeAmount float64 `json:"fee_amount" binding:"required,gte=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user volume if it's a trading fee
	if req.FeeType == "trading" {
		fs.UpdateUserVolume(req.UserID, req.Amount)
	}

	fs.RecordFee(req.UserID, req.FeeType, req.Amount, req.FeeAmount)

	c.JSON(http.StatusOK, gin.H{
		"status":  "recorded",
		"message": "Fee recorded successfully",
	})
}

func (fs *FeeService) getRevenueStatsHandler(c *gin.Context) {
	stats := fs.GetRevenueStats()
	c.JSON(http.StatusOK, stats)
}

func (fs *FeeService) getUserStatsHandler(c *gin.Context) {
	userID := c.Param("user_id")
	stats := fs.GetUserStats(userID)
	c.JSON(http.StatusOK, stats)
}

func (fs *FeeService) getFeeTiersHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"tiers": []gin.H{
			{
				"name":          "Standard",
				"min_volume":    0,
				"max_volume":    9999,
				"trading_fee":   2.5,
				"benefits":      []string{"Standard trading fees", "24/7 support"},
			},
			{
				"name":          "Bronze",
				"min_volume":    10000,
				"max_volume":    49999,
				"trading_fee":   2.0,
				"benefits":      []string{"Reduced fees", "Priority support", "Email notifications"},
			},
			{
				"name":          "Silver",
				"min_volume":    50000,
				"max_volume":    99999,
				"trading_fee":   1.5,
				"benefits":      []string{"Low fees", "Premium support", "Advanced analytics"},
			},
			{
				"name":          "Gold",
				"min_volume":    100000,
				"max_volume":    999999999,
				"trading_fee":   1.0,
				"benefits":      []string{"Lowest fees", "VIP support", "API access", "Custom markets"},
			},
		},
	})
}

func main() {
	feeService := NewFeeService()

	// Simulate some demo activity
	feeService.UpdateUserVolume("user123", 15000) // Bronze tier
	feeService.RecordFee("user123", "trading", 1000, 20)
	feeService.RecordFee("user123", "trading", 500, 10)

	feeService.UpdateUserVolume("user456", 60000) // Silver tier
	feeService.RecordFee("user456", "trading", 2000, 30)

	router := gin.Default()

	// Fee calculation endpoints
	router.POST("/api/fees/calculate/trading", feeService.calculateTradingFeeHandler)
	router.POST("/api/fees/calculate/withdrawal", feeService.calculateWithdrawalFeeHandler)

	// Fee recording endpoint
	router.POST("/api/fees/record", feeService.recordFeeHandler)

	// Stats endpoints
	router.GET("/api/fees/revenue", feeService.getRevenueStatsHandler)
	router.GET("/api/fees/users/:user_id", feeService.getUserStatsHandler)
	router.GET("/api/fees/tiers", feeService.getFeeTiersHandler)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	log.Println("Fee Service running on :8087")
	log.Println("Fee Tiers:")
	log.Println("  Standard (0-10K): 2.5%")
	log.Println("  Bronze (10K-50K): 2.0%")
	log.Println("  Silver (50K-100K): 1.5%")
	log.Println("  Gold (100K+): 1.0%")
	router.Run(":8087")
}
