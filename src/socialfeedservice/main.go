// Copyright 2025 Vibe Markets
// Licensed under the Apache License, Version 2.0

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var (
	log  *logrus.Logger
	port = "8080"
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

// FeedEvent represents an activity in the social feed
type FeedEvent struct {
	ID        string                 `json:"id"`
	EventType string                 `json:"event_type"` // "bet_placed", "market_update", "win", "market_created"
	UserID    string                 `json:"user_id"`
	MarketID  string                 `json:"market_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
	CreatedAt time.Time              `json:"created_at"`
}

// UserProfile represents a user's social profile
type UserProfile struct {
	UserID        string   `json:"user_id"`
	Username      string   `json:"username"`
	Following     []string `json:"following"`
	Followers     []string `json:"followers"`
	TotalBets     int      `json:"total_bets"`
	TotalVolume   float64  `json:"total_volume"`
	TotalPnL      float64  `json:"total_pnl"`
	WinRate       float64  `json:"win_rate"`
	Rank          int      `json:"rank"`
	JoinedAt      int64    `json:"joined_at"`
}

// SocialFeedService manages the social activity feed
type SocialFeedService struct {
	events        []FeedEvent
	profiles      map[string]*UserProfile
	mu            sync.RWMutex
	wsClients     map[*websocket.Conn]bool
	wsClientsMu   sync.RWMutex
	kafkaConsumer *kafka.Consumer
	upgrader      websocket.Upgrader
}

// NewSocialFeedService creates a new social feed service
func NewSocialFeedService() *SocialFeedService {
	return &SocialFeedService{
		events:    make([]FeedEvent, 0),
		profiles:  make(map[string]*UserProfile),
		wsClients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for demo
			},
		},
	}
}

// AddEvent adds a new event to the feed
func (sfs *SocialFeedService) AddEvent(event FeedEvent) {
	sfs.mu.Lock()
	defer sfs.mu.Unlock()

	event.ID = uuid.New().String()
	event.CreatedAt = time.Now()

	sfs.events = append([]FeedEvent{event}, sfs.events...) // Prepend (newest first)

	// Keep only last 1000 events
	if len(sfs.events) > 1000 {
		sfs.events = sfs.events[:1000]
	}

	// Update user profile stats
	sfs.updateUserProfile(event)

	// Broadcast to WebSocket clients
	go sfs.broadcastEvent(event)

	log.Infof("Added feed event: %s by %s on market %s", event.EventType, event.UserID, event.MarketID)
}

// GetFeed returns the activity feed with optional filtering
func (sfs *SocialFeedService) GetFeed(userID string, limit, offset int) []FeedEvent {
	sfs.mu.RLock()
	defer sfs.mu.RUnlock()

	var filtered []FeedEvent

	if userID == "" {
		// Global feed
		filtered = sfs.events
	} else {
		// Personalized feed (user + following)
		profile := sfs.profiles[userID]
		following := make(map[string]bool)
		following[userID] = true
		if profile != nil {
			for _, uid := range profile.Following {
				following[uid] = true
			}
		}

		for _, event := range sfs.events {
			if following[event.UserID] {
				filtered = append(filtered, event)
			}
		}
	}

	// Apply pagination
	if limit == 0 {
		limit = 50
	}
	start := offset
	end := offset + limit
	if start > len(filtered) {
		return []FeedEvent{}
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end]
}

// GetUserProfile returns a user's profile
func (sfs *SocialFeedService) GetUserProfile(userID string) *UserProfile {
	sfs.mu.RLock()
	defer sfs.mu.RUnlock()

	profile := sfs.profiles[userID]
	if profile == nil {
		// Create default profile
		profile = &UserProfile{
			UserID:      userID,
			Username:    fmt.Sprintf("User_%s", userID[:8]),
			Following:   []string{},
			Followers:   []string{},
			TotalBets:   0,
			TotalVolume: 0,
			TotalPnL:    0,
			WinRate:     0,
			Rank:        0,
			JoinedAt:    time.Now().Unix(),
		}
	}
	return profile
}

// FollowUser adds a following relationship
func (sfs *SocialFeedService) FollowUser(followerID, followeeID string) error {
	sfs.mu.Lock()
	defer sfs.mu.Unlock()

	if followerID == followeeID {
		return fmt.Errorf("cannot follow yourself")
	}

	// Get or create profiles
	follower := sfs.getOrCreateProfile(followerID)
	followee := sfs.getOrCreateProfile(followeeID)

	// Check if already following
	for _, uid := range follower.Following {
		if uid == followeeID {
			return fmt.Errorf("already following")
		}
	}

	// Add following/follower
	follower.Following = append(follower.Following, followeeID)
	followee.Followers = append(followee.Followers, followerID)

	log.Infof("User %s followed %s", followerID, followeeID)
	return nil
}

// UnfollowUser removes a following relationship
func (sfs *SocialFeedService) UnfollowUser(followerID, followeeID string) error {
	sfs.mu.Lock()
	defer sfs.mu.Unlock()

	follower := sfs.profiles[followerID]
	followee := sfs.profiles[followeeID]

	if follower == nil || followee == nil {
		return fmt.Errorf("user not found")
	}

	// Remove from following
	follower.Following = removeString(follower.Following, followeeID)
	followee.Followers = removeString(followee.Followers, followerID)

	log.Infof("User %s unfollowed %s", followerID, followeeID)
	return nil
}

// GetLeaderboard returns top users by P&L
func (sfs *SocialFeedService) GetLeaderboard(limit int) []*UserProfile {
	sfs.mu.RLock()
	defer sfs.mu.RUnlock()

	profiles := make([]*UserProfile, 0, len(sfs.profiles))
	for _, profile := range sfs.profiles {
		profiles = append(profiles, profile)
	}

	// Sort by total P&L descending
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].TotalPnL > profiles[j].TotalPnL
	})

	// Update ranks
	for i, profile := range profiles {
		profile.Rank = i + 1
	}

	if limit == 0 {
		limit = 10
	}
	if limit > len(profiles) {
		limit = len(profiles)
	}

	return profiles[:limit]
}

// updateUserProfile updates user stats based on event
func (sfs *SocialFeedService) updateUserProfile(event FeedEvent) {
	profile := sfs.getOrCreateProfile(event.UserID)

	switch event.EventType {
	case "bet_placed":
		profile.TotalBets++
		if amount, ok := event.Data["amount"].(float64); ok {
			profile.TotalVolume += amount
		}
	case "win":
		if pnl, ok := event.Data["pnl"].(float64); ok {
			profile.TotalPnL += pnl
		}
	}
}

// getOrCreateProfile gets or creates a user profile (must be called with lock held)
func (sfs *SocialFeedService) getOrCreateProfile(userID string) *UserProfile {
	profile := sfs.profiles[userID]
	if profile == nil {
		profile = &UserProfile{
			UserID:      userID,
			Username:    fmt.Sprintf("User_%s", userID[:min(8, len(userID))]),
			Following:   []string{},
			Followers:   []string{},
			TotalBets:   0,
			TotalVolume: 0,
			TotalPnL:    0,
			WinRate:     0,
			Rank:        0,
			JoinedAt:    time.Now().Unix(),
		}
		sfs.profiles[userID] = profile
	}
	return profile
}

// broadcastEvent sends event to all WebSocket clients
func (sfs *SocialFeedService) broadcastEvent(event FeedEvent) {
	sfs.wsClientsMu.RLock()
	defer sfs.wsClientsMu.RUnlock()

	data, _ := json.Marshal(event)
	for client := range sfs.wsClients {
		err := client.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Errorf("WebSocket write error: %v", err)
			client.Close()
			delete(sfs.wsClients, client)
		}
	}
}

// handleWebSocket handles WebSocket connections
func (sfs *SocialFeedService) handleWebSocket(c *gin.Context) {
	conn, err := sfs.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	sfs.wsClientsMu.Lock()
	sfs.wsClients[conn] = true
	sfs.wsClientsMu.Unlock()

	log.Info("New WebSocket client connected")

	// Send recent events
	recentEvents := sfs.GetFeed("", 10, 0)
	for _, event := range recentEvents {
		data, _ := json.Marshal(event)
		conn.WriteMessage(websocket.TextMessage, data)
	}

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			sfs.wsClientsMu.Lock()
			delete(sfs.wsClients, conn)
			sfs.wsClientsMu.Unlock()
			log.Info("WebSocket client disconnected")
			break
		}
	}
}

// StartKafkaConsumer starts consuming from Kafka topics
func (sfs *SocialFeedService) StartKafkaConsumer(bootstrapServers, apiKey, apiSecret string) error {
	config := &kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"security.protocol": "SASL_SSL",
		"sasl.mechanisms":   "PLAIN",
		"sasl.username":     apiKey,
		"sasl.password":     apiSecret,
		"group.id":          "social-feed-consumer",
		"auto.offset.reset": "earliest",
	}

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	sfs.kafkaConsumer = consumer

	// Subscribe to topics
	topics := []string{"bet-orders", "market-updates", "social-feed-events"}
	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	log.Infof("Subscribed to Kafka topics: %v", topics)

	// Start consuming in background
	go func() {
		for {
			msg, err := consumer.ReadMessage(-1)
			if err != nil {
				log.Errorf("Consumer error: %v", err)
				continue
			}

			sfs.processKafkaMessage(msg)
		}
	}()

	return nil
}

// processKafkaMessage processes a Kafka message and creates feed events
func (sfs *SocialFeedService) processKafkaMessage(msg *kafka.Message) {
	var data map[string]interface{}
	err := json.Unmarshal(msg.Value, &data)
	if err != nil {
		log.Errorf("Failed to unmarshal message: %v", err)
		return
	}

	topic := *msg.TopicPartition.Topic

	var event FeedEvent
	event.Timestamp = time.Now().Unix()
	event.Data = data

	switch topic {
	case "bet-orders":
		event.EventType = "bet_placed"
		event.UserID = getString(data, "user_id")
		event.MarketID = getString(data, "market_id")

	case "market-updates":
		event.EventType = "market_update"
		event.MarketID = getString(data, "market_id")
		event.UserID = "system"

	case "social-feed-events":
		event.EventType = getString(data, "event_type")
		event.UserID = getString(data, "user_id")
		event.MarketID = getString(data, "market_id")
	}

	sfs.AddEvent(event)
}

// Helper functions
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func removeString(slice []string, s string) []string {
	result := make([]string, 0)
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// REST API Handlers
func setupRoutes(sfs *SocialFeedService) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Feed endpoints
	router.GET("/api/feed", func(c *gin.Context) {
		userID := c.Query("user_id")
		limit := 50
		offset := 0
		fmt.Sscanf(c.Query("limit"), "%d", &limit)
		fmt.Sscanf(c.Query("offset"), "%d", &offset)

		events := sfs.GetFeed(userID, limit, offset)
		c.JSON(200, gin.H{
			"events": events,
			"total":  len(events),
		})
	})

	// User profile
	router.GET("/api/users/:user_id", func(c *gin.Context) {
		userID := c.Param("user_id")
		profile := sfs.GetUserProfile(userID)
		c.JSON(200, profile)
	})

	// Follow user
	router.POST("/api/users/:user_id/follow", func(c *gin.Context) {
		followerID := c.Query("follower_id")
		followeeID := c.Param("user_id")

		err := sfs.FollowUser(followerID, followeeID)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "success"})
	})

	// Unfollow user
	router.DELETE("/api/users/:user_id/follow", func(c *gin.Context) {
		followerID := c.Query("follower_id")
		followeeID := c.Param("user_id")

		err := sfs.UnfollowUser(followerID, followeeID)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "success"})
	})

	// Leaderboard
	router.GET("/api/leaderboard", func(c *gin.Context) {
		limit := 10
		fmt.Sscanf(c.Query("limit"), "%d", &limit)

		leaderboard := sfs.GetLeaderboard(limit)
		c.JSON(200, gin.H{
			"leaderboard": leaderboard,
		})
	})

	// WebSocket endpoint
	router.GET("/ws", sfs.handleWebSocket)

	return router
}

func main() {
	flag.Parse()

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	sfs := NewSocialFeedService()

	// Start Kafka consumer if configured
	if os.Getenv("ENABLE_KAFKA") == "1" {
		bootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
		apiKey := os.Getenv("KAFKA_API_KEY")
		apiSecret := os.Getenv("KAFKA_API_SECRET")

		if bootstrapServers != "" && apiKey != "" && apiSecret != "" {
			err := sfs.StartKafkaConsumer(bootstrapServers, apiKey, apiSecret)
			if err != nil {
				log.Warnf("Failed to start Kafka consumer: %v", err)
			} else {
				log.Info("Kafka consumer started successfully")
			}
		} else {
			log.Info("Kafka not configured")
		}
	} else {
		log.Info("Kafka disabled")
	}

	// Seed some demo events
	sfs.seedDemoData()

	router := setupRoutes(sfs)

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Infof("Starting Social Feed Service on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	if sfs.kafkaConsumer != nil {
		sfs.kafkaConsumer.Close()
	}

	log.Info("Server exited")
}

// seedDemoData adds some demo events for testing
func (sfs *SocialFeedService) seedDemoData() {
	demoEvents := []FeedEvent{
		{
			EventType: "bet_placed",
			UserID:    "alice123",
			MarketID:  "btc-150k-jun2025",
			Data: map[string]interface{}{
				"side":   "YES",
				"amount": 100.0,
				"shares": 90.91,
			},
			Timestamp: time.Now().Unix() - 3600,
		},
		{
			EventType: "bet_placed",
			UserID:    "bob456",
			MarketID:  "eth-5k-dec2025",
			Data: map[string]interface{}{
				"side":   "NO",
				"amount": 50.0,
				"shares": 45.45,
			},
			Timestamp: time.Now().Unix() - 1800,
		},
		{
			EventType: "win",
			UserID:    "charlie789",
			MarketID:  "trump-president-2025",
			Data: map[string]interface{}{
				"amount": 250.0,
				"pnl":    150.0,
			},
			Timestamp: time.Now().Unix() - 900,
		},
	}

	for _, event := range demoEvents {
		sfs.AddEvent(event)
	}

	log.Info("Seeded demo data")
}
