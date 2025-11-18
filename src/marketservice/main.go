// Copyright 2025 Vibe Markets
// Licensed under the Apache License, Version 2.0

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

var (
	log           *logrus.Logger
	port          = "3550"
	catalogMutex  *sync.RWMutex
	kafkaProducer *KafkaProducer
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
	catalogMutex = &sync.RWMutex{}
}

// Market represents a prediction market
type Market struct {
	ID                 string  `json:"id"`
	Title              string  `json:"title"`
	Description        string  `json:"description"`
	ResolutionCriteria string  `json:"resolution_criteria"`
	EndDate            int64   `json:"end_date"`
	CreatorID          string  `json:"creator_id"`
	Status             string  `json:"status"`
	Outcome            string  `json:"outcome"`
	YesPrice           float64 `json:"yes_price"`
	NoPrice            float64 `json:"no_price"`
	TotalVolume        float64 `json:"total_volume"`
	TotalTraders       int32   `json:"total_traders"`
	Category           string  `json:"category"`
	ImageURL           string  `json:"image_url"`
	CreatedAt          int64   `json:"created_at"`
}

type marketService struct {
	markets []Market
}

func (m *marketService) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (m *marketService) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

func (m *marketService) ListMarkets(category, statusFilter string, limit, offset int32) ([]Market, int32, error) {
	catalogMutex.RLock()
	defer catalogMutex.RUnlock()

	var filtered []Market
	for _, market := range m.markets {
		// Apply filters
		if category != "" && market.Category != category {
			continue
		}
		if statusFilter != "" && market.Status != statusFilter {
			continue
		}
		filtered = append(filtered, market)
	}

	total := int32(len(filtered))

	// Apply pagination
	if limit == 0 {
		limit = 50
	}
	start := offset
	end := offset + limit
	if start > int32(len(filtered)) {
		return []Market{}, total, nil
	}
	if end > int32(len(filtered)) {
		end = int32(len(filtered))
	}

	return filtered[start:end], total, nil
}

func (m *marketService) GetMarket(marketID string) (*Market, error) {
	catalogMutex.RLock()
	defer catalogMutex.RUnlock()

	for _, market := range m.markets {
		if market.ID == marketID {
			return &market, nil
		}
	}
	return nil, fmt.Errorf("market not found: %s", marketID)
}

func (m *marketService) SearchMarkets(query, category string) ([]Market, error) {
	catalogMutex.RLock()
	defer catalogMutex.RUnlock()

	var results []Market
	queryLower := query

	for _, market := range m.markets {
		// Check category filter
		if category != "" && market.Category != category {
			continue
		}

		// Simple search on title and description
		if query == "" ||
			contains(market.Title, queryLower) ||
			contains(market.Description, queryLower) {
			results = append(results, market)
		}
	}

	return results, nil
}

func (m *marketService) UpdateMarketPrice(marketID string, yesPrice, noPrice, volume float64) error {
	catalogMutex.Lock()
	defer catalogMutex.Unlock()

	for i, market := range m.markets {
		if market.ID == marketID {
			m.markets[i].YesPrice = yesPrice
			m.markets[i].NoPrice = noPrice
			m.markets[i].TotalVolume += volume

			// Publish to Kafka
			if kafkaProducer != nil {
				err := kafkaProducer.PublishMarketUpdate(marketID, yesPrice, noPrice, m.markets[i].TotalVolume)
				if err != nil {
					log.Errorf("Failed to publish market update to Kafka: %v", err)
				}
			}

			log.Infof("Updated market %s: YES=%.2f, NO=%.2f, Volume=%.2f", marketID, yesPrice, noPrice, m.markets[i].TotalVolume)
			return nil
		}
	}
	return fmt.Errorf("market not found: %s", marketID)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0)
}

func loadMarkets(markets *[]Market) error {
	file, err := os.ReadFile("data/demo_markets.json")
	if err != nil {
		return fmt.Errorf("failed to read markets file: %w", err)
	}

	err = json.Unmarshal(file, markets)
	if err != nil {
		return fmt.Errorf("failed to parse markets JSON: %w", err)
	}

	log.Infof("Loaded %d markets from catalog", len(*markets))
	return nil
}

func main() {
	flag.Parse()

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	// Initialize Kafka producer if configured
	if os.Getenv("ENABLE_KAFKA") == "1" {
		bootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
		apiKey := os.Getenv("KAFKA_API_KEY")
		apiSecret := os.Getenv("KAFKA_API_SECRET")

		if bootstrapServers != "" && apiKey != "" && apiSecret != "" {
			var err error
			kafkaProducer, err = NewKafkaProducer(bootstrapServers, apiKey, apiSecret)
			if err != nil {
				log.Warnf("Failed to initialize Kafka producer: %v", err)
			} else {
				log.Info("Kafka producer initialized successfully")
			}
		} else {
			log.Info("Kafka not configured, running without event streaming")
		}
	} else {
		log.Info("Kafka disabled via ENABLE_KAFKA flag")
	}

	// Handle graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Printf("Received signal: %s, shutting down gracefully", sig)
		if kafkaProducer != nil {
			kafkaProducer.Close()
		}
		os.Exit(0)
	}()

	log.Infof("starting Market Service gRPC server at :%s", port)
	run(port)
	select {}
}

func run(port string) string {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer()

	svc := &marketService{}
	err = loadMarkets(&svc.markets)
	if err != nil {
		log.Fatalf("could not load market catalog: %v", err)
	}

	// Note: gRPC registration would go here once we generate proto code
	// For now, this service can be called directly via Go
	healthpb.RegisterHealthServer(srv, svc)

	go srv.Serve(listener)

	return listener.Addr().String()
}
