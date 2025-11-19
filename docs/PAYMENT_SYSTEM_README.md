# VibeCast Markets - Crypto Payment System

## 🚀 Quick Start

### Running the Payment System

```bash
# Start all services including payment infrastructure
docker-compose -f docker-compose.payment.yml up

# Or run individually for testing
cd src/walletservice && go run main.go  # Port 8085
cd src/paymentservice && go run main.go # Port 8086
cd src/feeservice && go run main.go     # Port 8087
```

### Service Endpoints

| Service | Port | Health Check |
|---------|------|--------------|
| Wallet Service | 8085 | http://localhost:8085/health |
| Payment Service | 8086 | http://localhost:8086/health |
| Fee Service | 8087 | http://localhost:8087/health |

---

## 📚 API Documentation

### Wallet Service (Port 8085)

#### Create Wallet
```bash
POST /api/wallets
{
  "user_id": "user123"
}

Response:
{
  "id": "wallet_abc",
  "user_id": "user123",
  "balance": 0,
  "locked_balance": 0,
  "deposit_address": "0x...",
  "created_at": "2025-01-19T..."
}
```

#### Get Wallet
```bash
GET /api/wallets/:user_id

Response:
{
  "id": "wallet_abc",
  "user_id": "user123",
  "balance": 1000.50,
  "locked_balance": 200.00,
  "deposit_address": "0x...",
  "total_deposited": 1500.00,
  "total_withdrawn": 300.00,
  "total_earned": 50.50,
  "total_fees_paid": 50.00
}
```

#### Deposit
```bash
POST /api/deposit
{
  "user_id": "user123",
  "amount": 100.00,
  "currency": "USDC",
  "tx_hash": "0x..."
}
```

#### Withdraw
```bash
POST /api/withdraw
{
  "user_id": "user123",
  "amount": 50.00,
  "withdraw_address": "0x...",
  "currency": "USDC"
}
```

#### Get Transaction History
```bash
GET /api/transactions/:user_id

Response:
{
  "transactions": [
    {
      "id": "tx_123",
      "type": "deposit",
      "amount": 100.00,
      "fee": 0,
      "status": "completed",
      "created_at": "..."
    }
  ],
  "total": 5
}
```

---

### Payment Service (Port 8086)

#### Create Deposit Request
```bash
POST /api/deposits
{
  "user_id": "user123",
  "currency": "USDC"
}

Response:
{
  "deposit_id": "dep_abc",
  "deposit_address": "0x123...",
  "currency": "USDC",
  "status": "pending",
  "instructions": "Send USDC to address: 0x123...",
  "required_confirmations": 10
}
```

#### Get Deposit Status
```bash
GET /api/deposits/:id

Response:
{
  "id": "dep_abc",
  "user_id": "user123",
  "amount": 100.00,
  "currency": "USDC",
  "deposit_address": "0x123...",
  "tx_hash": "0xabc...",
  "confirmations": 10,
  "status": "completed",
  "created_at": "...",
  "completed_at": "..."
}
```

#### Create Withdrawal Request
```bash
POST /api/withdrawals
{
  "user_id": "user123",
  "amount": 100.00,
  "to_address": "0x456...",
  "currency": "USDC"
}

Response:
{
  "withdrawal_id": "wd_xyz",
  "amount": 100.00,
  "to_address": "0x456...",
  "platform_fee": 1.00,
  "network_fee": 0.05,
  "total_cost": 101.05,
  "status": "pending",
  "currency": "USDC"
}
```

#### Get Supported Currencies
```bash
GET /api/currencies

Response:
{
  "currencies": [
    {
      "code": "USDC",
      "name": "USD Coin",
      "network": "Polygon",
      "decimals": 6,
      "min_deposit": 10.0,
      "min_withdraw": 20.0
    }
  ]
}
```

---

### Fee Service (Port 8087)

#### Calculate Trading Fee
```bash
POST /api/fees/calculate/trading
{
  "user_id": "user123",
  "amount": 1000.00
}

Response:
{
  "amount": 1000.00,
  "fee_type": "trading",
  "fee_percent": 2.5,
  "fee_amount": 25.00,
  "total_fee": 25.00,
  "user_tier": "standard",
  "breakdown": {
    "trading_fee": 25.00
  }
}
```

#### Calculate Withdrawal Fee
```bash
POST /api/fees/calculate/withdrawal
{
  "amount": 500.00,
  "currency": "USDC"
}

Response:
{
  "amount": 500.00,
  "fee_type": "withdrawal",
  "fee_percent": 1.0,
  "fee_amount": 5.00,
  "network_fee": 0.05,
  "total_fee": 5.05,
  "breakdown": {
    "platform_fee": 5.00,
    "network_fee": 0.05
  }
}
```

#### Get User Fee Stats
```bash
GET /api/fees/users/:user_id

Response:
{
  "user_id": "user123",
  "tier": "bronze",
  "monthly_volume": 15000.00,
  "trading_fee": 2.0,
  "next_tier": "silver",
  "volume_to_next": 35000.00
}
```

#### Get Platform Revenue
```bash
GET /api/fees/revenue

Response:
{
  "total_revenue": 12500.50,
  "revenue_by_type": {
    "trading": 11250.00,
    "withdrawal": 1250.50
  },
  "revenue_by_tier": {
    "standard": 6000.00,
    "bronze": 4500.00,
    "silver": 2000.50
  },
  "total_volume": 500000.00,
  "transaction_count": 450,
  "average_revenue_per_tx": 27.78,
  "last_updated": "2025-01-19T..."
}
```

#### Get Fee Tiers
```bash
GET /api/fees/tiers

Response:
{
  "tiers": [
    {
      "name": "Standard",
      "min_volume": 0,
      "max_volume": 9999,
      "trading_fee": 2.5,
      "benefits": ["Standard trading fees", "24/7 support"]
    },
    {
      "name": "Bronze",
      "min_volume": 10000,
      "max_volume": 49999,
      "trading_fee": 2.0,
      "benefits": ["Reduced fees", "Priority support"]
    }
  ]
}
```

---

## 💸 Complete User Flow Example

### 1. User Deposits $100 USDC

```bash
# Step 1: Create wallet (if not exists)
curl -X POST http://localhost:8085/api/wallets \
  -H "Content-Type: application/json" \
  -d '{"user_id": "alice"}'

# Step 2: Request deposit address
curl -X POST http://localhost:8086/api/deposits \
  -H "Content-Type: application/json" \
  -d '{"user_id": "alice", "currency": "USDC"}'

# Returns: { "deposit_address": "0x123...", ... }

# Step 3: User sends USDC from their wallet to deposit address
# (Done externally via MetaMask, Coinbase Wallet, etc.)

# Step 4: Payment service monitors blockchain and detects deposit
# After 10 confirmations, automatically credits wallet

# Step 5: Check wallet balance
curl http://localhost:8085/api/wallets/alice
# Returns: { "balance": 100.00, ... }
```

### 2. User Places $50 Bet

```bash
# Calculate fee first
curl -X POST http://localhost:8087/api/fees/calculate/trading \
  -H "Content-Type: application/json" \
  -d '{"user_id": "alice", "amount": 50.00}'
# Returns: { "fee_amount": 1.25, ... }

# Place bet (calls wallet service internally)
curl -X POST http://localhost:8082/api/bets \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "alice",
    "market_id": "market_crypto",
    "side": "YES",
    "amount": 50.00
  }'

# Wallet now shows:
# - Balance: 48.75 (100 - 50 - 1.25 fee)
# - Locked: 50.00
# - Fees paid: 1.25
```

### 3. Market Resolves - User Wins

```bash
# Market resolves YES, user wins!
# Payout service calculates: 50 shares × $1 = $75 payout

# Wallet service unlocks and credits:
# - Locked: 0
# - Balance: 123.75 (48.75 + 75)
# - Total earned: 25.00 (75 - 50 original bet)
```

### 4. User Withdraws $100

```bash
# Calculate withdrawal fee
curl -X POST http://localhost:8087/api/fees/calculate/withdrawal \
  -H "Content-Type: application/json" \
  -d '{"amount": 100.00, "currency": "USDC"}'
# Returns: { "platform_fee": 1.00, "network_fee": 0.05, "total_fee": 1.05 }

# Request withdrawal
curl -X POST http://localhost:8086/api/withdrawals \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "alice",
    "amount": 100.00,
    "to_address": "0xalice_wallet",
    "currency": "USDC"
  }'

# Payment service:
# 1. Deducts 101.05 from wallet
# 2. Sends blockchain transaction
# 3. User receives 100 USDC in their wallet
```

---

## 🔧 Configuration

### Environment Variables

```bash
# Wallet Service
PORT=8085
KAFKA_BROKERS=kafka:9092

# Payment Service
PORT=8086
WALLET_SERVICE_URL=http://localhost:8085
BLOCKCHAIN_NETWORK=polygon-mainnet
BLOCKCHAIN_RPC=https://polygon-rpc.com
USDC_CONTRACT=0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174
USDT_CONTRACT=0xc2132D05D31c914a87C6611C10748AEb04B58e8F

# Fee Service
PORT=8087
```

### Fee Configuration

Edit `feeservice/main.go` to customize:

```go
TradingFees: map[FeeTier]float64{
    TierStandard: 0.025, // 2.5%
    TierBronze:   0.020, // 2.0%
    TierSilver:   0.015, // 1.5%
    TierGold:     0.010, // 1.0%
},
WithdrawalFeePercent: 0.01,  // 1%
WithdrawalFeeMin:     1.0,   // $1 minimum
```

---

## 🏗️ Architecture

```
┌────────────────────────────────────────────────────────┐
│                      Frontend                          │
│              (Next.js + TypeScript)                    │
└───────────────────┬────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────────────────────┐
│                   API Gateway                          │
│                  (Port 9000)                           │
└──┬──────────┬──────────┬──────────┬──────────┬─────────┘
   │          │          │          │          │
   ▼          ▼          ▼          ▼          ▼
┌─────┐  ┌─────┐  ┌─────┐  ┌──────┐  ┌──────┐
│Wallet│  │Payment│ │Fee  │  │Betting│ │Market│
│8085 │  │8086  │  │8087 │  │8082  │  │8081  │
└──┬──┘  └──┬───┘  └─────┘  └──┬───┘  └──────┘
   │        │                   │
   └────────┴───────────────────┴───────────────┐
                                                 │
                    ┌────────────────────────────▼┐
                    │    Kafka Event Bus         │
                    │  (Transaction Events)      │
                    └────────────────────────────┘
```

---

## 🔒 Security Best Practices

### Current Implementation (Demo/MVP)
- ✅ In-memory wallet storage
- ✅ Mock blockchain addresses
- ✅ Simulated transactions
- ⚠️ No encryption
- ⚠️ No hot/cold wallet separation

### Production Requirements
- [ ] PostgreSQL for persistent storage
- [ ] Encrypted private keys (HSM recommended)
- [ ] Hot wallet (5%) / Cold wallet (95%) split
- [ ] Multi-signature withdrawals
- [ ] Rate limiting on withdrawals
- [ ] KYC/AML compliance
- [ ] Real blockchain integration
- [ ] Transaction monitoring
- [ ] Audit logging

---

## 📊 Testing

### Manual Testing

```bash
# Test wallet creation
curl -X POST http://localhost:8085/api/wallets \
  -H "Content-Type: application/json" \
  -d '{"user_id": "test_user"}'

# Test deposit
curl -X POST http://localhost:8085/api/deposit \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test_user",
    "amount": 100,
    "currency": "USDC",
    "tx_hash": "0xtest123"
  }'

# Test fee calculation
curl -X POST http://localhost:8087/api/fees/calculate/trading \
  -H "Content-Type: application/json" \
  -d '{"user_id": "test_user", "amount": 100}'

# Check revenue
curl http://localhost:8087/api/fees/revenue
```

---

## 📈 Monitoring & Metrics

### Key Metrics to Track

1. **Financial Metrics**
   - Total deposits (daily/monthly)
   - Total withdrawals
   - Platform revenue
   - Revenue by fee type
   - Average transaction value

2. **User Metrics**
   - Active wallets
   - Deposit/withdrawal success rates
   - Average user balance
   - Fee tier distribution

3. **System Metrics**
   - Transaction processing time
   - Blockchain confirmation time
   - Failed transactions
   - Pending withdrawals

### Accessing Metrics

```bash
# Platform revenue
curl http://localhost:8087/api/fees/revenue

# User stats
curl http://localhost:8087/api/fees/users/alice

# Transaction history
curl http://localhost:8085/api/transactions/alice
```

---

## 🚦 Deployment Checklist

### Before Production

- [ ] Switch to real blockchain RPC endpoints
- [ ] Set up hot/cold wallet infrastructure
- [ ] Implement proper key management (HSM)
- [ ] Add PostgreSQL database
- [ ] Set up transaction monitoring
- [ ] Implement KYC/AML checks
- [ ] Add rate limiting
- [ ] Set up logging and monitoring
- [ ] Create backup/recovery procedures
- [ ] Test withdrawal limits
- [ ] Security audit
- [ ] Load testing

---

## 📚 Additional Resources

- [Monetary Model Documentation](./MONETARY_MODEL.md)
- [Polygon Network Docs](https://docs.polygon.technology/)
- [USDC Contract](https://polygonscan.com/token/0x2791bca1f2de4661ed88a30c99a7a9449aa84174)
- [Web3 Integration Guide](https://web3js.readthedocs.io/)

---

**Built for Google AI Partner Catalyst Hackathon 2025**
