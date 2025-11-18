# Vibe Markets - Social Prediction Market Platform

**Hackathon:** Google AI Partner Catalyst
**Challenge:** Confluent - Real-time AI Application
**Deadline:** December 31, 2025

## Overview

Vibe Markets combines social media dynamics with prediction markets, powered by real-time streaming data (Confluent) and AI (Google Vertex AI + Gemini). Users can bet on outcomes, share predictions, and compete on leaderboards - all in real-time.

## The Problem

- **Polymarket** has $3.7B+ volume but lacks virality and user growth mechanisms
- **Social media** has distribution but struggles with monetization
- **Traditional prediction markets** have poor UX and liquidity bootstrapping problems

## Our Solution

A social-first prediction market where:
- Real-time odds updates via Confluent Kafka
- AI-powered market recommendations via Vertex AI
- Gemini-generated market insights and summaries
- Social feed creates viral growth loops
- Every bet becomes shareable content

## Technical Architecture

### Microservices Adaptation

| Original Service | New Purpose | Technology |
|-----------------|-------------|------------|
| `productcatalogservice` | **marketservice** - Browse and search prediction markets | Go + PostgreSQL |
| `cartservice` | **positionservice** - Track user betting positions | C# + Redis |
| `checkoutservice` | **bettingservice** - Execute bets and order matching | Go + Kafka |
| `currencyservice` | **balanceservice** - User balance management (play money) | Node.js |
| `frontend` | **Social betting UI** - Feed, markets, leaderboards | Go (templates) |
| `loadgenerator` | **Demo traffic generator** - Simulate realistic usage | Python/Locust |

### New Services

| Service | Purpose | Technology |
|---------|---------|------------|
| `socialfeedservice` | Generate personalized activity feeds | Go + Redis |
| `resolutionservice` | Resolve market outcomes, calculate payouts | Python |
| `kafkaproducer` | Centralized event publishing to Confluent | Go |

### Removed Services (Not needed for MVP)

- `adservice` - Not needed for demo
- `emailservice` - No email notifications in MVP
- `paymentservice` - Using play money only
- `shippingservice` - Not applicable
- `recommendationservice` - Replaced by Vertex AI

## Confluent Integration

### Kafka Topics (Event Streams)

1. **`bet-orders`**
   - User bet placement events
   - Consumed by: BettingService, FraudDetectionService
   - Schema: `{user_id, market_id, amount, side (YES/NO), timestamp}`

2. **`market-updates`**
   - Real-time price and volume changes
   - Consumed by: Frontend (WebSocket), AnalyticsService
   - Schema: `{market_id, yes_price, no_price, volume, timestamp}`

3. **`social-feed-events`**
   - User actions (bets, wins, market creation)
   - Consumed by: SocialFeedService
   - Schema: `{event_type, user_id, market_id, data, timestamp}`

4. **`user-activity`**
   - User behavior for ML training
   - Consumed by: Vertex AI training pipeline
   - Schema: `{user_id, action, market_id, timestamp, context}`

5. **`resolution-events`**
   - Market outcomes and payouts
   - Consumed by: BalanceService, NotificationService
   - Schema: `{market_id, outcome, winner_payouts[], timestamp}`

### Confluent Features Demonstrated

- **Kafka Streams** - Process bet orders in real-time
- **ksqlDB** - SQL queries for leaderboards, live statistics
- **Schema Registry** - Ensure data consistency across services
- **Flink** - Fraud detection (wash trading, suspicious patterns)
- **Connectors** - Sync to BigQuery for analytics

## Google Cloud AI Integration

### Vertex AI Use Cases

1. **Market Recommendation Model**
   - Collaborative filtering based on user history
   - Input: User bet history, market views, social graph
   - Output: Top 5 recommended markets
   - Model: AutoML Tables or custom TensorFlow

2. **Fraud Detection Model**
   - Anomaly detection on betting patterns
   - Input: Bet timing, amounts, user behavior
   - Output: Fraud probability score (0-1)
   - Model: Isolation Forest (scikit-learn)

3. **Market Sentiment Analysis**
   - Analyze comments and social posts
   - Input: Text from comments/shares
   - Output: Bullish/bearish sentiment
   - Model: Fine-tuned BERT

### Gemini Use Cases

1. **Market Summary Generation**
   - Input: Market data, recent bets, news context
   - Output: Human-readable market insights
   - Example: "Bitcoin is 52% likely to hit $150k by June based on 1,234 bets totaling $45k volume. Odds shifted 8% in last hour."

2. **Market Quality Scoring**
   - Input: User-created market description
   - Output: Quality score + improvement suggestions
   - Prevents ambiguous/spam markets

3. **Conversational Market Creation**
   - Users describe markets in natural language
   - Gemini generates formal market parameters
   - Example: "Create a market for next UFC fight" → Generates title, description, resolution criteria

## Data Models

### Market

```go
type Market struct {
    ID              string
    Title           string
    Description     string
    ResolutionCriteria string
    EndDate         time.Time
    CreatorID       string
    Status          string // "open", "closed", "resolved"
    Outcome         string // "YES", "NO", "CANCELED"
    YesPrice        float64
    NoPrice         float64
    TotalVolume     float64
    YesPool         float64
    NoPool          float64
    CreatedAt       time.Time
}
```

### User Position

```go
type Position struct {
    ID          string
    UserID      string
    MarketID    string
    Side        string // "YES" or "NO"
    Shares      float64
    AverageCost float64
    CurrentValue float64
    PnL         float64
    CreatedAt   time.Time
}
```

### Bet Order

```go
type BetOrder struct {
    ID        string
    UserID    string
    MarketID  string
    Side      string // "YES" or "NO"
    Amount    float64
    Price     float64
    Status    string // "pending", "executed", "failed"
    CreatedAt time.Time
}
```

## AMM (Automated Market Maker) Logic

Using Constant Product Market Maker (CPMM) formula:

```
k = yesPool * noPool (constant)

For a bet of amount A on YES:
newYesPool = yesPool + A
newNoPool = k / newYesPool

Shares received = yesPool - newYesPool
Average price = A / sharesReceived

New YES price = newYesPool / (newYesPool + newNoPool)
```

## MVP Feature Scope

### ✅ MUST HAVE

- [ ] Browse 10 pre-seeded markets
- [ ] View market details with AI-generated summary (Gemini)
- [ ] Place bets (YES/NO, simple market order)
- [ ] Real-time odds updates via Kafka streams
- [ ] Social activity feed (recent bets, wins)
- [ ] Leaderboard (top traders by P&L)
- [ ] Vertex AI market recommendations
- [ ] Fraud detection demo (flag suspicious pattern)
- [ ] Mobile-responsive web UI
- [ ] Play money only ($1,000 starting balance)

### ❌ SKIP FOR MVP

- User-generated markets (only pre-seeded markets)
- Comments/discussions
- Follow/unfollow users
- Push notifications
- Advanced trading (limit orders, shorts)
- Native mobile apps
- Real money/crypto integration
- Authentication (simple demo accounts or anonymous)

## Demo Scenarios

### Scenario 1: Real-Time Betting
1. User A bets $100 on "Bitcoin hits $150k" (YES)
2. Event published to `bet-orders` Kafka topic
3. BettingService consumes, executes trade
4. Price updates from 48% → 49%
5. Event published to `market-updates`
6. Frontend receives update via WebSocket
7. All users see new price instantly

### Scenario 2: AI Recommendations
1. User views market about Bitcoin
2. Vertex AI model analyzes user history
3. Recommends related markets: "Ethereum price", "Crypto regulation"
4. User clicks recommendation, places bet
5. Feedback loop improves recommendations

### Scenario 3: Fraud Detection
1. Suspicious user places rapid buy/sell orders
2. Flink job detects wash trading pattern
3. Alert published to `fraud-alerts` topic
4. Admin dashboard flags account
5. Demo shows real-time fraud prevention

### Scenario 4: Social Virality
1. User wins $250 on Bitcoin market
2. "You won!" notification appears
3. One-click share to Twitter
4. Tweet shows: "I just won $250 betting Bitcoin hits $150k! Think I'm wrong? Prove it: [link]"
5. Friends click, see market, place counter-bets

## Development Phases

### Week 1: Foundation (Days 1-7)
- Set up Confluent Cloud
- Set up Google Cloud (Vertex AI, Cloud Run)
- Create Kafka topics
- Basic Go service → Kafka integration working

### Week 2: Core Betting (Days 8-14)
- MarketService (CRUD, 10 demo markets)
- Simple AMM implementation
- BettingService (order execution)
- Real-time updates working

### Week 3: AI Integration (Days 15-21)
- ksqlDB for real-time stats
- Vertex AI recommendation model
- Gemini market summaries
- Fraud detection (Flink)

### Week 4: Social Feed (Days 22-28)
- SocialFeedService
- Activity stream
- Frontend polish
- Deploy to Cloud Run

### Week 5: Polish (Days 29-35)
- UI/UX improvements
- Mobile responsiveness
- Demo data seeding
- Performance optimization

### Week 6: Submission (Days 36-43)
- Record demo video
- Write documentation
- Create architecture diagrams
- Submit to Devpost

## Demo Video Script (3 minutes)

**[0:00-0:30] Hook**
"Polymarket hit $3.7 billion in volume, but it has a fatal flaw: no virality. Social media has distribution but can't monetize. We combined both with Vibe Markets."

**[0:30-1:00] Real-Time Betting Demo**
Show: Live bet → Kafka event → Instant odds update

**[1:00-1:30] AI Features**
Show: Vertex AI recommendations + Gemini summaries

**[1:30-2:00] Fraud Detection**
Show: Flink detecting suspicious pattern in real-time

**[2:00-2:30] Social Features**
Show: Activity feed, leaderboard, share to Twitter

**[2:30-3:00] Impact**
"With Confluent handling millions of events per second and Vertex AI ensuring fair markets, we can scale to millions of users. This is the future of social media monetization."

## Success Metrics

### Judging Criteria

1. **Innovation (30%)** - First social prediction market with real-time streaming
2. **Technical Execution (30%)** - Confluent + Vertex AI + Gemini integration
3. **Problem/Market Fit (20%)** - Solves real liquidity + virality problem
4. **Design/UX (10%)** - Polished, mobile-first interface
5. **Completeness (10%)** - Working demo, open source, documented

### Target Metrics (Demo)

- 10 pre-seeded markets
- 1,000+ demo bets executed
- Sub-100ms latency for odds updates
- 15+ fraud patterns detected
- 100% uptime during judging

## Technology Stack

### Backend
- **Go** - MarketService, BettingService, SocialFeedService, Frontend
- **C#/.NET** - PositionService (from CartService)
- **Node.js** - BalanceService
- **Python** - ResolutionService, Fraud Detection

### Data
- **PostgreSQL** - Markets, users, orders
- **Redis** - Positions, social feed, caching
- **Confluent Kafka** - Event streaming
- **BigQuery** - Analytics

### AI/ML
- **Vertex AI** - Recommendation models, fraud detection
- **Gemini Pro** - Market summaries, quality scoring

### Infrastructure
- **Google Cloud Run** - Serverless deployment
- **Confluent Cloud** - Managed Kafka
- **Docker** - Containerization
- **Kubernetes** - Optional (if needed)

## Repository Structure

```
vibe-markets/
├── src/
│   ├── marketservice/         # Browse/search markets
│   ├── bettingservice/        # Execute bets, AMM logic
│   ├── positionservice/       # User positions (from cartservice)
│   ├── balanceservice/        # User balances
│   ├── socialfeedservice/     # Activity feed
│   ├── resolutionservice/     # Market resolution
│   ├── frontend/              # Web UI
│   └── loadgenerator/         # Demo traffic
├── kafka/
│   ├── topics.yaml            # Kafka topic definitions
│   └── schemas/               # Avro schemas
├── ai/
│   ├── vertex-recommendations/ # Recommendation model
│   ├── vertex-fraud/          # Fraud detection
│   └── gemini-integration/    # Gemini API calls
├── docs/
│   ├── architecture.md
│   ├── api.md
│   └── demo-script.md
├── kubernetes-manifests/      # K8s deployment (optional)
├── docker-compose.yaml        # Local development
└── README.md
```

## Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.23+
- Node.js 18+
- Python 3.11+
- Confluent Cloud account
- Google Cloud account with Vertex AI enabled

### Local Development

```bash
# Clone repository
git clone https://github.com/yourusername/vibe-markets.git
cd vibe-markets

# Set environment variables
cp .env.example .env
# Add Confluent and GCP credentials

# Start services
docker-compose up

# Visit http://localhost:8080
```

## Deployment

### Google Cloud Run

```bash
# Build and deploy
./deploy.sh

# Services will be available at:
# https://vibe-markets-frontend-xxx.run.app
```

## License

Apache 2.0 (Open Source for hackathon requirements)

## Team

[Your name/team]

## Links

- **Live Demo:** https://vibe-markets-demo.run.app
- **GitHub:** https://github.com/yourusername/vibe-markets
- **Video:** https://youtube.com/watch?v=xxx
- **Devpost:** https://devpost.com/software/vibe-markets

---

Built for Google AI Partner Catalyst Hackathon 2025
