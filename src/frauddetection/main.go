// Copyright 2025 Vibe Markets
// Licensed under the Apache License, Version 2.0

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

var (
	log *logrus.Logger
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

// BetEvent represents a betting event
type BetEvent struct {
	UserID    string  `json:"user_id"`
	MarketID  string  `json:"market_id"`
	Side      string  `json:"side"`
	Amount    float64 `json:"amount"`
	Timestamp int64   `json:"timestamp"`
}

// FraudAlert represents a detected fraud pattern
type FraudAlert struct {
	AlertID     string    `json:"alert_id"`
	AlertType   string    `json:"alert_type"`
	UserID      string    `json:"user_id"`
	MarketID    string    `json:"market_id"`
	Severity    string    `json:"severity"` // "low", "medium", "high"
	Description string    `json:"description"`
	Evidence    map[string]interface{} `json:"evidence"`
	Timestamp   time.Time `json:"timestamp"`
}

// UserActivity tracks betting activity per user
type UserActivity struct {
	UserID       string
	Bets         []BetEvent
	LastBetTime  time.Time
	TotalBets    int
	TotalVolume  float64
	Markets      map[string]int
	RapidBets    int
	Alternations int
}

// FraudDetector monitors betting patterns for suspicious activity
type FraudDetector struct {
	userActivity    map[string]*UserActivity
	mu              sync.RWMutex
	alertProducer   *kafka.Producer
	consumer        *kafka.Consumer
	fraudAlerts     []FraudAlert
	alertThresholds FraudThresholds
}

// FraudThresholds defines detection thresholds
type FraudThresholds struct {
	RapidBetsWindow      time.Duration // Time window for rapid bets
	RapidBetsCount       int           // Number of bets to trigger alert
	LargeBetAmount       float64       // Amount that triggers large bet alert
	WashTradingWindow    time.Duration // Time window for wash trading
	WashTradingMinBets   int           // Min alternating bets
	UnusualVolumeMultiplier float64    // Multiplier for unusual volume
}

// NewFraudDetector creates a new fraud detection service
func NewFraudDetector(producer *kafka.Producer) *FraudDetector {
	return &FraudDetector{
		userActivity:  make(map[string]*UserActivity),
		alertProducer: producer,
		fraudAlerts:   make([]FraudAlert, 0),
		alertThresholds: FraudThresholds{
			RapidBetsWindow:         1 * time.Minute,
			RapidBetsCount:          10,
			LargeBetAmount:          1000.0,
			WashTradingWindow:       5 * time.Minute,
			WashTradingMinBets:      6,
			UnusualVolumeMultiplier: 5.0,
		},
	}
}

// ProcessBet analyzes a bet for fraud patterns
func (fd *FraudDetector) ProcessBet(bet BetEvent) {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	// Get or create user activity
	activity := fd.userActivity[bet.UserID]
	if activity == nil {
		activity = &UserActivity{
			UserID:  bet.UserID,
			Bets:    make([]BetEvent, 0),
			Markets: make(map[string]int),
		}
		fd.userActivity[bet.UserID] = activity
	}

	// Update activity
	activity.Bets = append(activity.Bets, bet)
	activity.LastBetTime = time.Unix(bet.Timestamp, 0)
	activity.TotalBets++
	activity.TotalVolume += bet.Amount
	activity.Markets[bet.MarketID]++

	// Run fraud checks
	fd.checkRapidBetting(activity, bet)
	fd.checkLargeBet(bet)
	fd.checkWashTrading(activity, bet)
	fd.checkBotBehavior(activity, bet)

	// Clean old data (keep only last 1 hour)
	fd.cleanOldActivity(activity)
}

// checkRapidBetting detects users placing many bets in short time
func (fd *FraudDetector) checkRapidBetting(activity *UserActivity, bet BetEvent) {
	now := time.Unix(bet.Timestamp, 0)
	windowStart := now.Add(-fd.alertThresholds.RapidBetsWindow)

	// Count recent bets
	recentBets := 0
	for _, b := range activity.Bets {
		if time.Unix(b.Timestamp, 0).After(windowStart) {
			recentBets++
		}
	}

	if recentBets >= fd.alertThresholds.RapidBetsCount {
		alert := FraudAlert{
			AlertID:     fmt.Sprintf("rapid-%s-%d", bet.UserID, time.Now().Unix()),
			AlertType:   "rapid_betting",
			UserID:      bet.UserID,
			MarketID:    bet.MarketID,
			Severity:    "medium",
			Description: fmt.Sprintf("User placed %d bets in %v", recentBets, fd.alertThresholds.RapidBetsWindow),
			Evidence: map[string]interface{}{
				"bet_count":     recentBets,
				"time_window":   fd.alertThresholds.RapidBetsWindow.String(),
				"latest_amount": bet.Amount,
			},
			Timestamp: now,
		}
		fd.raiseAlert(alert)
	}
}

// checkLargeBet detects unusually large bets
func (fd *FraudDetector) checkLargeBet(bet BetEvent) {
	if bet.Amount >= fd.alertThresholds.LargeBetAmount {
		alert := FraudAlert{
			AlertID:     fmt.Sprintf("large-%s-%d", bet.UserID, time.Now().Unix()),
			AlertType:   "large_bet",
			UserID:      bet.UserID,
			MarketID:    bet.MarketID,
			Severity:    "low",
			Description: fmt.Sprintf("Large bet of $%.2f detected", bet.Amount),
			Evidence: map[string]interface{}{
				"amount": bet.Amount,
				"side":   bet.Side,
			},
			Timestamp: time.Unix(bet.Timestamp, 0),
		}
		fd.raiseAlert(alert)
	}
}

// checkWashTrading detects alternating YES/NO bets (potential wash trading)
func (fd *FraudDetector) checkWashTrading(activity *UserActivity, bet BetEvent) {
	now := time.Unix(bet.Timestamp, 0)
	windowStart := now.Add(-fd.alertThresholds.WashTradingWindow)

	// Get recent bets on this market
	var recentBets []BetEvent
	for _, b := range activity.Bets {
		if b.MarketID == bet.MarketID && time.Unix(b.Timestamp, 0).After(windowStart) {
			recentBets = append(recentBets, b)
		}
	}

	if len(recentBets) < fd.alertThresholds.WashTradingMinBets {
		return
	}

	// Count alternations
	alternations := 0
	for i := 1; i < len(recentBets); i++ {
		if recentBets[i].Side != recentBets[i-1].Side {
			alternations++
		}
	}

	// If most bets alternate sides, suspicious
	if alternations >= len(recentBets)/2 {
		alert := FraudAlert{
			AlertID:     fmt.Sprintf("wash-%s-%s-%d", bet.UserID, bet.MarketID, time.Now().Unix()),
			AlertType:   "wash_trading",
			UserID:      bet.UserID,
			MarketID:    bet.MarketID,
			Severity:    "high",
			Description: fmt.Sprintf("Suspected wash trading: %d alternating bets", len(recentBets)),
			Evidence: map[string]interface{}{
				"total_bets":    len(recentBets),
				"alternations":  alternations,
				"time_window":   fd.alertThresholds.WashTradingWindow.String(),
			},
			Timestamp: now,
		}
		fd.raiseAlert(alert)
	}
}

// checkBotBehavior detects bot-like betting patterns
func (fd *FraudDetector) checkBotBehavior(activity *UserActivity, bet BetEvent) {
	if len(activity.Bets) < 10 {
		return
	}

	// Check for perfectly timed bets (every X seconds)
	recentBets := activity.Bets[len(activity.Bets)-10:]
	intervals := make([]int64, 0)

	for i := 1; i < len(recentBets); i++ {
		interval := recentBets[i].Timestamp - recentBets[i-1].Timestamp
		intervals = append(intervals, interval)
	}

	// Check if all intervals are very similar (bot-like)
	if len(intervals) > 0 {
		avg := averageInt64(intervals)
		variance := varianceInt64(intervals, avg)

		// Low variance = very consistent timing = bot-like
		if variance < 5 { // seconds
			alert := FraudAlert{
				AlertID:     fmt.Sprintf("bot-%s-%d", bet.UserID, time.Now().Unix()),
				AlertType:   "bot_behavior",
				UserID:      bet.UserID,
				MarketID:    bet.MarketID,
				Severity:    "medium",
				Description: "Bot-like betting pattern detected (consistent timing)",
				Evidence: map[string]interface{}{
					"avg_interval":  avg,
					"variance":      variance,
					"recent_bets":   len(recentBets),
				},
				Timestamp: time.Unix(bet.Timestamp, 0),
			}
			fd.raiseAlert(alert)
		}
	}
}

// raiseAlert publishes an alert
func (fd *FraudDetector) raiseAlert(alert FraudAlert) {
	fd.fraudAlerts = append(fd.fraudAlerts, alert)

	// Keep only last 1000 alerts
	if len(fd.fraudAlerts) > 1000 {
		fd.fraudAlerts = fd.fraudAlerts[len(fd.fraudAlerts)-1000:]
	}

	// Log alert
	log.Warnf("FRAUD ALERT [%s]: %s - User: %s, Market: %s",
		alert.Severity, alert.AlertType, alert.UserID, alert.MarketID)

	// Publish to Kafka (if producer available)
	if fd.alertProducer != nil {
		data, _ := json.Marshal(alert)
		topic := "fraud-alerts"
		fd.alertProducer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          data,
		}, nil)
	}
}

// cleanOldActivity removes old betting data
func (fd *FraudDetector) cleanOldActivity(activity *UserActivity) {
	cutoff := time.Now().Add(-1 * time.Hour)

	var recent []BetEvent
	for _, bet := range activity.Bets {
		if time.Unix(bet.Timestamp, 0).After(cutoff) {
			recent = append(recent, bet)
		}
	}
	activity.Bets = recent
}

// Helper functions
func averageInt64(nums []int64) float64 {
	if len(nums) == 0 {
		return 0
	}
	sum := int64(0)
	for _, n := range nums {
		sum += n
	}
	return float64(sum) / float64(len(nums))
}

func varianceInt64(nums []int64, avg float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	sumSq := 0.0
	for _, n := range nums {
		diff := float64(n) - avg
		sumSq += diff * diff
	}
	return sumSq / float64(len(nums))
}

// StartKafkaConsumer starts consuming bet-orders
func (fd *FraudDetector) StartKafkaConsumer(bootstrapServers, apiKey, apiSecret string) error {
	config := &kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"security.protocol": "SASL_SSL",
		"sasl.mechanisms":   "PLAIN",
		"sasl.username":     apiKey,
		"sasl.password":     apiSecret,
		"group.id":          "fraud-detection",
		"auto.offset.reset": "earliest",
	}

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	fd.consumer = consumer

	err = consumer.SubscribeTopics([]string{"bet-orders"}, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	log.Info("Kafka consumer started, monitoring bet-orders for fraud")

	// Consume in background
	go func() {
		for {
			msg, err := consumer.ReadMessage(-1)
			if err != nil {
				log.Errorf("Consumer error: %v", err)
				continue
			}

			var bet BetEvent
			err = json.Unmarshal(msg.Value, &bet)
			if err != nil {
				log.Errorf("Failed to unmarshal bet: %v", err)
				continue
			}

			fd.ProcessBet(bet)
		}
	}()

	return nil
}

// GetAlerts returns recent fraud alerts
func (fd *FraudDetector) GetAlerts(limit int) []FraudAlert {
	fd.mu.RLock()
	defer fd.mu.RUnlock()

	if limit == 0 || limit > len(fd.fraudAlerts) {
		limit = len(fd.fraudAlerts)
	}

	// Return most recent alerts
	start := len(fd.fraudAlerts) - limit
	if start < 0 {
		start = 0
	}

	return fd.fraudAlerts[start:]
}

// GetUserActivity returns activity for a specific user
func (fd *FraudDetector) GetUserActivity(userID string) *UserActivity {
	fd.mu.RLock()
	defer fd.mu.RUnlock()

	return fd.userActivity[userID]
}

func main() {
	flag.Parse()

	// Create Kafka producer for alerts
	var producer *kafka.Producer

	if os.Getenv("ENABLE_KAFKA") == "1" {
		bootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
		apiKey := os.Getenv("KAFKA_API_KEY")
		apiSecret := os.Getenv("KAFKA_API_SECRET")

		if bootstrapServers != "" && apiKey != "" && apiSecret != "" {
			config := &kafka.ConfigMap{
				"bootstrap.servers": bootstrapServers,
				"security.protocol": "SASL_SSL",
				"sasl.mechanisms":   "PLAIN",
				"sasl.username":     apiKey,
				"sasl.password":     apiSecret,
			}

			var err error
			producer, err = kafka.NewProducer(config)
			if err != nil {
				log.Warnf("Failed to create producer: %v", err)
			} else {
				log.Info("Kafka producer initialized")
			}
		}
	}

	fd := NewFraudDetector(producer)

	// Start consumer if Kafka enabled
	if os.Getenv("ENABLE_KAFKA") == "1" {
		bootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
		apiKey := os.Getenv("KAFKA_API_KEY")
		apiSecret := os.Getenv("KAFKA_API_SECRET")

		if bootstrapServers != "" && apiKey != "" && apiSecret != "" {
			err := fd.StartKafkaConsumer(bootstrapServers, apiKey, apiSecret)
			if err != nil {
				log.Fatalf("Failed to start consumer: %v", err)
			}
		}
	}

	log.Info("Fraud Detection Service started")

	// Wait for interrupt signal
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	<-sigchan

	log.Info("Shutting down fraud detection service...")

	if fd.consumer != nil {
		fd.consumer.Close()
	}
	if producer != nil {
		producer.Flush(5000)
		producer.Close()
	}

	log.Info("Shutdown complete")
}
