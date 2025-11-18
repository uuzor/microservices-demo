// Copyright 2025 Vibe Markets
// Licensed under the Apache License, Version 2.0

package main

import (
	"fmt"
	"testing"
)

func TestAMMBasic(t *testing.T) {
	engine := NewBettingEngine()

	// Initialize market with 1000 liquidity (50/50 odds)
	marketID := "test-market-1"
	engine.InitializeMarket(marketID, 1000.0)

	// Check initial price
	yesPrice, noPrice, _ := engine.GetCurrentPrice(marketID)
	if yesPrice != 0.5 || noPrice != 0.5 {
		t.Errorf("Initial prices should be 0.5/0.5, got %.2f/%.2f", yesPrice, noPrice)
	}

	fmt.Printf("\nInitial State:\n")
	fmt.Printf("YES Price: %.4f (%.1f%%)\n", yesPrice, yesPrice*100)
	fmt.Printf("NO Price: %.4f (%.1f%%)\n", noPrice, noPrice*100)

	// User bets $100 on YES
	result, err := engine.PlaceBet("user1", marketID, "YES", 100.0)
	if err != nil {
		t.Fatalf("Bet failed: %v", err)
	}

	fmt.Printf("\nAfter $100 bet on YES:\n")
	fmt.Printf("Shares received: %.4f\n", result.SharesReceived)
	fmt.Printf("Average price: $%.4f per share\n", result.AveragePrice)
	fmt.Printf("New YES price: %.4f (%.1f%%)\n", result.NewYesPrice, result.NewYesPrice*100)
	fmt.Printf("New NO price: %.4f (%.1f%%)\n", result.NewNoPrice, result.NewNoPrice*100)

	// Price should have moved in favor of YES
	if result.NewYesPrice <= 0.5 {
		t.Errorf("YES price should increase after YES bet, got %.4f", result.NewYesPrice)
	}

	// Another user bets $200 on NO
	result2, err := engine.PlaceBet("user2", marketID, "NO", 200.0)
	if err != nil {
		t.Fatalf("Second bet failed: %v", err)
	}

	fmt.Printf("\nAfter $200 bet on NO:\n")
	fmt.Printf("Shares received: %.4f\n", result2.SharesReceived)
	fmt.Printf("Average price: $%.4f per share\n", result2.AveragePrice)
	fmt.Printf("New YES price: %.4f (%.1f%%)\n", result2.NewYesPrice, result2.NewYesPrice*100)
	fmt.Printf("New NO price: %.4f (%.1f%%)\n", result2.NewNoPrice, result2.NewNoPrice*100)

	// Print market stats
	stats := engine.GetMarketStats(marketID)
	fmt.Printf("\nFinal Market Stats:\n")
	fmt.Printf("YES Pool: $%.2f\n", stats["yes_pool"])
	fmt.Printf("NO Pool: $%.2f\n", stats["no_pool"])
	fmt.Printf("Total Volume: $%.2f\n", 100.0+200.0)
	fmt.Printf("K (constant): %.2f\n", stats["k"])

	// Check positions
	positions := engine.GetUserPositions("user1", marketID)
	if len(positions) != 1 {
		t.Errorf("User1 should have 1 position, got %d", len(positions))
	}

	fmt.Printf("\nUser1 Position:\n")
	fmt.Printf("Side: %s\n", positions[0].Side)
	fmt.Printf("Shares: %.4f\n", positions[0].Shares)
	fmt.Printf("Average Cost: $%.4f\n", positions[0].AverageCost)

	currentValue, pnl, _ := engine.CalculatePositionValue(positions[0])
	fmt.Printf("Current Value: $%.2f\n", currentValue)
	fmt.Printf("P&L: $%.2f (%.1f%%)\n", pnl, (pnl/(positions[0].Shares*positions[0].AverageCost))*100)
}

func TestAMMMultipleBets(t *testing.T) {
	engine := NewBettingEngine()
	marketID := "test-market-2"
	engine.InitializeMarket(marketID, 1000.0)

	fmt.Printf("\n\n=== Testing Multiple Bets ===\n")

	// Simulate 10 bets
	bets := []struct {
		userID string
		side   string
		amount float64
	}{
		{"alice", "YES", 50},
		{"bob", "YES", 100},
		{"charlie", "NO", 75},
		{"dave", "YES", 150},
		{"eve", "NO", 200},
		{"frank", "YES", 80},
		{"grace", "NO", 120},
		{"henry", "YES", 60},
		{"iris", "NO", 90},
		{"jack", "YES", 110},
	}

	for i, bet := range bets {
		result, err := engine.PlaceBet(bet.userID, marketID, bet.side, bet.amount)
		if err != nil {
			t.Fatalf("Bet %d failed: %v", i+1, err)
		}

		fmt.Printf("\nBet %d: %s bets $%.0f on %s\n", i+1, bet.userID, bet.amount, bet.side)
		fmt.Printf("  Shares: %.4f @ $%.4f avg\n", result.SharesReceived, result.AveragePrice)
		fmt.Printf("  Market: YES %.1f%% | NO %.1f%%\n",
			result.NewYesPrice*100, result.NewNoPrice*100)
	}

	// Final stats
	stats := engine.GetMarketStats(marketID)
	totalVolume := 50 + 100 + 75 + 150 + 200 + 80 + 120 + 60 + 90 + 110

	fmt.Printf("\n=== Final Market State ===\n")
	fmt.Printf("Total Volume: $%d\n", totalVolume)
	fmt.Printf("YES Pool: $%.2f\n", stats["yes_pool"])
	fmt.Printf("NO Pool: $%.2f\n", stats["no_pool"])
	fmt.Printf("Final Odds: YES %.1f%% | NO %.1f%%\n",
		stats["yes_price"].(float64)*100, stats["no_price"].(float64)*100)

	// Check Alice's position
	positions := engine.GetUserPositions("alice", marketID)
	currentValue, pnl, _ := engine.CalculatePositionValue(positions[0])

	fmt.Printf("\n=== Alice's Position ===\n")
	fmt.Printf("Initial Bet: $%.0f on %s\n", 50.0, positions[0].Side)
	fmt.Printf("Shares Held: %.4f\n", positions[0].Shares)
	fmt.Printf("Current Value: $%.2f\n", currentValue)
	fmt.Printf("P&L: $%.2f (%.1f%%)\n", pnl, (pnl/50.0)*100)
}

func TestAMMPriceMovement(t *testing.T) {
	engine := NewBettingEngine()
	marketID := "test-market-3"
	engine.InitializeMarket(marketID, 1000.0)

	fmt.Printf("\n\n=== Testing Price Movement ===\n")

	// Show how price moves with different bet sizes
	betSizes := []float64{10, 50, 100, 250, 500, 1000}

	for _, size := range betSizes {
		// Reset market
		engine.InitializeMarket(marketID, 1000.0)

		result, _ := engine.PlaceBet("test-user", marketID, "YES", size)

		fmt.Printf("$%.0f bet on YES: %.1f%% → %.1f%% (%.1f%% move)\n",
			size, 50.0, result.NewYesPrice*100, (result.NewYesPrice-0.5)*100)
	}
}
