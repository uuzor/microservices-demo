// Copyright 2025 Vibe Markets
// Licensed under the Apache License, Version 2.0

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	genai "github.com/google/generative-ai-go/genai"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

var (
	log      *logrus.Logger
	port     = "8081"
	geminiClient *genai.Client
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

// Market represents a prediction market
type Market struct {
	ID                 string  `json:"id"`
	Title              string  `json:"title"`
	Description        string  `json:"description"`
	ResolutionCriteria string  `json:"resolution_criteria"`
	EndDate            int64   `json:"end_date"`
	Status             string  `json:"status"`
	YesPrice           float64 `json:"yes_price"`
	NoPrice            float64 `json:"no_price"`
	TotalVolume        float64 `json:"total_volume"`
	TotalTraders       int32   `json:"total_traders"`
	Category           string  `json:"category"`
}

// UserActivity represents user interaction history
type UserActivity struct {
	UserID        string   `json:"user_id"`
	ViewedMarkets []string `json:"viewed_markets"`
	BetMarkets    []string `json:"bet_markets"`
	Categories    []string `json:"categories"`
	TotalBets     int      `json:"total_bets"`
	TotalVolume   float64  `json:"total_volume"`
}

// Recommendation represents a market recommendation
type Recommendation struct {
	MarketID    string  `json:"market_id"`
	Score       float64 `json:"score"`
	Reason      string  `json:"reason"`
	Confidence  float64 `json:"confidence"`
}

// MarketSummary represents AI-generated market insights
type MarketSummary struct {
	MarketID    string    `json:"market_id"`
	Summary     string    `json:"summary"`
	KeyInsights []string  `json:"key_insights"`
	Sentiment   string    `json:"sentiment"` // "bullish", "bearish", "neutral"
	GeneratedAt time.Time `json:"generated_at"`
}

// AIService provides AI-powered features
type AIService struct {
	markets        []Market
	userActivities map[string]*UserActivity
	geminiModel    *genai.GenerativeModel
}

// NewAIService creates a new AI service
func NewAIService() *AIService {
	return &AIService{
		markets:        loadDemoMarkets(),
		userActivities: make(map[string]*UserActivity),
	}
}

// InitializeGemini sets up Gemini API client
func (ai *AIService) InitializeGemini(ctx context.Context, apiKey string) error {
	if apiKey == "" {
		log.Warn("Gemini API key not provided, summaries will use fallback")
		return nil
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}

	geminiClient = client
	ai.geminiModel = client.GenerativeModel("gemini-pro")

	// Configure model
	ai.geminiModel.SetTemperature(0.7)
	ai.geminiModel.SetTopK(40)
	ai.geminiModel.SetTopP(0.95)
	ai.geminiModel.SetMaxOutputTokens(1024)

	log.Info("Gemini client initialized successfully")
	return nil
}

// GetRecommendations returns personalized market recommendations
func (ai *AIService) GetRecommendations(userID string, limit int) []Recommendation {
	if limit == 0 {
		limit = 5
	}

	activity := ai.getUserActivity(userID)

	// Calculate scores for all markets
	var recommendations []Recommendation
	for _, market := range ai.markets {
		if market.Status != "open" {
			continue
		}

		score := ai.calculateRecommendationScore(market, activity)
		reason := ai.generateReason(market, activity)

		recommendations = append(recommendations, Recommendation{
			MarketID:   market.ID,
			Score:      score,
			Reason:     reason,
			Confidence: min(score/100.0, 1.0),
		})
	}

	// Sort by score descending
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	log.Infof("Generated %d recommendations for user %s", len(recommendations), userID)
	return recommendations
}

// calculateRecommendationScore uses collaborative filtering + content-based filtering
func (ai *AIService) calculateRecommendationScore(market Market, activity *UserActivity) float64 {
	score := 50.0 // Base score

	// Content-based: Category preference
	for _, cat := range activity.Categories {
		if cat == market.Category {
			score += 25.0
			break
		}
	}

	// Collaborative: Similar markets
	for _, betMarket := range activity.BetMarkets {
		if betMarket == market.ID {
			score -= 100.0 // Already bet on this market
			break
		}
	}

	// Popularity boost
	if market.TotalVolume > 10000 {
		score += 15.0
	}
	if market.TotalTraders > 100 {
		score += 10.0
	}

	// Balanced markets (close to 50/50) are more interesting
	balance := math.Abs(market.YesPrice - 0.5)
	if balance < 0.2 {
		score += 10.0
	}

	// Recent activity boost (markets ending soon)
	daysUntilEnd := float64(market.EndDate-time.Now().Unix()) / 86400.0
	if daysUntilEnd < 30 && daysUntilEnd > 0 {
		score += (30.0 - daysUntilEnd) / 3.0 // Up to 10 points
	}

	// Add some randomness for diversity
	score += rand.Float64() * 10.0

	return math.Max(0, score)
}

// generateReason creates a human-readable recommendation reason
func (ai *AIService) generateReason(market Market, activity *UserActivity) string {
	reasons := []string{}

	// Category match
	for _, cat := range activity.Categories {
		if cat == market.Category {
			reasons = append(reasons, fmt.Sprintf("You've shown interest in %s markets", cat))
			break
		}
	}

	// Popularity
	if market.TotalVolume > 10000 {
		reasons = append(reasons, "Highly active market")
	}

	// Balanced odds
	balance := math.Abs(market.YesPrice - 0.5)
	if balance < 0.2 {
		reasons = append(reasons, "Closely contested")
	}

	// Ending soon
	daysUntilEnd := float64(market.EndDate-time.Now().Unix()) / 86400.0
	if daysUntilEnd < 30 && daysUntilEnd > 0 {
		reasons = append(reasons, fmt.Sprintf("Resolves in %.0f days", daysUntilEnd))
	}

	if len(reasons) == 0 {
		return "Popular in your category"
	}

	return strings.Join(reasons, " • ")
}

// GenerateMarketSummary creates AI-powered market insights using Gemini
func (ai *AIService) GenerateMarketSummary(ctx context.Context, marketID string) (*MarketSummary, error) {
	// Find market
	var market *Market
	for _, m := range ai.markets {
		if m.ID == marketID {
			market = &m
			break
		}
	}

	if market == nil {
		return nil, fmt.Errorf("market not found")
	}

	// Use Gemini if available, otherwise fallback
	if ai.geminiModel != nil {
		return ai.generateSummaryWithGemini(ctx, market)
	}

	return ai.generateFallbackSummary(market), nil
}

// generateSummaryWithGemini uses Gemini AI to generate market summary
func (ai *AIService) generateSummaryWithGemini(ctx context.Context, market *Market) (*MarketSummary, error) {
	prompt := fmt.Sprintf(`You are an expert prediction market analyst. Analyze this prediction market and provide insights:

Market: %s
Description: %s
Resolution Criteria: %s
Current Odds: YES %.0f%% / NO %.0f%%
Total Volume: $%.2f
Total Traders: %d
Category: %s
Days until resolution: %.0f

Provide:
1. A brief 2-3 sentence summary of the market
2. 3 key insights about the current odds and activity
3. Overall sentiment (bullish/bearish/neutral)

Keep it concise and actionable for traders.`,
		market.Title,
		market.Description,
		market.ResolutionCriteria,
		market.YesPrice*100,
		market.NoPrice*100,
		market.TotalVolume,
		market.TotalTraders,
		market.Category,
		float64(market.EndDate-time.Now().Unix())/86400.0,
	)

	resp, err := ai.geminiModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Errorf("Gemini API error: %v", err)
		return ai.generateFallbackSummary(market), nil // Fallback on error
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return ai.generateFallbackSummary(market), nil
	}

	// Extract text from response
	summaryText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Parse insights (simple line split)
	lines := strings.Split(summaryText, "\n")
	insights := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			insights = append(insights, trimmed)
		}
	}

	// Determine sentiment from odds
	sentiment := "neutral"
	if market.YesPrice > 0.65 {
		sentiment = "bullish"
	} else if market.YesPrice < 0.35 {
		sentiment = "bearish"
	}

	summary := &MarketSummary{
		MarketID:    market.ID,
		Summary:     summaryText,
		KeyInsights: insights[:min(3, len(insights))],
		Sentiment:   sentiment,
		GeneratedAt: time.Now(),
	}

	log.Infof("Generated Gemini summary for market %s", marketID)
	return summary, nil
}

// generateFallbackSummary creates a rule-based summary when Gemini is unavailable
func (ai *AIService) generateFallbackSummary(market *Market) *MarketSummary {
	// Sentiment based on odds
	sentiment := "neutral"
	if market.YesPrice > 0.65 {
		sentiment = "bullish"
	} else if market.YesPrice < 0.35 {
		sentiment = "bearish"
	}

	// Generate summary text
	summary := fmt.Sprintf("%s is currently trading at %.0f%% YES / %.0f%% NO with $%.0f in total volume from %d traders. ",
		market.Title,
		market.YesPrice*100,
		market.NoPrice*100,
		market.TotalVolume,
		market.TotalTraders,
	)

	if sentiment == "bullish" {
		summary += "The market shows strong confidence in a YES outcome. "
	} else if sentiment == "bearish" {
		summary += "Traders are skeptical, leaning toward a NO outcome. "
	} else {
		summary += "The market is closely contested with balanced sentiment. "
	}

	daysUntilEnd := float64(market.EndDate-time.Now().Unix()) / 86400.0
	if daysUntilEnd < 30 {
		summary += fmt.Sprintf("Resolution expected in %.0f days.", daysUntilEnd)
	}

	// Key insights
	insights := []string{
		fmt.Sprintf("Current probability: %.0f%% YES", market.YesPrice*100),
		fmt.Sprintf("$%.0f traded by %d participants", market.TotalVolume, market.TotalTraders),
	}

	if market.TotalVolume > 10000 {
		insights = append(insights, "High liquidity market")
	} else {
		insights = append(insights, "Developing market with room to grow")
	}

	return &MarketSummary{
		MarketID:    market.ID,
		Summary:     summary,
		KeyInsights: insights,
		Sentiment:   sentiment,
		GeneratedAt: time.Now(),
	}
}

// TrackUserActivity records user interactions
func (ai *AIService) TrackUserActivity(userID, action, marketID, category string) {
	activity := ai.getUserActivity(userID)

	switch action {
	case "view":
		if !contains(activity.ViewedMarkets, marketID) {
			activity.ViewedMarkets = append(activity.ViewedMarkets, marketID)
		}
	case "bet":
		if !contains(activity.BetMarkets, marketID) {
			activity.BetMarkets = append(activity.BetMarkets, marketID)
			activity.TotalBets++
		}
	}

	if category != "" && !contains(activity.Categories, category) {
		activity.Categories = append(activity.Categories, category)
	}
}

// getUserActivity gets or creates user activity
func (ai *AIService) getUserActivity(userID string) *UserActivity {
	if activity, exists := ai.userActivities[userID]; exists {
		return activity
	}

	activity := &UserActivity{
		UserID:        userID,
		ViewedMarkets: []string{},
		BetMarkets:    []string{},
		Categories:    []string{},
		TotalBets:     0,
		TotalVolume:   0,
	}
	ai.userActivities[userID] = activity
	return activity
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// loadDemoMarkets loads markets from Week 1
func loadDemoMarkets() []Market {
	// Load from file or return demo data
	return []Market{
		{
			ID: "btc-150k-jun2025", Title: "Will Bitcoin reach $150,000 by June 2025?",
			Category: "crypto", YesPrice: 0.48, NoPrice: 0.52, TotalVolume: 12500, TotalTraders: 234,
			EndDate: 1751241600, Status: "open",
		},
		{
			ID: "eth-5k-dec2025", Title: "Will Ethereum reach $5,000 by end of 2025?",
			Category: "crypto", YesPrice: 0.62, NoPrice: 0.38, TotalVolume: 8900, TotalTraders: 156,
			EndDate: 1767225600, Status: "open",
		},
		{
			ID: "superbowl-chiefs-2026", Title: "Will Kansas City Chiefs win Super Bowl LX?",
			Category: "sports", YesPrice: 0.35, NoPrice: 0.65, TotalVolume: 15200, TotalTraders: 412,
			EndDate: 1738627200, Status: "open",
		},
		{
			ID: "trump-president-2025", Title: "Will Donald Trump be inaugurated in January 2025?",
			Category: "politics", YesPrice: 0.89, NoPrice: 0.11, TotalVolume: 45300, TotalTraders: 1823,
			EndDate: 1737417600, Status: "open",
		},
		{
			ID: "ai-agi-2025", Title: "Will OpenAI announce AGI in 2025?",
			Category: "tech", YesPrice: 0.15, NoPrice: 0.85, TotalVolume: 6700, TotalTraders: 289,
			EndDate: 1767225600, Status: "open",
		},
	}
}

// REST API handlers
func setupRoutes(ai *AIService) *gin.Engine {
	router := gin.Default()

	// CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "ai"})
	})

	// Get personalized recommendations
	router.GET("/api/recommendations/:user_id", func(c *gin.Context) {
		userID := c.Param("user_id")
		limit := 5
		fmt.Sscanf(c.Query("limit"), "%d", &limit)

		recommendations := ai.GetRecommendations(userID, limit)
		c.JSON(200, gin.H{
			"user_id":         userID,
			"recommendations": recommendations,
			"generated_at":    time.Now(),
		})
	})

	// Get market summary
	router.GET("/api/summary/:market_id", func(c *gin.Context) {
		marketID := c.Param("market_id")

		summary, err := ai.GenerateMarketSummary(c.Request.Context(), marketID)
		if err != nil {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, summary)
	})

	// Track user activity
	router.POST("/api/activity", func(c *gin.Context) {
		var req struct {
			UserID   string `json:"user_id"`
			Action   string `json:"action"`
			MarketID string `json:"market_id"`
			Category string `json:"category"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		ai.TrackUserActivity(req.UserID, req.Action, req.MarketID, req.Category)
		c.JSON(200, gin.H{"status": "recorded"})
	})

	// Batch summaries
	router.POST("/api/summaries/batch", func(c *gin.Context) {
		var req struct {
			MarketIDs []string `json:"market_ids"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		summaries := make([]*MarketSummary, 0)
		for _, marketID := range req.MarketIDs {
			summary, err := ai.GenerateMarketSummary(c.Request.Context(), marketID)
			if err == nil {
				summaries = append(summaries, summary)
			}
		}

		c.JSON(200, gin.H{
			"summaries": summaries,
			"count":     len(summaries),
		})
	})

	return router
}

func main() {
	flag.Parse()

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	ai := NewAIService()

	// Initialize Gemini
	ctx := context.Background()
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey != "" {
		err := ai.InitializeGemini(ctx, geminiKey)
		if err != nil {
			log.Warnf("Failed to initialize Gemini: %v", err)
		}
	} else {
		log.Warn("GEMINI_API_KEY not set, using fallback summaries")
	}

	router := setupRoutes(ai)

	log.Infof("Starting AI Service on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
