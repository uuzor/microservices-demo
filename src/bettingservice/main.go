// Copyright 2025 Vibe Markets
// Licensed under the Apache License, Version 2.0

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	log  *logrus.Logger
	port = "3551"
)

func init() {
	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout
}

// MarketState represents the liquidity pools for a market
type MarketState struct {
	MarketID string
	YesPool  float64 // Liquidity in YES pool
	NoPool   float64 // Liquidity in NO pool
	K        float64 // Constant product (k = x * y)
	mu       sync.RWMutex
}

// BetResult contains the result of a bet execution
type BetResult struct {
	BetID           string
	SharesReceived  float64
	AveragePrice    float64
	NewMarketPrice  float64
	NewYesPrice     float64
	NewNoPrice      float64
	ExecutedAt      time.Time
	Status          string
	ErrorMessage    string
}

// UserPosition tracks a user's position in a market
type UserPosition struct {
	UserID      string
	MarketID    string
	Side        string  // "YES" or "NO"
	Shares      float64
	AverageCost float64
	CreatedAt   time.Time
}

// BettingEngine manages all betting operations with AMM
type BettingEngine struct {
	markets   map[string]*MarketState
	positions map[string][]UserPosition // key: userID
	mu        sync.RWMutex
}

// NewBettingEngine creates a new betting engine
func NewBettingEngine() *BettingEngine {
	return &BettingEngine{
		markets:   make(map[string]*MarketState),
		positions: make(map[string][]UserPosition),
	}
}

// InitializeMarket sets up initial liquidity for a market
func (be *BettingEngine) InitializeMarket(marketID string, initialLiquidity float64) {
	be.mu.Lock()
	defer be.mu.Unlock()

	if _, exists := be.markets[marketID]; exists {
		return // Market already initialized
	}

	// Start with equal liquidity (50/50 odds)
	be.markets[marketID] = &MarketState{
		MarketID: marketID,
		YesPool:  initialLiquidity,
		NoPool:   initialLiquidity,
		K:        initialLiquidity * initialLiquidity,
	}

	log.Infof("Initialized market %s with %.2f liquidity (k=%.2f)", marketID, initialLiquidity, initialLiquidity*initialLiquidity)
}

// PlaceBet executes a bet using Constant Product Market Maker (CPMM)
func (be *BettingEngine) PlaceBet(userID, marketID, side string, amount float64) (*BetResult, error) {
	if amount <= 0 {
		return nil, errors.New("bet amount must be positive")
	}

	if side != "YES" && side != "NO" {
		return nil, errors.New("side must be YES or NO")
	}

	be.mu.Lock()
	defer be.mu.Unlock()

	market, exists := be.markets[marketID]
	if !exists {
		// Auto-initialize with default liquidity
		be.markets[marketID] = &MarketState{
			MarketID: marketID,
			YesPool:  1000.0,
			NoPool:   1000.0,
			K:        1000000.0,
		}
		market = be.markets[marketID]
		log.Infof("Auto-initialized market %s", marketID)
	}

	market.mu.Lock()
	defer market.mu.Unlock()

	var shares float64
	var avgPrice float64

	// CPMM Formula: k = x * y (constant)
	// When buying YES shares, we add to YES pool and calculate new NO pool
	// Shares received = change in NO pool

	if side == "YES" {
		// User buys YES by adding to YES pool
		newYesPool := market.YesPool + amount
		newNoPool := market.K / newYesPool

		// Shares = reduction in NO pool
		shares = market.NoPool - newNoPool
		avgPrice = amount / shares

		// Update pools
		market.YesPool = newYesPool
		market.NoPool = newNoPool

		log.Infof("YES bet: YesPool %.2f->%.2f, NoPool %.2f->%.2f, Shares %.4f",
			market.YesPool-amount, newYesPool, market.NoPool+shares, newNoPool, shares)

	} else {
		// User buys NO by adding to NO pool
		newNoPool := market.NoPool + amount
		newYesPool := market.K / newNoPool

		// Shares = reduction in YES pool
		shares = market.YesPool - newYesPool
		avgPrice = amount / shares

		// Update pools
		market.YesPool = newYesPool
		market.NoPool = newNoPool

		log.Infof("NO bet: NoPool %.2f->%.2f, YesPool %.2f->%.2f, Shares %.4f",
			market.NoPool-amount, newNoPool, market.YesPool+shares, newYesPool, shares)
	}

	// Calculate new prices (probability)
	totalPool := market.YesPool + market.NoPool
	yesPrice := market.YesPool / totalPool
	noPrice := market.NoPool / totalPool

	// Record position
	position := UserPosition{
		UserID:      userID,
		MarketID:    marketID,
		Side:        side,
		Shares:      shares,
		AverageCost: avgPrice,
		CreatedAt:   time.Now(),
	}

	be.positions[userID] = append(be.positions[userID], position)

	result := &BetResult{
		BetID:          uuid.New().String(),
		SharesReceived: shares,
		AveragePrice:   avgPrice,
		NewMarketPrice: yesPrice,
		NewYesPrice:    yesPrice,
		NewNoPrice:     noPrice,
		ExecutedAt:     time.Now(),
		Status:         "success",
	}

	log.Infof("Bet executed: User=%s, Market=%s, Side=%s, Amount=%.2f, Shares=%.4f, Price=%.4f, NewYesPrice=%.4f",
		userID, marketID, side, amount, shares, avgPrice, yesPrice)

	return result, nil
}

// GetCurrentPrice returns the current market price
func (be *BettingEngine) GetCurrentPrice(marketID string) (yesPrice, noPrice float64, err error) {
	be.mu.RLock()
	defer be.mu.RUnlock()

	market, exists := be.markets[marketID]
	if !exists {
		return 0.5, 0.5, nil // Default 50/50 for uninitialized markets
	}

	market.mu.RLock()
	defer market.mu.RUnlock()

	total := market.YesPool + market.NoPool
	yesPrice = market.YesPool / total
	noPrice = market.NoPool / total

	return yesPrice, noPrice, nil
}

// GetUserPositions returns all positions for a user
func (be *BettingEngine) GetUserPositions(userID string, marketID string) []UserPosition {
	be.mu.RLock()
	defer be.mu.RUnlock()

	positions := be.positions[userID]
	if marketID == "" {
		return positions
	}

	// Filter by market
	var filtered []UserPosition
	for _, pos := range positions {
		if pos.MarketID == marketID {
			filtered = append(filtered, pos)
		}
	}
	return filtered
}

// CalculatePositionValue calculates current value of a position
func (be *BettingEngine) CalculatePositionValue(position UserPosition) (currentValue, pnl float64, err error) {
	yesPrice, noPrice, err := be.GetCurrentPrice(position.MarketID)
	if err != nil {
		return 0, 0, err
	}

	var marketPrice float64
	if position.Side == "YES" {
		marketPrice = yesPrice
	} else {
		marketPrice = noPrice
	}

	currentValue = position.Shares * marketPrice
	pnl = currentValue - (position.Shares * position.AverageCost)

	return currentValue, pnl, nil
}

// GetMarketStats returns statistics for a market
func (be *BettingEngine) GetMarketStats(marketID string) map[string]interface{} {
	be.mu.RLock()
	defer be.mu.RUnlock()

	market, exists := be.markets[marketID]
	if !exists {
		return map[string]interface{}{
			"error": "market not found",
		}
	}

	market.mu.RLock()
	defer market.mu.RUnlock()

	total := market.YesPool + market.NoPool
	yesPrice := market.YesPool / total
	noPrice := market.NoPool / total

	return map[string]interface{}{
		"market_id":  marketID,
		"yes_pool":   market.YesPool,
		"no_pool":    market.NoPool,
		"k":          market.K,
		"yes_price":  yesPrice,
		"no_price":   noPrice,
		"total_pool": total,
	}
}

type bettingService struct {
	engine *BettingEngine
}

func (b *bettingService) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (b *bettingService) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return errors.New("health check via Watch not implemented")
}

func main() {
	flag.Parse()

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	log.Infof("starting Betting Service gRPC server at :%s", port)
	run(port)
	select {}
}

func run(port string) string {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer()

	engine := NewBettingEngine()

	// Initialize some demo markets with initial liquidity
	demoMarkets := []string{
		"btc-150k-jun2025",
		"eth-5k-dec2025",
		"superbowl-chiefs-2026",
		"trump-president-2025",
		"ai-agi-2025",
		"spacex-mars-2025",
		"gta6-release-2025",
		"stock-spy-500-2025",
		"arsenal-epl-2025",
		"solana-200-2025",
	}

	for _, marketID := range demoMarkets {
		engine.InitializeMarket(marketID, 1000.0)
	}

	svc := &bettingService{engine: engine}
	healthpb.RegisterHealthServer(srv, svc)

	go srv.Serve(listener)

	log.Info("Betting Service started successfully")
	return listener.Addr().String()
}
