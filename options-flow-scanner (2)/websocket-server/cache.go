package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// ================================================
// DRAGONFLYDB CACHE LAYER
// ================================================
// High-performance caching for hot trading data
// 0.2-2ms query times
// TTL-based expiration
// ================================================

var (
	dragonflyClient *redis.Client
	cacheCtx        = context.Background()
)

// TTL constants (in seconds)
const (
	TTL_HOT_DATA = 3600  // 1 hour for general trading data
	TTL_PRICE    = 300   // 5 minutes for price data
	TTL_GEX      = 1800  // 30 minutes for GEX calculations
	TTL_FLOW     = 3600  // 1 hour for options flow
	TTL_GREEKS   = 600   // 10 minutes for Greeks
	TTL_METRICS  = 900   // 15 minutes for market metrics
)

// ================================================
// INITIALIZATION
// ================================================

// InitDragonfly initializes the DragonflyDB connection
func InitDragonfly() error {
	dragonflyURL := getEnv("DRAGONFLY_URL", "")
	if dragonflyURL == "" {
		return fmt.Errorf("DRAGONFLY_URL environment variable not set")
	}

	opt, err := redis.ParseURL(dragonflyURL)
	if err != nil {
		return fmt.Errorf("failed to parse DRAGONFLY_URL: %w", err)
	}

	// Configure connection pool for high performance
	opt.PoolSize = 100
	opt.MinIdleConns = 10
	opt.MaxRetries = 3
	opt.DialTimeout = 5 * time.Second
	opt.ReadTimeout = 3 * time.Second
	opt.WriteTimeout = 3 * time.Second

	dragonflyClient = redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(cacheCtx, 5*time.Second)
	defer cancel()

	_, err = dragonflyClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to ping DragonflyDB: %w", err)
	}

	log.Println("✅ Connected to DragonflyDB successfully")
	return nil
}

// CloseDragonfly closes the DragonflyDB connection
func CloseDragonfly() error {
	if dragonflyClient != nil {
		return dragonflyClient.Close()
	}
	return nil
}

// ================================================
// CACHE OPERATIONS - SIMPLE KEY/VALUE
// ================================================

// CacheSet stores a value with TTL
func CacheSet(key string, value interface{}, ttl time.Duration) error {
	if dragonflyClient == nil {
		return fmt.Errorf("DragonflyDB not initialized")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	ctx, cancel := context.WithTimeout(cacheCtx, 2*time.Second)
	defer cancel()

	return dragonflyClient.Set(ctx, key, data, ttl).Err()
}

// CacheGet retrieves a value from cache
func CacheGet(key string) ([]byte, error) {
	if dragonflyClient == nil {
		return nil, fmt.Errorf("DragonflyDB not initialized")
	}

	ctx, cancel := context.WithTimeout(cacheCtx, 2*time.Second)
	defer cancel()

	return dragonflyClient.Get(ctx, key).Bytes()
}

// CacheDel deletes a key from cache
func CacheDel(key string) error {
	if dragonflyClient == nil {
		return fmt.Errorf("DragonflyDB not initialized")
	}

	ctx, cancel := context.WithTimeout(cacheCtx, 2*time.Second)
	defer cancel()

	return dragonflyClient.Del(ctx, key).Err()
}

// ================================================
// CACHE OPERATIONS - HASH (for GEX data)
// ================================================

// CacheHSet sets a hash field
func CacheHSet(key, field string, value interface{}) error {
	if dragonflyClient == nil {
		return fmt.Errorf("DragonflyDB not initialized")
	}

	ctx, cancel := context.WithTimeout(cacheCtx, 2*time.Second)
	defer cancel()

	return dragonflyClient.HSet(ctx, key, field, value).Err()
}

// CacheHMSet sets multiple hash fields
func CacheHMSet(key string, values map[string]interface{}, ttl time.Duration) error {
	if dragonflyClient == nil {
		return fmt.Errorf("DragonflyDB not initialized")
	}

	ctx, cancel := context.WithTimeout(cacheCtx, 2*time.Second)
	defer cancel()

	// Set hash fields
	if err := dragonflyClient.HSet(ctx, key, values).Err(); err != nil {
		return err
	}

	// Set TTL
	return dragonflyClient.Expire(ctx, key, ttl).Err()
}

// CacheHGetAll gets all hash fields
func CacheHGetAll(key string) (map[string]string, error) {
	if dragonflyClient == nil {
		return nil, fmt.Errorf("DragonflyDB not initialized")
	}

	ctx, cancel := context.WithTimeout(cacheCtx, 2*time.Second)
	defer cancel()

	return dragonflyClient.HGetAll(ctx, key).Result()
}

// ================================================
// CACHE OPERATIONS - LIST (for options flow)
// ================================================

// CacheLPush pushes value to list (left/head)
func CacheLPush(key string, value interface{}, maxLength int, ttl time.Duration) error {
	if dragonflyClient == nil {
		return fmt.Errorf("DragonflyDB not initialized")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	ctx, cancel := context.WithTimeout(cacheCtx, 2*time.Second)
	defer cancel()

	// Use pipeline for atomic operations
	pipe := dragonflyClient.Pipeline()
	pipe.LPush(ctx, key, data)
	pipe.LTrim(ctx, key, 0, int64(maxLength-1))
	pipe.Expire(ctx, key, ttl)

	_, err = pipe.Exec(ctx)
	return err
}

// CacheLRange gets list range
func CacheLRange(key string, start, stop int64) ([]string, error) {
	if dragonflyClient == nil {
		return nil, fmt.Errorf("DragonflyDB not initialized")
	}

	ctx, cancel := context.WithTimeout(cacheCtx, 2*time.Second)
	defer cancel()

	return dragonflyClient.LRange(ctx, key, start, stop).Result()
}

// ================================================
// TRADING DATA CACHE FUNCTIONS
// ================================================

// CachePrice caches the latest price for a symbol
func CachePrice(symbol string, price float64) error {
	key := fmt.Sprintf("price:%s", symbol)
	return CacheSet(key, price, TTL_PRICE*time.Second)
}

// CacheGEX caches GEX data for a symbol/strike
func CacheGEX(symbol string, gexData map[string]interface{}) error {
	key := fmt.Sprintf("gex:%s", symbol)
	return CacheHMSet(key, gexData, TTL_GEX*time.Second)
}

// CacheFlow caches options flow data
func CacheFlow(symbol string, flowData interface{}) error {
	key := fmt.Sprintf("flow:%s", symbol)
	return CacheLPush(key, flowData, 100, TTL_FLOW*time.Second)
}

// CacheGreeks caches Greeks data for a symbol
func CacheGreeks(symbol string, greeks map[string]interface{}) error {
	key := fmt.Sprintf("greeks:%s", symbol)
	return CacheSet(key, greeks, TTL_GREEKS*time.Second)
}

// CacheTrade caches trade data (for quick recent lookups)
func CacheTrade(symbol string, trade interface{}) error {
	key := fmt.Sprintf("trade:%s", symbol)
	return CacheSet(key, trade, TTL_PRICE*time.Second)
}

// CacheQuote caches quote data (bid/ask)
func CacheQuote(symbol string, quote interface{}) error {
	key := fmt.Sprintf("quote:%s", symbol)
	return CacheSet(key, quote, TTL_PRICE*time.Second)
}

// CacheAggregate caches 1-minute candle data
func CacheAggregate(symbol string, aggregate interface{}) error {
	key := fmt.Sprintf("agg:1min:%s", symbol)
	return CacheSet(key, aggregate, TTL_METRICS*time.Second)
}

// ================================================
// BATCH OPERATIONS (for performance)
// ================================================

// CacheBatchSet sets multiple keys atomically
func CacheBatchSet(keyValues map[string]interface{}, ttl time.Duration) error {
	if dragonflyClient == nil {
		return fmt.Errorf("DragonflyDB not initialized")
	}

	ctx, cancel := context.WithTimeout(cacheCtx, 5*time.Second)
	defer cancel()

	pipe := dragonflyClient.Pipeline()
	
	for key, value := range keyValues {
		data, err := json.Marshal(value)
		if err != nil {
			log.Printf("❌ Failed to marshal value for key %s: %v", key, err)
			continue
		}
		pipe.Set(ctx, key, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// ================================================
// CACHE STATISTICS
// ================================================

// GetCacheStats returns cache statistics
func GetCacheStats() (map[string]interface{}, error) {
	if dragonflyClient == nil {
		return nil, fmt.Errorf("DragonflyDB not initialized")
	}

	ctx, cancel := context.WithTimeout(cacheCtx, 5*time.Second)
	defer cancel()

	// Get memory stats
	info, err := dragonflyClient.Info(ctx, "memory").Result()
	if err != nil {
		return nil, err
	}

	// Get key count
	dbSize, err := dragonflyClient.DBSize(ctx).Result()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"connected": true,
		"keys":      dbSize,
		"info":      info,
	}, nil
}

// ================================================
// HEALTH CHECK
// ================================================

// CheckCacheHealth performs a health check on DragonflyDB
func CheckCacheHealth() error {
	if dragonflyClient == nil {
		return fmt.Errorf("DragonflyDB not initialized")
	}

	ctx, cancel := context.WithTimeout(cacheCtx, 3*time.Second)
	defer cancel()

	_, err := dragonflyClient.Ping(ctx).Result()
	return err
}

// ================================================
// UTILITY FUNCTIONS
// ================================================

// IsCache Ready returns true if cache is initialized and connected
func IsCacheReady() bool {
	if dragonflyClient == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(cacheCtx, 1*time.Second)
	defer cancel()

	_, err := dragonflyClient.Ping(ctx).Result()
	return err == nil
}

// FlushCache flushes all cache data (use with caution!)
func FlushCache() error {
	if dragonflyClient == nil {
		return fmt.Errorf("DragonflyDB not initialized")
	}

	ctx, cancel := context.WithTimeout(cacheCtx, 10*time.Second)
	defer cancel()

	return dragonflyClient.FlushAll(ctx).Err()
}
