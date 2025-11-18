// Copyright 2025 Vibe Markets
// Licensed under the Apache License, Version 2.0

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var (
	log  *logrus.Logger
	port = "9000"

	// Service URLs
	marketServiceURL     string
	bettingServiceURL    string
	socialFeedServiceURL string
	aiServiceURL         string

	wsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
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

	// Load service URLs from environment
	marketServiceURL = getEnv("MARKET_SERVICE_URL", "http://localhost:3550")
	bettingServiceURL = getEnv("BETTING_SERVICE_URL", "http://localhost:3551")
	socialFeedServiceURL = getEnv("SOCIAL_FEED_SERVICE_URL", "http://localhost:8080")
	aiServiceURL = getEnv("AI_SERVICE_URL", "http://localhost:8081")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// APIGateway aggregates all microservices
type APIGateway struct {
	httpClient *http.Client
}

// NewAPIGateway creates a new API gateway
func NewAPIGateway() *APIGateway {
	return &APIGateway{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Forward a request to a service
func (gw *APIGateway) forwardRequest(method, serviceURL, path string, body io.Reader, query url.Values) ([]byte, int, error) {
	// Build full URL
	fullURL := serviceURL + path
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, 500, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := gw.httpClient.Do(req)
	if err != nil {
		return nil, 503, fmt.Errorf("service unavailable: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 500, err
	}

	return responseBody, resp.StatusCode, nil
}

// Setup routes
func setupRoutes(gw *APIGateway) *gin.Engine {
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
		c.JSON(200, gin.H{
			"status":   "healthy",
			"service":  "api-gateway",
			"services": gin.H{
				"market":      marketServiceURL,
				"betting":     bettingServiceURL,
				"social_feed": socialFeedServiceURL,
				"ai":          aiServiceURL,
			},
		})
	})

	// ================== MARKET SERVICE ROUTES ==================
	markets := router.Group("/api/markets")
	{
		// List markets
		markets.GET("", func(c *gin.Context) {
			// For now, return demo markets (Week 1 markets service would be called here)
			markets := []map[string]interface{}{
				{
					"id": "btc-150k-jun2025", "title": "Will Bitcoin reach $150,000 by June 2025?",
					"category": "crypto", "yes_price": 0.48, "no_price": 0.52,
					"total_volume": 12500.0, "total_traders": 234, "status": "open",
				},
				{
					"id": "eth-5k-dec2025", "title": "Will Ethereum reach $5,000 by end of 2025?",
					"category": "crypto", "yes_price": 0.62, "no_price": 0.38,
					"total_volume": 8900.0, "total_traders": 156, "status": "open",
				},
				{
					"id": "superbowl-chiefs-2026", "title": "Will Kansas City Chiefs win Super Bowl LX?",
					"category": "sports", "yes_price": 0.35, "no_price": 0.65,
					"total_volume": 15200.0, "total_traders": 412, "status": "open",
				},
				{
					"id": "trump-president-2025", "title": "Will Donald Trump be inaugurated in January 2025?",
					"category": "politics", "yes_price": 0.89, "no_price": 0.11,
					"total_volume": 45300.0, "total_traders": 1823, "status": "open",
				},
				{
					"id": "ai-agi-2025", "title": "Will OpenAI announce AGI in 2025?",
					"category": "tech", "yes_price": 0.15, "no_price": 0.85,
					"total_volume": 6700.0, "total_traders": 289, "status": "open",
				},
				{
					"id": "spacex-mars-2025", "title": "Will SpaceX launch Starship to Mars in 2025?",
					"category": "tech", "yes_price": 0.28, "no_price": 0.72,
					"total_volume": 9400.0, "total_traders": 234, "status": "open",
				},
				{
					"id": "gta6-release-2025", "title": "Will GTA 6 release in 2025?",
					"category": "entertainment", "yes_price": 0.72, "no_price": 0.28,
					"total_volume": 11200.0, "total_traders": 567, "status": "open",
				},
				{
					"id": "stock-spy-500-2025", "title": "Will SPY close above $500 in 2025?",
					"category": "crypto", "yes_price": 0.81, "no_price": 0.19,
					"total_volume": 23400.0, "total_traders": 892, "status": "open",
				},
				{
					"id": "arsenal-epl-2025", "title": "Will Arsenal win Premier League 2024-25?",
					"category": "sports", "yes_price": 0.42, "no_price": 0.58,
					"total_volume": 7800.0, "total_traders": 312, "status": "open",
				},
				{
					"id": "solana-200-2025", "title": "Will Solana reach $200 in 2025?",
					"category": "crypto", "yes_price": 0.56, "no_price": 0.44,
					"total_volume": 5600.0, "total_traders": 178, "status": "open",
				},
			}

			c.JSON(200, gin.H{
				"markets": markets,
				"total":   len(markets),
			})
		})

		// Get market by ID
		markets.GET("/:market_id", func(c *gin.Context) {
			// Demo response
			c.JSON(200, gin.H{
				"id": c.Param("market_id"),
				"title": "Will Bitcoin reach $150,000 by June 2025?",
				"description": "This market resolves YES if Bitcoin trades at $150k+ before June 30, 2025.",
				"yes_price": 0.48,
				"no_price": 0.52,
				"total_volume": 12500.0,
				"status": "open",
			})
		})
	}

	// ================== BETTING SERVICE ROUTES ==================
	betting := router.Group("/api/betting")
	{
		// Place a bet
		betting.POST("/bet", func(c *gin.Context) {
			var req map[string]interface{}
			if err := c.BindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": "invalid request"})
				return
			}

			// Demo response (would call betting service)
			c.JSON(200, gin.H{
				"bet_id":          "bet-123",
				"shares_received": 90.91,
				"average_price":   1.10,
				"new_market_price": 0.52,
				"status":          "success",
			})
		})

		// Get user positions
		betting.GET("/positions/:user_id", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"user_id": c.Param("user_id"),
				"positions": []map[string]interface{}{
					{
						"market_id": "btc-150k-jun2025",
						"side": "YES",
						"shares": 90.91,
						"average_cost": 1.10,
						"current_value": 95.00,
						"pnl": 5.00,
					},
				},
				"total_value": 95.00,
				"total_pnl": 5.00,
			})
		})

		// Get user balance
		betting.GET("/balance/:user_id", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"user_id": c.Param("user_id"),
				"balance": 1000.00,
				"invested": 100.00,
				"available": 900.00,
			})
		})
	}

	// ================== SOCIAL FEED ROUTES ==================
	social := router.Group("/api/social")
	{
		// Get feed
		social.GET("/feed", func(c *gin.Context) {
			resp, status, err := gw.forwardRequest("GET", socialFeedServiceURL, "/api/feed", nil, c.Request.URL.Query())
			if err != nil {
				c.JSON(status, gin.H{"error": err.Error()})
				return
			}

			var data map[string]interface{}
			json.Unmarshal(resp, &data)
			c.JSON(status, data)
		})

		// Get user profile
		social.GET("/users/:user_id", func(c *gin.Context) {
			userID := c.Param("user_id")
			resp, status, err := gw.forwardRequest("GET", socialFeedServiceURL, "/api/users/"+userID, nil, nil)
			if err != nil {
				c.JSON(status, gin.H{"error": err.Error()})
				return
			}

			var data map[string]interface{}
			json.Unmarshal(resp, &data)
			c.JSON(status, data)
		})

		// Follow user
		social.POST("/users/:user_id/follow", func(c *gin.Context) {
			userID := c.Param("user_id")
			resp, status, err := gw.forwardRequest("POST", socialFeedServiceURL, "/api/users/"+userID+"/follow", nil, c.Request.URL.Query())
			if err != nil {
				c.JSON(status, gin.H{"error": err.Error()})
				return
			}

			var data map[string]interface{}
			json.Unmarshal(resp, &data)
			c.JSON(status, data)
		})

		// Leaderboard
		social.GET("/leaderboard", func(c *gin.Context) {
			resp, status, err := gw.forwardRequest("GET", socialFeedServiceURL, "/api/leaderboard", nil, c.Request.URL.Query())
			if err != nil {
				c.JSON(status, gin.H{"error": err.Error()})
				return
			}

			var data map[string]interface{}
			json.Unmarshal(resp, &data)
			c.JSON(status, data)
		})
	}

	// ================== AI SERVICE ROUTES ==================
	ai := router.Group("/api/ai")
	{
		// Get recommendations
		ai.GET("/recommendations/:user_id", func(c *gin.Context) {
			userID := c.Param("user_id")
			resp, status, err := gw.forwardRequest("GET", aiServiceURL, "/api/recommendations/"+userID, nil, c.Request.URL.Query())
			if err != nil {
				c.JSON(status, gin.H{"error": err.Error()})
				return
			}

			var data map[string]interface{}
			json.Unmarshal(resp, &data)
			c.JSON(status, data)
		})

		// Get market summary
		ai.GET("/summary/:market_id", func(c *gin.Context) {
			marketID := c.Param("market_id")
			resp, status, err := gw.forwardRequest("GET", aiServiceURL, "/api/summary/"+marketID, nil, nil)
			if err != nil {
				c.JSON(status, gin.H{"error": err.Error()})
				return
			}

			var data map[string]interface{}
			json.Unmarshal(resp, &data)
			c.JSON(status, data)
		})

		// Track activity
		ai.POST("/activity", func(c *gin.Context) {
			body, _ := io.ReadAll(c.Request.Body)
			resp, status, err := gw.forwardRequest("POST", aiServiceURL, "/api/activity", bytes.NewReader(body), nil)
			if err != nil {
				c.JSON(status, gin.H{"error": err.Error()})
				return
			}

			var data map[string]interface{}
			json.Unmarshal(resp, &data)
			c.JSON(status, data)
		})
	}

	// ================== WEBSOCKET PROXY ==================
	router.GET("/ws", func(c *gin.Context) {
		// Proxy WebSocket to social feed service
		proxyWebSocket(c.Writer, c.Request, socialFeedServiceURL+"/ws")
	})

	// ================== AGGREGATED ENDPOINTS ==================
	// Get market with AI summary
	router.GET("/api/market-detail/:market_id", func(c *gin.Context) {
		marketID := c.Param("market_id")

		// Get market data (demo)
		market := map[string]interface{}{
			"id": marketID,
			"title": "Will Bitcoin reach $150,000 by June 2025?",
			"yes_price": 0.48,
			"no_price": 0.52,
			"total_volume": 12500.0,
		}

		// Get AI summary
		summaryResp, _, _ := gw.forwardRequest("GET", aiServiceURL, "/api/summary/"+marketID, nil, nil)
		var summary map[string]interface{}
		json.Unmarshal(summaryResp, &summary)

		c.JSON(200, gin.H{
			"market": market,
			"ai_summary": summary,
		})
	})

	// Get personalized homepage
	router.GET("/api/homepage/:user_id", func(c *gin.Context) {
		userID := c.Param("user_id")

		// Get recommendations
		recResp, _, _ := gw.forwardRequest("GET", aiServiceURL, "/api/recommendations/"+userID+"?limit=5", nil, nil)
		var recommendations map[string]interface{}
		json.Unmarshal(recResp, &recommendations)

		// Get recent feed
		feedResp, _, _ := gw.forwardRequest("GET", socialFeedServiceURL, "/api/feed?user_id="+userID+"&limit=10", nil, nil)
		var feed map[string]interface{}
		json.Unmarshal(feedResp, &feed)

		// Get leaderboard
		leaderboardResp, _, _ := gw.forwardRequest("GET", socialFeedServiceURL, "/api/leaderboard?limit=5", nil, nil)
		var leaderboard map[string]interface{}
		json.Unmarshal(leaderboardResp, &leaderboard)

		c.JSON(200, gin.H{
			"user_id": userID,
			"recommendations": recommendations,
			"recent_activity": feed,
			"leaderboard": leaderboard,
		})
	})

	return router
}

// proxyWebSocket proxies WebSocket connections
func proxyWebSocket(w http.ResponseWriter, r *http.Request, targetURL string) {
	// Upgrade client connection
	clientConn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("WebSocket upgrade error: %v", err)
		return
	}
	defer clientConn.Close()

	// Connect to backend
	backendConn, _, err := websocket.DefaultDialer.Dial(targetURL, nil)
	if err != nil {
		log.Errorf("WebSocket backend connection error: %v", err)
		return
	}
	defer backendConn.Close()

	// Proxy messages
	go func() {
		for {
			_, msg, err := backendConn.ReadMessage()
			if err != nil {
				return
			}
			clientConn.WriteMessage(websocket.TextMessage, msg)
		}
	}()

	for {
		_, msg, err := clientConn.ReadMessage()
		if err != nil {
			return
		}
		backendConn.WriteMessage(websocket.TextMessage, msg)
	}
}

func main() {
	flag.Parse()

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	gw := NewAPIGateway()
	router := setupRoutes(gw)

	log.Infof("Starting API Gateway on port %s", port)
	log.Infof("Forwarding to services:")
	log.Infof("  - Market: %s", marketServiceURL)
	log.Infof("  - Betting: %s", bettingServiceURL)
	log.Infof("  - Social Feed: %s", socialFeedServiceURL)
	log.Infof("  - AI: %s", aiServiceURL)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start gateway: %v", err)
	}
}
