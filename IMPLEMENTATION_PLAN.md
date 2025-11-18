# Vibe Markets - Implementation Plan

## Current Status

✅ Created hackathon branch: `hackathon-vibe-markets`
✅ Architecture document created: `HACKATHON_VIBE_MARKETS.md`
✅ Repository structure defined

## Next 7 Days - Critical Path

### Day 1-2: Environment Setup & Confluent Integration

**Priority: CRITICAL**

#### 1. Set Up Confluent Cloud Account
```bash
# Go to: https://confluent.cloud/signup
# Create free trial account
# Create a new cluster (Basic tier is free for trial)
# Note down:
# - Bootstrap servers
# - API Key
# - API Secret
```

#### 2. Set Up Google Cloud Account
```bash
# Go to: https://cloud.google.com/
# Activate $300 free credits
# Enable APIs:
gcloud services enable aiplatform.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable cloudbuild.googleapis.com
```

#### 3. Create Environment Configuration
```bash
# Create .env file
cp .env.example .env

# Add these variables:
CONFLUENT_BOOTSTRAP_SERVERS=pkc-xxx.us-east-1.aws.confluent.cloud:9092
CONFLUENT_API_KEY=your_api_key
CONFLUENT_API_SECRET=your_api_secret
GOOGLE_CLOUD_PROJECT=your_project_id
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
```

#### 4. Create Kafka Topics in Confluent Cloud
```bash
# Using Confluent Cloud Console or CLI:
# Topic 1: bet-orders (partitions: 3, retention: 7 days)
# Topic 2: market-updates (partitions: 3, retention: 7 days)
# Topic 3: social-feed-events (partitions: 3, retention: 7 days)
# Topic 4: user-activity (partitions: 3, retention: 30 days)
# Topic 5: resolution-events (partitions: 1, retention: 7 days)
```

### Day 3-4: Core Market Service

**Priority: HIGH**

#### 1. Create Proto Definitions

Create `protos/markets.proto`:
```protobuf
syntax = "proto3";

package vibemarket;

option go_package = "github.com/youruser/vibe-markets/genproto";

// Market Service
service MarketService {
    rpc ListMarkets(Empty) returns (ListMarketsResponse) {}
    rpc GetMarket(GetMarketRequest) returns (Market) {}
    rpc SearchMarkets(SearchMarketsRequest) returns (SearchMarketsResponse) {}
}

message Market {
    string id = 1;
    string title = 2;
    string description = 3;
    string resolution_criteria = 4;
    int64 end_date = 5;  // Unix timestamp
    string status = 6;  // "open", "closed", "resolved"
    double yes_price = 7;  // 0.0 to 1.0
    double no_price = 8;   // 0.0 to 1.0
    double total_volume = 9;
    string category = 10;  // "crypto", "sports", "politics"
    string created_at = 11;
}

message ListMarketsResponse {
    repeated Market markets = 1;
}

message GetMarketRequest {
    string id = 1;
}

message SearchMarketsRequest {
    string query = 1;
    string category = 2;
}

message SearchMarketsResponse {
    repeated Market results = 1;
}

message Empty {}
```

#### 2. Generate Go Code
```bash
cd protos
protoc --go_out=../src/marketservice/genproto \
       --go-grpc_out=../src/marketservice/genproto \
       markets.proto
```

#### 3. Implement Market Service

Create `src/marketservice/markets.go`:
```go
package main

import (
    "context"
    "time"
    pb "yourrepo/src/marketservice/genproto"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type marketService struct {
    pb.UnimplementedMarketServiceServer
    markets []*pb.Market
}

func (m *marketService) ListMarkets(ctx context.Context, req *pb.Empty) (*pb.ListMarketsResponse, error) {
    return &pb.ListMarketsResponse{Markets: m.markets}, nil
}

func (m *marketService) GetMarket(ctx context.Context, req *pb.GetMarketRequest) (*pb.Market, error) {
    for _, market := range m.markets {
        if market.Id == req.Id {
            return market, nil
        }
    }
    return nil, status.Errorf(codes.NotFound, "market not found")
}

func (m *marketService) SearchMarkets(ctx context.Context, req *pb.SearchMarketsRequest) (*pb.SearchMarketsResponse, error) {
    var results []*pb.Market
    for _, market := range m.markets {
        if matchesQuery(market, req.Query, req.Category) {
            results = append(results, market)
        }
    }
    return &pb.SearchMarketsResponse{Results: results}, nil
}

func matchesQuery(market *pb.Market, query, category string) bool {
    // Simple search implementation
    // TODO: Improve with better matching
    return true
}
```

#### 4. Seed 10 Demo Markets

Create `src/marketservice/demo_markets.go`:
```go
package main

import pb "yourrepo/src/marketservice/genproto"

func loadDemoMarkets() []*pb.Market {
    return []*pb.Market{
        {
            Id: "btc-150k-2025",
            Title: "Will Bitcoin reach $150,000 by June 2025?",
            Description: "Resolves YES if Bitcoin (BTC) trades at or above $150,000 on any major exchange before June 30, 2025.",
            ResolutionCriteria: "Based on CoinMarketCap daily high price",
            EndDate: 1751241600, // June 30, 2025
            Status: "open",
            YesPrice: 0.48,
            NoPrice: 0.52,
            TotalVolume: 12500.0,
            Category: "crypto",
        },
        {
            Id: "eth-5k-2025",
            Title: "Will Ethereum reach $5,000 by end of 2025?",
            Description: "Resolves YES if ETH trades at $5,000+ before Dec 31, 2025.",
            ResolutionCriteria: "CoinMarketCap daily high",
            EndDate: 1767225600,
            Status: "open",
            YesPrice: 0.62,
            NoPrice: 0.38,
            TotalVolume: 8900.0,
            Category: "crypto",
        },
        {
            Id: "superbowl-chiefs-2026",
            Title: "Will Kansas City Chiefs win Super Bowl LX?",
            Description: "Resolves YES if Chiefs win Super Bowl in February 2026.",
            ResolutionCriteria: "Official NFL result",
            EndDate: 1738627200,
            Status: "open",
            YesPrice: 0.35,
            NoPrice: 0.65,
            TotalVolume: 15200.0,
            Category: "sports",
        },
        // Add 7 more markets...
    }
}
```

### Day 5-6: Kafka Integration

**Priority: CRITICAL**

#### 1. Install Kafka Client
```bash
cd src/marketservice
go get github.com/confluentinc/confluent-kafka-go/v2/kafka
```

#### 2. Create Kafka Producer

Create `src/marketservice/kafka_producer.go`:
```go
package main

import (
    "encoding/json"
    "github.com/confluentinc/confluent-kafka-go/v2/kafka"
    "log"
)

type KafkaProducer struct {
    producer *kafka.Producer
}

func NewKafkaProducer(bootstrapServers, apiKey, apiSecret string) (*KafkaProducer, error) {
    p, err := kafka.NewProducer(&kafka.ConfigMap{
        "bootstrap.servers": bootstrapServers,
        "security.protocol": "SASL_SSL",
        "sasl.mechanisms":   "PLAIN",
        "sasl.username":     apiKey,
        "sasl.password":     apiSecret,
    })
    if err != nil {
        return nil, err
    }
    return &KafkaProducer{producer: p}, nil
}

func (kp *KafkaProducer) PublishMarketUpdate(marketID string, yesPrice, noPrice, volume float64) error {
    event := map[string]interface{}{
        "market_id": marketID,
        "yes_price": yesPrice,
        "no_price":  noPrice,
        "volume":    volume,
        "timestamp": time.Now().Unix(),
    }

    data, err := json.Marshal(event)
    if err != nil {
        return err
    }

    topic := "market-updates"
    return kp.producer.Produce(&kafka.Message{
        TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
        Value:          data,
    }, nil)
}
```

### Day 7: Simple Betting Engine

**Priority: HIGH**

#### 1. Create Betting Service

Create `src/bettingservice/main.go`:
```go
package main

import (
    "context"
    "errors"
    "sync"
)

type BettingEngine struct {
    markets map[string]*MarketState
    mu      sync.RWMutex
}

type MarketState struct {
    YesPool float64
    NoPool  float64
    K       float64  // Constant product (k = x * y)
}

func NewBettingEngine() *BettingEngine {
    return &BettingEngine{
        markets: make(map[string]*MarketState),
    }
}

// Simple CPMM (Constant Product Market Maker)
func (be *BettingEngine) PlaceBet(marketID string, side string, amount float64) (shares float64, newPrice float64, err error) {
    be.mu.Lock()
    defer be.mu.Unlock()

    market, exists := be.markets[marketID]
    if !exists {
        // Initialize market with equal liquidity
        market = &MarketState{
            YesPool: 1000.0,
            NoPool:  1000.0,
            K:       1000000.0,  // 1000 * 1000
        }
        be.markets[marketID] = market
    }

    if side == "YES" {
        // Buy YES shares
        newYesPool := market.YesPool + amount
        newNoPool := market.K / newYesPool
        shares = market.NoPool - newNoPool
        market.YesPool = newYesPool
        market.NoPool = newNoPool
        newPrice = newYesPool / (newYesPool + newNoPool)
    } else if side == "NO" {
        // Buy NO shares
        newNoPool := market.NoPool + amount
        newYesPool := market.K / newNoPool
        shares = market.YesPool - newYesPool
        market.NoPool = newNoPool
        market.YesPool = newYesPool
        newPrice = newYesPool / (newYesPool + newNoPool)
    } else {
        return 0, 0, errors.New("invalid side")
    }

    return shares, newPrice, nil
}

func (be *BettingEngine) GetPrice(marketID string) (yesPrice float64, noPrice float64) {
    be.mu.RLock()
    defer be.mu.RUnlock()

    market, exists := be.markets[marketID]
    if !exists {
        return 0.5, 0.5  // Default 50/50
    }

    total := market.YesPool + market.NoPool
    yesPrice = market.YesPool / total
    noPrice = market.NoPool / total
    return yesPrice, noPrice
}
```

## Week 2 Plan (Days 8-14)

### Critical Tasks:

1. **Vertex AI Integration** (Days 8-10)
   - Set up Vertex AI project
   - Create simple recommendation model
   - Deploy and test

2. **Gemini Integration** (Days 11-12)
   - Get Gemini API key
   - Implement market summary generation
   - Test with demo markets

3. **Basic Frontend** (Days 13-14)
   - Create simple web UI
   - Show market list
   - Allow betting
   - Display real-time odds

## Week 3 Plan (Days 15-21)

1. **Social Feed Service** (Days 15-17)
2. **Frontend Polish** (Days 18-20)
3. **End-to-end Testing** (Day 21)

## Week 4 Plan (Days 22-28)

1. **Deploy to Google Cloud Run** (Days 22-24)
2. **Performance Testing** (Days 25-26)
3. **Bug Fixes** (Days 27-28)

## Week 5 Plan (Days 29-35)

1. **UI Polish & Demo Data** (Days 29-32)
2. **Architecture Diagrams** (Days 33-34)
3. **Buffer Day** (Day 35)

## Week 6 Plan (Days 36-43)

1. **Record Demo Video** (Days 36-38)
2. **Documentation** (Days 39-40)
3. **Submission** (Days 41-42)
4. **Submit to Devpost** (Day 43)

## Critical Resources

### Confluent Resources
- [Confluent Cloud Quickstart](https://docs.confluent.io/cloud/current/get-started/index.html)
- [Kafka Go Client](https://docs.confluent.io/kafka-clients/go/current/overview.html)
- [ksqlDB Tutorial](https://docs.confluent.io/platform/current/ksqldb/tutorials/basics-tutorial.html)

### Google Cloud Resources
- [Vertex AI Quickstart](https://cloud.google.com/vertex-ai/docs/start/quickstarts)
- [Gemini API Docs](https://ai.google.dev/docs)
- [Cloud Run Quickstart](https://cloud.google.com/run/docs/quickstarts)

### Example Code References
- [Polymarket API Docs](https://docs.polymarket.com/)
- [Manifold Markets GitHub](https://github.com/manifoldmarkets/manifold)

## Daily Checklist Template

### Daily Standup (5 minutes)
- [ ] What did I accomplish yesterday?
- [ ] What will I accomplish today?
- [ ] Any blockers?

### End of Day (5 minutes)
- [ ] Commit code to GitHub
- [ ] Update todo list
- [ ] Document any issues

## Emergency Contacts & Help

### If Stuck on Confluent:
- Confluent Community Slack
- Stack Overflow tag: `apache-kafka`
- GitHub Issues: `confluentinc/confluent-kafka-go`

### If Stuck on Google Cloud:
- Google Cloud Community
- Stack Overflow tag: `google-cloud-platform`
- Vertex AI Discord

### If Stuck on Architecture:
- Review Polymarket for inspiration
- Review Manifold Markets code
- Ask in hackathon Discord

## Success Metrics

### Week 1 (Days 1-7)
- [ ] Confluent Cloud configured
- [ ] Google Cloud configured
- [ ] Market service listing 10 markets
- [ ] Kafka producing market events
- [ ] Simple bet execution working

### Week 2 (Days 8-14)
- [ ] Vertex AI recommendations working
- [ ] Gemini summaries working
- [ ] Basic frontend deployed
- [ ] Can place bets via UI

### Week 3 (Days 15-21)
- [ ] Social feed showing activity
- [ ] Real-time updates working
- [ ] Fraud detection demo working
- [ ] End-to-end flow complete

### Week 4 (Days 22-28)
- [ ] Deployed to Google Cloud
- [ ] All services connected
- [ ] Demo scenarios working
- [ ] Performance acceptable

### Week 5 (Days 29-35)
- [ ] UI looks professional
- [ ] Demo data realistic
- [ ] Architecture documented
- [ ] Ready to record video

### Week 6 (Days 36-43)
- [ ] Video recorded and edited
- [ ] GitHub documented
- [ ] Devpost submission complete
- [ ] SUBMITTED!

## Key Decision Points

### Decision 1: gRPC vs REST? (Day 3)
**Recommendation:** Use REST for MVP to simplify frontend integration
- gRPC for internal services
- REST API for frontend
- Gateway pattern if needed

### Decision 2: Database? (Day 4)
**Recommendation:** PostgreSQL for markets, Redis for positions
- PostgreSQL: Persistent market data
- Redis: Fast position lookups, social feed
- Skip complex database setup for MVP

### Decision 3: Authentication? (Day 10)
**Recommendation:** Simple demo accounts or anonymous for MVP
- Don't spend time on auth
- Use session IDs like original demo
- Add proper auth post-hackathon

## Risk Mitigation

### Risk: Confluent setup takes too long
**Mitigation:** Use Docker Kafka locally as backup, migrate to Confluent Cloud later

### Risk: Vertex AI model training slow
**Mitigation:** Use pre-trained models, simple rules-based recommendations as backup

### Risk: Frontend takes too long
**Mitigation:** Use existing frontend templates, keep UI minimal

### Risk: Running out of time
**Mitigation:** Cut scope aggressively, focus on core demo path

## Absolute Minimum for Submission

If you run out of time, you MUST have:
1. ✅ 5 markets viewable
2. ✅ Can place one bet
3. ✅ One Kafka topic working
4. ✅ One AI feature (either Vertex or Gemini)
5. ✅ Deployed and accessible
6. ✅ 3-minute video

Everything else is optional!

## Final Week Checklist (Dec 24-31)

### Dec 24-26: Video Production
- [ ] Script finalized
- [ ] Screen recording done
- [ ] Voiceover recorded
- [ ] Editing complete
- [ ] Uploaded to YouTube

### Dec 27-29: Documentation
- [ ] README.md polished
- [ ] Architecture diagram created
- [ ] Setup instructions tested
- [ ] Screenshots added

### Dec 30: Submission Prep
- [ ] Devpost form drafted
- [ ] All links tested
- [ ] Team info complete
- [ ] Proofread everything

### Dec 31: SUBMIT
- [ ] Submit before 11am GMT-11 (check your timezone!)
- [ ] Confirm submission received
- [ ] Share on social media
- [ ] Celebrate! 🎉

## You Got This!

Remember:
- **Ship > Perfect** - Done is better than perfect
- **Cut scope ruthlessly** - Focus on demo path
- **Ask for help early** - Don't waste time stuck
- **Document as you go** - Makes video easier
- **Test on real users** - Get feedback early

Good luck! 🚀
