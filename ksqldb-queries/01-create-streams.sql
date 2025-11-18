-- ====================================================================
-- Vibe Markets - ksqlDB Streams and Tables
-- ====================================================================
-- These queries create real-time streams and tables for:
-- - Leaderboards
-- - Market statistics
-- - User activity tracking
-- - Fraud detection
-- ====================================================================

-- ====================================================================
-- 1. CREATE STREAMS FROM KAFKA TOPICS
-- ====================================================================

-- Stream for bet orders
CREATE STREAM bet_orders_stream (
    user_id VARCHAR,
    market_id VARCHAR,
    side VARCHAR,
    amount DOUBLE,
    timestamp BIGINT,
    event_type VARCHAR
) WITH (
    KAFKA_TOPIC='bet-orders',
    VALUE_FORMAT='JSON',
    TIMESTAMP='timestamp'
);

-- Stream for market updates
CREATE STREAM market_updates_stream (
    market_id VARCHAR,
    yes_price DOUBLE,
    no_price DOUBLE,
    volume DOUBLE,
    timestamp BIGINT,
    event_type VARCHAR
) WITH (
    KAFKA_TOPIC='market-updates',
    VALUE_FORMAT='JSON',
    TIMESTAMP='timestamp'
);

-- Stream for social feed events
CREATE STREAM social_feed_stream (
    event_type VARCHAR,
    user_id VARCHAR,
    market_id VARCHAR,
    data MAP<VARCHAR, VARCHAR>,
    timestamp BIGINT
) WITH (
    KAFKA_TOPIC='social-feed-events',
    VALUE_FORMAT='JSON',
    TIMESTAMP='timestamp'
);

-- Stream for user activity (ML training)
CREATE STREAM user_activity_stream (
    user_id VARCHAR,
    action VARCHAR,
    market_id VARCHAR,
    context MAP<VARCHAR, VARCHAR>,
    timestamp BIGINT
) WITH (
    KAFKA_TOPIC='user-activity',
    VALUE_FORMAT='JSON',
    TIMESTAMP='timestamp'
);

-- ====================================================================
-- 2. LEADERBOARD TABLES
-- ====================================================================

-- User statistics table (aggregated from bets)
CREATE TABLE user_stats AS
SELECT
    user_id,
    COUNT(*) AS total_bets,
    SUM(amount) AS total_volume,
    AVG(amount) AS avg_bet_size,
    MIN(amount) AS min_bet,
    MAX(amount) AS max_bet,
    LATEST_BY_OFFSET(timestamp) AS last_bet_time
FROM bet_orders_stream
GROUP BY user_id
EMIT CHANGES;

-- Top traders by volume (last 24 hours)
CREATE TABLE top_traders_24h AS
SELECT
    user_id,
    COUNT(*) AS bets_24h,
    SUM(amount) AS volume_24h
FROM bet_orders_stream
WINDOW TUMBLING (SIZE 24 HOURS)
GROUP BY user_id
HAVING SUM(amount) > 100
EMIT CHANGES;

-- ====================================================================
-- 3. MARKET STATISTICS TABLES
-- ====================================================================

-- Market statistics (all-time)
CREATE TABLE market_stats AS
SELECT
    market_id,
    COUNT(*) AS total_bets,
    SUM(amount) AS total_volume,
    AVG(amount) AS avg_bet_size,
    LATEST_BY_OFFSET(yes_price) AS current_yes_price,
    LATEST_BY_OFFSET(no_price) AS current_no_price,
    LATEST_BY_OFFSET(volume) AS current_volume
FROM bet_orders_stream
GROUP BY market_id
EMIT CHANGES;

-- Market activity in last hour
CREATE TABLE market_activity_1h AS
SELECT
    market_id,
    COUNT(*) AS bets_last_hour,
    SUM(amount) AS volume_last_hour
FROM bet_orders_stream
WINDOW TUMBLING (SIZE 1 HOUR)
GROUP BY market_id
EMIT CHANGES;

-- Hot markets (most active in last 10 minutes)
CREATE TABLE hot_markets AS
SELECT
    market_id,
    COUNT(*) AS bets_10min,
    SUM(amount) AS volume_10min
FROM bet_orders_stream
WINDOW TUMBLING (SIZE 10 MINUTES)
GROUP BY market_id
HAVING COUNT(*) > 5
EMIT CHANGES;

-- ====================================================================
-- 4. BETTING PATTERNS (for fraud detection)
-- ====================================================================

-- Users with rapid sequential bets (potential wash trading)
CREATE STREAM rapid_bets AS
SELECT
    user_id,
    market_id,
    COUNT(*) AS bet_count,
    WINDOWSTART AS window_start,
    WINDOWEND AS window_end
FROM bet_orders_stream
WINDOW TUMBLING (SIZE 1 MINUTE)
GROUP BY user_id, market_id
HAVING COUNT(*) > 10
EMIT CHANGES;

-- Large bets (> $1000) for monitoring
CREATE STREAM large_bets AS
SELECT
    user_id,
    market_id,
    side,
    amount,
    timestamp
FROM bet_orders_stream
WHERE amount > 1000
EMIT CHANGES;

-- Suspicious patterns: alternating YES/NO bets
CREATE TABLE bet_side_patterns AS
SELECT
    user_id,
    market_id,
    COLLECT_LIST(side) AS recent_sides,
    COUNT(*) AS recent_bet_count
FROM bet_orders_stream
WINDOW TUMBLING (SIZE 5 MINUTES)
GROUP BY user_id, market_id
EMIT CHANGES;

-- ====================================================================
-- 5. REAL-TIME AGGREGATIONS
-- ====================================================================

-- Total platform statistics
CREATE TABLE platform_stats AS
SELECT
    'platform' AS key,
    COUNT(DISTINCT user_id) AS total_users,
    COUNT(DISTINCT market_id) AS total_markets,
    COUNT(*) AS total_bets,
    SUM(amount) AS total_volume
FROM bet_orders_stream
GROUP BY 'platform'
EMIT CHANGES;

-- Betting velocity (bets per minute)
CREATE TABLE betting_velocity AS
SELECT
    TIMESTAMPTOSTRING(WINDOWSTART, 'yyyy-MM-dd HH:mm') AS minute,
    COUNT(*) AS bets_per_minute,
    SUM(amount) AS volume_per_minute
FROM bet_orders_stream
WINDOW TUMBLING (SIZE 1 MINUTE)
GROUP BY 1
EMIT CHANGES;

-- Market sentiment (YES vs NO betting)
CREATE TABLE market_sentiment AS
SELECT
    market_id,
    SUM(CASE WHEN side = 'YES' THEN amount ELSE 0 END) AS yes_volume,
    SUM(CASE WHEN side = 'NO' THEN amount ELSE 0 END) AS no_volume,
    COUNT(CASE WHEN side = 'YES' THEN 1 END) AS yes_bets,
    COUNT(CASE WHEN side = 'NO' THEN 1 END) AS no_bets
FROM bet_orders_stream
GROUP BY market_id
EMIT CHANGES;

-- ====================================================================
-- 6. USER ENGAGEMENT METRICS
-- ====================================================================

-- Active users (bet in last hour)
CREATE TABLE active_users_1h AS
SELECT
    user_id,
    COUNT(*) AS bets_last_hour,
    SUM(amount) AS volume_last_hour,
    LATEST_BY_OFFSET(market_id) AS last_market
FROM bet_orders_stream
WINDOW TUMBLING (SIZE 1 HOUR)
GROUP BY user_id
EMIT CHANGES;

-- User market diversity (how many different markets)
CREATE TABLE user_market_diversity AS
SELECT
    user_id,
    COUNT(DISTINCT market_id) AS markets_traded,
    COUNT(*) AS total_bets
FROM bet_orders_stream
GROUP BY user_id
EMIT CHANGES;

-- ====================================================================
-- 7. MATERIALIZED VIEWS FOR LEADERBOARD
-- ====================================================================

-- Top 10 traders by volume (for leaderboard)
CREATE TABLE leaderboard_volume AS
SELECT
    user_id,
    SUM(amount) AS total_volume,
    COUNT(*) AS total_bets
FROM bet_orders_stream
GROUP BY user_id
EMIT CHANGES;

-- Most active traders by bet count
CREATE TABLE leaderboard_activity AS
SELECT
    user_id,
    COUNT(*) AS total_bets,
    SUM(amount) AS total_volume,
    COUNT(DISTINCT market_id) AS markets_traded
FROM bet_orders_stream
GROUP BY user_id
EMIT CHANGES;

-- ====================================================================
-- END OF ksqlDB SETUP
-- ====================================================================

-- To query these tables in real-time:
-- SELECT * FROM user_stats EMIT CHANGES;
-- SELECT * FROM market_stats WHERE market_id = 'btc-150k-jun2025' EMIT CHANGES;
-- SELECT * FROM hot_markets EMIT CHANGES;
-- SELECT * FROM leaderboard_volume ORDER BY total_volume DESC LIMIT 10;
