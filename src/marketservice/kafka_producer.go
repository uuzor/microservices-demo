// Copyright 2025 Vibe Markets
// Licensed under the Apache License, Version 2.0

package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// KafkaProducer wraps the Confluent Kafka producer
type KafkaProducer struct {
	producer *kafka.Producer
}

// MarketUpdateEvent represents a market price/volume update
type MarketUpdateEvent struct {
	MarketID  string  `json:"market_id"`
	YesPrice  float64 `json:"yes_price"`
	NoPrice   float64 `json:"no_price"`
	Volume    float64 `json:"volume"`
	Timestamp int64   `json:"timestamp"`
	EventType string  `json:"event_type"`
}

// BetOrderEvent represents a new bet order
type BetOrderEvent struct {
	UserID    string  `json:"user_id"`
	MarketID  string  `json:"market_id"`
	Side      string  `json:"side"`
	Amount    float64 `json:"amount"`
	Timestamp int64   `json:"timestamp"`
	EventType string  `json:"event_type"`
}

// SocialFeedEvent represents an action for the social feed
type SocialFeedEvent struct {
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id"`
	MarketID  string                 `json:"market_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

// NewKafkaProducer creates a new Kafka producer for Confluent Cloud
func NewKafkaProducer(bootstrapServers, apiKey, apiSecret string) (*KafkaProducer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"security.protocol": "SASL_SSL",
		"sasl.mechanisms":   "PLAIN",
		"sasl.username":     apiKey,
		"sasl.password":     apiSecret,
		"acks":              "all",
		"retries":           3,
		"client.id":         "vibe-markets-producer",
	}

	p, err := kafka.NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	// Handle delivery reports in background
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Errorf("Delivery failed: %v", ev.TopicPartition.Error)
				} else {
					log.Debugf("Delivered message to %v", ev.TopicPartition)
				}
			}
		}
	}()

	return &KafkaProducer{producer: p}, nil
}

// PublishMarketUpdate publishes a market price/volume update to Kafka
func (kp *KafkaProducer) PublishMarketUpdate(marketID string, yesPrice, noPrice, volume float64) error {
	event := MarketUpdateEvent{
		MarketID:  marketID,
		YesPrice:  yesPrice,
		NoPrice:   noPrice,
		Volume:    volume,
		Timestamp: time.Now().Unix(),
		EventType: "market_update",
	}

	return kp.publishEvent("market-updates", marketID, event)
}

// PublishBetOrder publishes a bet order to Kafka
func (kp *KafkaProducer) PublishBetOrder(userID, marketID, side string, amount float64) error {
	event := BetOrderEvent{
		UserID:    userID,
		MarketID:  marketID,
		Side:      side,
		Amount:    amount,
		Timestamp: time.Now().Unix(),
		EventType: "bet_placed",
	}

	return kp.publishEvent("bet-orders", marketID, event)
}

// PublishSocialFeedEvent publishes a social feed event
func (kp *KafkaProducer) PublishSocialFeedEvent(eventType, userID, marketID string, data map[string]interface{}) error {
	event := SocialFeedEvent{
		EventType: eventType,
		UserID:    userID,
		MarketID:  marketID,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	return kp.publishEvent("social-feed-events", userID, event)
}

// PublishUserActivity publishes user activity for ML training
func (kp *KafkaProducer) PublishUserActivity(userID, action, marketID string, context map[string]interface{}) error {
	event := map[string]interface{}{
		"user_id":   userID,
		"action":    action,
		"market_id": marketID,
		"context":   context,
		"timestamp": time.Now().Unix(),
	}

	return kp.publishEvent("user-activity", userID, event)
}

// publishEvent is a helper to publish any event to Kafka
func (kp *KafkaProducer) publishEvent(topic string, key string, event interface{}) error {
	if kp.producer == nil {
		return fmt.Errorf("kafka producer not initialized")
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = kp.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(key),
		Value: data,
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

// Flush waits for all messages to be delivered
func (kp *KafkaProducer) Flush(timeoutMs int) int {
	if kp.producer == nil {
		return 0
	}
	return kp.producer.Flush(timeoutMs)
}

// Close closes the producer
func (kp *KafkaProducer) Close() {
	if kp.producer != nil {
		kp.producer.Flush(5000)
		kp.producer.Close()
		log.Info("Kafka producer closed")
	}
}
