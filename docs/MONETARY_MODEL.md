# VibeCast Markets - Monetary Model & Payment System

## 📊 Complete Financial Architecture

This document explains how money flows through VibeCast Markets, how users earn, how the platform earns, and how all transactions work.

---

## 🏦 System Overview

```
┌─────────────┐
│   User      │
└──────┬──────┘
       │
       │ Deposit USDC/USDT
       ▼
┌──────────────────┐
│ Payment Service  │ ← Monitors blockchain
│ (Port 8086)      │ ← Generates addresses
└────────┬─────────┘
         │ Credits
         ▼
┌──────────────────┐
│ Wallet Service   │ ← Manages balances
│ (Port 8085)      │ ← Locks funds for bets
└────────┬─────────┘
         │ Places bet
         ▼
┌──────────────────┐
│ Betting Service  │ ← CPMM Algorithm
│ (Port 8082)      │ ← Market resolution
└────────┬─────────┘
         │ Win/Loss
         ▼
┌──────────────────┐
│ Payout Service   │ ← Distributes winnings
│                  │ ← Automated settlements
└────────┬─────────┘
         │ Withdraw
         ▼
┌──────────────────┐
│ Payment Service  │ ← Sends blockchain tx
│                  │ ← Processes withdrawals
└──────────────────┘
```

---

## 💰 How Users Earn Money

### 1. **Winning Predictions**
- Place bet on outcome (YES/NO)
- If correct when market resolves → earn payout
- Example:
  - Bet $100 on YES at 60% odds
  - Get 166.67 shares
  - If YES wins → receive $166.67
  - **Profit: $66.67** (before fees)

### 2. **Early Trading**
- Buy shares at low price
- Sell when price increases
- Example:
  - Buy YES at 40% for $100 → get 250 shares
  - Price moves to 60% → shares now worth $150
  - Sell for **$50 profit** (before fees)

### 3. **Market Making**
- Provide liquidity to markets
- Earn from bid-ask spreads
- (Future feature)

### 4. **Referral Bonuses**
- Invite friends → earn % of their trading fees
- Tiered rewards: 5% → 10% → 15%
- (Future feature)

### 5. **Leaderboard Rewards**
- Weekly/monthly competitions
- Top traders win prize pools
- Rankings by volume, win rate, profit
- (Future feature)

---

## 💵 How Platform Earns Revenue

### 1. **Trading Fees (Primary Revenue)**
- **2.5% fee on every bet placed**
- Deducted automatically when placing bet
- Example: $100 bet → $2.50 platform fee

**Annual Projection:**
- 10,000 active users
- $500 avg monthly volume per user
- Total monthly volume: $5M
- **Monthly revenue: $125,000** (2.5% of $5M)
- **Annual revenue: $1.5M**

### 2. **Withdrawal Fees**
- **1% fee on withdrawals (min $1)**
- Plus blockchain network fee (passed through)
- Example: $500 withdrawal → $5 platform fee + $0.05 network fee

**Annual Projection:**
- 20% of users withdraw monthly
- Avg withdrawal: $300
- **Monthly revenue: $6,000**
- **Annual revenue: $72,000**

### 3. **Market Creation Fees**
- Premium users can create markets
- $50-500 fee depending on category
- (Future feature)

### 4. **Sponsored Markets**
- Brands pay to create branded markets
- $1,000-10,000 per market
- Example: "Will [Product] launch by Q2?"
- (Future feature)

### 5. **Premium Subscriptions**
- Advanced analytics: $10/month
- AI insights & recommendations: $20/month
- API access: $50/month
- (Future feature)

### 6. **Data Licensing**
- Sell aggregated prediction data
- Research institutions, hedge funds
- (Future feature)

**Total Potential Annual Revenue:**
- Year 1: **$1.5M-2M**
- Year 2: **$5M-8M** (with growth)
- Year 3: **$15M-25M** (with all features)

---

## 🔄 Transaction Flow Detailed

### Deposit Flow

```
1. User requests deposit
   POST /api/deposits
   { "user_id": "user123", "currency": "USDC" }

2. Payment Service generates unique address
   Response: { "deposit_address": "0xABC..." }

3. User sends USDC from their wallet
   Blockchain transaction

4. Payment Service monitors blockchain
   - Detects incoming transaction
   - Waits for confirmations (10 blocks)

5. Payment Service credits Wallet Service
   POST /api/deposit
   { "user_id": "user123", "amount": 100, "tx_hash": "0x..." }

6. Wallet Service updates balance
   Balance: $0 → $100

7. User can now place bets!
```

### Betting Flow

```
1. User places bet
   POST /api/bets
   { "user_id": "user123", "market_id": "m1", "side": "YES", "amount": 100 }

2. Wallet Service locks funds
   - Calculate fee: $100 × 2.5% = $2.50
   - Total deducted: $102.50
   - Balance: $100 → $0 (available)
   - Locked: $0 → $100 (in bet)
   - Fees paid: $0 → $2.50 (platform revenue)

3. Betting Service calculates shares
   - YES pool: $1000, NO pool: $666.67, K = 666,667
   - Add $100 to YES pool → new YES pool = $1100
   - New NO pool: K / 1100 = 606.06
   - Shares received: 666.67 - 606.06 = 60.61 shares

4. Transaction recorded
   Type: "bet", Amount: $100, Fee: $2.50, Status: "completed"
```

### Payout Flow (When Market Resolves)

```
1. Market resolves to YES
   Admin resolves market, outcome = "YES"

2. Payout Service calculates winnings
   For each YES bet:
   - Shares: 60.61
   - Price per share: $1 (YES won)
   - Winnings: 60.61 × $1 = $60.61

3. Wallet Service unlocks + credits
   - Locked: $100 → $0
   - Balance: $0 → $60.61
   - Total earned: $0 → $60.61 (net profit after original bet)

4. Transaction recorded
   Type: "win", Amount: $60.61, Status: "completed"
```

### Withdrawal Flow

```
1. User requests withdrawal
   POST /api/withdrawals
   { "user_id": "user123", "amount": 50, "to_address": "0xXYZ...", "currency": "USDC" }

2. Wallet Service validates + deducts
   - Calculate fees: $50 × 1% = $0.50 (min $1) → $1 platform fee
   - Network fee: $0.05 (Polygon)
   - Total deducted: $51.05
   - Balance: $60.61 → $9.56

3. Payment Service processes
   - Creates blockchain transaction
   - Sends $50 USDC to user's address
   - Waits for confirmation

4. Transaction completes
   - Status: "pending" → "completed"
   - TX hash: 0x...

5. User receives $50 USDC in their wallet
```

---

## 💳 Payment Methods & Networks

### Supported Cryptocurrencies

| Currency | Network | Contract Address | Min Deposit | Min Withdraw |
|----------|---------|------------------|-------------|--------------|
| **USDC** | Polygon | 0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174 | $10 | $20 |
| **USDT** | Polygon | 0xc2132D05D31c914a87C6611C10748AEb04B58e8F | $10 | $20 |

**Why Polygon?**
- ✅ Extremely low fees (~$0.01-0.10 per transaction)
- ✅ Fast confirmations (~2 seconds)
- ✅ Ethereum-compatible (easy integration)
- ✅ Widely supported by exchanges and wallets
- ✅ Same USDC/USDT as mainnet (bridged)

**Future Networks:**
- Arbitrum (Layer 2)
- Optimism (Layer 2)
- Base (Coinbase L2)

---

## 📈 Fee Structure

### Trading Fees
- **Standard: 2.5%** per bet
- High volume discount tiers:
  - $10K+ monthly volume: **2.0%**
  - $50K+ monthly volume: **1.5%**
  - $100K+ monthly volume: **1.0%**

### Withdrawal Fees
- **Platform fee: 1%** (minimum $1)
- **Network fee: $0.05-0.10** (Polygon gas)
- No fees on deposits

### Example Total Costs

**Scenario: User bets $1,000 and wins $1,500**
- Deposit: $1,000 (no fee)
- Place bet: $1,000 + $25 fee = **$1,025 deducted**
- Win: $1,500 credited
- Withdraw: $1,500 - $15 fee - $0.05 network = **$1,484.95 received**
- **Net profit: $484.95 - $25 initial fee = $459.95**
- Effective fee rate: **4.0%** of winnings

---

## 🔐 Security & Compliance

### Wallet Security
- **Hot wallet**: For daily operations (withdrawals)
- **Cold wallet**: For majority of funds (95%+)
- Multi-signature requirements
- Daily withdrawal limits
- Automated security monitoring

### AML/KYC (Future)
- Identity verification for withdrawals >$1,000
- Transaction monitoring
- Suspicious activity reporting
- Compliance with FinCEN regulations

### Auditing
- All transactions logged to blockchain
- Immutable audit trail
- Real-time reconciliation
- Monthly financial reports

---

## 📊 Platform Revenue Dashboard

### Key Metrics Tracked

```javascript
GET /api/admin/revenue

Response:
{
  "total_revenue": 125000.50,
  "revenue_by_type": {
    "trading_fees": 112500.00,
    "withdrawal_fees": 12500.50
  },
  "total_volume": 5000000.00,
  "total_users": 10000,
  "active_users_30d": 4500,
  "avg_revenue_per_user": 12.50,
  "revenue_growth_30d": 25.5
}
```

### Revenue Breakdown
- **Trading fees**: 90% of revenue
- **Withdrawal fees**: 10% of revenue
- **Target margins**: 70-80% (after operational costs)

---

## 🚀 Future Enhancements

### Phase 2 (Q2 2025)
- [ ] Credit/debit card deposits (Stripe)
- [ ] Bank transfers (ACH, wire)
- [ ] More cryptocurrencies (ETH, SOL, MATIC)
- [ ] Tiered fee structure
- [ ] Referral program

### Phase 3 (Q3 2025)
- [ ] Premium subscriptions
- [ ] API access for developers
- [ ] Sponsored markets
- [ ] Market creation fees
- [ ] NFT badges for top traders

### Phase 4 (Q4 2025)
- [ ] DeFi integration (stake USDC, earn yield)
- [ ] DAO governance token
- [ ] Decentralized market creation
- [ ] Cross-chain bridges

---

## 🧮 Revenue Projections

### Conservative Scenario (Year 1)
- 5,000 monthly active users
- $300 avg volume per user per month
- Total monthly volume: $1.5M
- Revenue at 2.5%: **$37,500/month**
- **Annual: $450K**

### Moderate Scenario (Year 1)
- 10,000 monthly active users
- $500 avg volume per user per month
- Total monthly volume: $5M
- Revenue at 2.5%: **$125,000/month**
- **Annual: $1.5M**

### Aggressive Scenario (Year 1)
- 25,000 monthly active users
- $800 avg volume per user per month
- Total monthly volume: $20M
- Revenue at 2.5%: **$500,000/month**
- **Annual: $6M**

---

## 🔗 Integration Guide

### For Frontend Developers

```typescript
// 1. Create deposit
const response = await fetch('http://localhost:8086/api/deposits', {
  method: 'POST',
  body: JSON.stringify({
    user_id: 'user123',
    currency: 'USDC'
  })
});

const { deposit_address } = await response.json();
// Show QR code with deposit address

// 2. Check deposit status
const deposit = await fetch(`http://localhost:8086/api/deposits/${depositId}`);
// Poll until status === 'completed'

// 3. Request withdrawal
await fetch('http://localhost:8086/api/withdrawals', {
  method: 'POST',
  body: JSON.stringify({
    user_id: 'user123',
    amount: 100,
    to_address: '0x...',
    currency: 'USDC'
  })
});
```

---

## 📞 Support & Questions

For questions about the monetary model:
- Technical: Check API documentation
- Business: See revenue projections
- Security: Review security section

---

**Built for Google AI Partner Catalyst Hackathon 2025**
