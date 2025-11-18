# Vibe Markets - Quick Start Guide

## What We've Built So Far ✅

1. **Branch Created:** `hackathon-vibe-markets`
2. **Architecture Designed:** See `HACKATHON_VIBE_MARKETS.md`
3. **43-Day Plan:** See `IMPLEMENTATION_PLAN.md`
4. **Repository Structure:** Defined and documented

## Your Next Steps (THIS WEEK)

### Step 1: Set Up Confluent Cloud (30 minutes)

1. Go to https://confluent.cloud/signup
2. Sign up for free trial (no credit card for basic tier)
3. Create a new cluster:
   - Choose "Basic" cluster (free)
   - Choose closest region
   - Name it "vibe-markets-dev"

4. Create API credentials:
   - Go to "Data Integration" → "API Keys"
   - Click "Add Key"
   - Select "Global Access"
   - **SAVE THESE CREDENTIALS SECURELY**

5. Create 5 Kafka topics:
   ```
   Topic Name: bet-orders
   Partitions: 3
   Retention: 7 days

   Topic Name: market-updates
   Partitions: 3
   Retention: 7 days

   Topic Name: social-feed-events
   Partitions: 3
   Retention: 7 days

   Topic Name: user-activity
   Partitions: 3
   Retention: 30 days

   Topic Name: resolution-events
   Partitions: 1
   Retention: 7 days
   ```

### Step 2: Set Up Google Cloud (30 minutes)

1. Go to https://cloud.google.com/
2. Create new project: "vibe-markets"
3. Activate $300 free credits
4. Enable required APIs:
   ```bash
   gcloud config set project vibe-markets
   gcloud services enable aiplatform.googleapis.com
   gcloud services enable run.googleapis.com
   gcloud services enable cloudbuild.googleapis.com
   ```

5. Get Gemini API Key:
   - Go to https://ai.google.dev/
   - Click "Get API Key"
   - Create key for "vibe-markets" project

6. Create service account:
   ```bash
   gcloud iam service-accounts create vibe-markets-sa \
     --display-name="Vibe Markets Service Account"

   gcloud projects add-iam-policy-binding vibe-markets \
     --member="serviceAccount:vibe-markets-sa@vibe-markets.iam.gserviceaccount.com" \
     --role="roles/aiplatform.user"

   gcloud iam service-accounts keys create ~/vibe-markets-key.json \
     --iam-account=vibe-markets-sa@vibe-markets.iam.gserviceaccount.com
   ```

### Step 3: Configure Environment (10 minutes)

Create `.env` file in project root:
```bash
# Confluent Cloud
CONFLUENT_BOOTSTRAP_SERVERS=pkc-xxxxx.us-east-1.aws.confluent.cloud:9092
CONFLUENT_API_KEY=YOUR_CONFLUENT_KEY
CONFLUENT_API_SECRET=YOUR_CONFLUENT_SECRET

# Google Cloud
GOOGLE_CLOUD_PROJECT=vibe-markets
GOOGLE_APPLICATION_CREDENTIALS=/path/to/vibe-markets-key.json
GEMINI_API_KEY=YOUR_GEMINI_KEY

# Application
PORT=8080
ENV=development
```

### Step 4: Start Building (REST OF WEEK)

Follow the `IMPLEMENTATION_PLAN.md` day-by-day guide.

**This week's goal:** Get Market Service running with Kafka integration.

## Quick Commands

### Development
```bash
# Switch to hackathon branch
git checkout hackathon-vibe-markets

# Pull latest changes
git pull origin hackathon-vibe-markets

# Run a service locally (once built)
cd src/marketservice
go run .

# Test Kafka connection
# (we'll create test script next)
```

### Deployment (Week 4)
```bash
# Build and deploy to Cloud Run
gcloud run deploy marketservice \
  --source=./src/marketservice \
  --region=us-central1 \
  --allow-unauthenticated
```

## Project Structure

```
microservices-demo/
├── HACKATHON_VIBE_MARKETS.md   ← Architecture & requirements
├── IMPLEMENTATION_PLAN.md       ← Week-by-week plan
├── QUICKSTART.md               ← This file
├── src/
│   ├── marketservice/          ← TODO: Create this week
│   ├── bettingservice/         ← TODO: Week 2
│   ├── socialfeedservice/      ← TODO: Week 3
│   ├── frontend/               ← Adapt existing
│   └── ...
├── protos/
│   ├── demo.proto             ← Existing
│   └── markets.proto          ← TODO: Create this week
└── .env                       ← TODO: Create now
```

## This Week's Milestones

By end of Week 1 (Day 7), you should have:

- [x] Hackathon branch created
- [x] Architecture documented
- [x] Implementation plan created
- [ ] Confluent Cloud configured
- [ ] Google Cloud configured
- [ ] Environment variables set
- [ ] Market Service created
- [ ] 10 demo markets loaded
- [ ] Kafka producer working
- [ ] Can publish market update events

## Daily Time Commitment

**Minimum:** 2-3 hours/day
**Recommended:** 4-6 hours/day
**Final week:** 8-12 hours/day

## Getting Help

### If you get stuck:

1. **Check docs:**
   - Confluent: https://docs.confluent.io/cloud/
   - Vertex AI: https://cloud.google.com/vertex-ai/docs
   - Go gRPC: https://grpc.io/docs/languages/go/

2. **Community help:**
   - Hackathon Discord (if available)
   - Stack Overflow
   - Confluent Community Slack

3. **Ask me:** Continue this conversation for technical help

## Critical Dates

- **Today:** Environment setup
- **Day 7:** Market service working
- **Day 14:** AI features integrated
- **Day 21:** Basic frontend working
- **Day 28:** Deployed to Cloud
- **Day 35:** Ready for video
- **Day 43:** SUBMIT (Dec 31, 11am GMT-11)

## Emergency Simplifications

If you're running behind, you can simplify:

1. **Week 1-2 behind:** Use local Kafka (Docker) instead of Confluent Cloud
2. **Week 3 behind:** Skip social feed, focus on betting only
3. **Week 4 behind:** Use simple rules instead of Vertex AI
4. **Week 5 behind:** Use basic HTML instead of polished UI
5. **Week 6 behind:** Shorter video, less documentation

## What Success Looks Like

**Minimum Viable Submission:**
- 5 markets viewable ✓
- Can place 1 bet ✓
- Bet publishes to Kafka ✓
- 1 AI feature working (Gemini OR Vertex) ✓
- Hosted on Cloud Run ✓
- 3-minute video ✓

**Winning Submission:**
- 10 markets, polished UI
- Real-time betting with live odds
- Both Gemini AND Vertex AI
- Social feed showing activity
- Fraud detection demo
- Professional video with great story

## Resources Bookmarked

### Confluent
- [Cloud Quickstart](https://docs.confluent.io/cloud/current/get-started/index.html)
- [Go Client Docs](https://docs.confluent.io/kafka-clients/go/current/overview.html)
- [ksqlDB Guide](https://docs.confluent.io/platform/current/ksqldb/)

### Google Cloud AI
- [Vertex AI Tutorials](https://cloud.google.com/vertex-ai/docs/tutorials)
- [Gemini API Guide](https://ai.google.dev/tutorials)
- [Cloud Run Quickstart](https://cloud.google.com/run/docs/quickstarts)

### Inspiration
- [Polymarket](https://polymarket.com/) - Study their UX
- [Manifold Markets](https://manifold.markets/) - Study their social features

## Keeping Momentum

### Daily Ritual:
1. **Morning (5 min):** Review today's tasks in IMPLEMENTATION_PLAN.md
2. **Work Session:** 2-6 hours of focused building
3. **Evening (5 min):** Commit code, update todos, plan tomorrow
4. **Weekly (30 min):** Review progress, adjust plan if needed

### Stay Motivated:
- **Visualize:** $25k prize money, demo day success
- **Small wins:** Celebrate each service that works
- **Progress:** Track completed tasks daily
- **Community:** Share progress with others
- **Remember:** Done is better than perfect

## Let's Build This! 🚀

You have everything you need to start. The architecture is solid, the plan is clear, and the path to submission is mapped out.

**Your immediate next action:**
1. Set up Confluent Cloud (30 min)
2. Set up Google Cloud (30 min)
3. Come back here and tell me you're ready for Day 3

Questions? Ask me anything. Let's win this hackathon! 💪
