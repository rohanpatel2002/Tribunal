package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// CacheLayer provides interface for caching implementations
type CacheLayer interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	FlushAll(ctx context.Context) error
}

// RedisCache implements CacheLayer with Redis
type RedisCache struct {
	// In production, this would be *redis.Client
	// For MVP, we'll keep in-memory fallback
	localCache map[string]cacheEntry
}

type cacheEntry struct {
	value  string
	expiry time.Time
}

// NewRedisCache initializes Redis cache
func NewRedisCache(redisURL string) (*RedisCache, error) {
	// In production:
	// client := redis.NewClient(&redis.Options{URL: redisURL})
	// For MVP, use in-memory
	return &RedisCache{
		localCache: make(map[string]cacheEntry),
	}, nil
}

// Get retrieves value from cache
func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
	entry, exists := rc.localCache[key]
	if !exists {
		return "", fmt.Errorf("cache miss: key not found")
	}

	if time.Now().After(entry.expiry) {
		delete(rc.localCache, key)
		return "", fmt.Errorf("cache miss: key expired")
	}

	return entry.value, nil
}

// Set stores value in cache with TTL
func (rc *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	rc.localCache[key] = cacheEntry{
		value:  value,
		expiry: time.Now().Add(ttl),
	}
	return nil
}

// Delete removes key from cache
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	delete(rc.localCache, key)
	return nil
}

// Exists checks if key exists in cache
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	entry, exists := rc.localCache[key]
	if !exists {
		return false, nil
	}

	if time.Now().After(entry.expiry) {
		delete(rc.localCache, key)
		return false, nil
	}

	return true, nil
}

// FlushAll clears entire cache
func (rc *RedisCache) FlushAll(ctx context.Context) error {
	rc.localCache = make(map[string]cacheEntry)
	return nil
}

// CacheKey generates consistent cache key
func CacheKey(prefix, repo, id string) string {
	return fmt.Sprintf("tribunal:%s:%s:%s", prefix, repo, id)
}

// GetCachedAuditSummary retrieves cached audit summary
func GetCachedAuditSummary(ctx context.Context, cache CacheLayer, repo string) (*AuditSummary, error) {
	if cache == nil {
		return nil, fmt.Errorf("cache not available")
	}

	key := CacheKey("summary", repo, "")
	data, err := cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var summary AuditSummary
	if err := json.Unmarshal([]byte(data), &summary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return &summary, nil
}

// SetCachedAuditSummary caches audit summary with 5-minute TTL
func SetCachedAuditSummary(ctx context.Context, cache CacheLayer, repo string, summary *AuditSummary) error {
	if cache == nil {
		return nil
	}

	data, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	key := CacheKey("summary", repo, "")
	ttl := 5 * time.Minute

	return cache.Set(ctx, key, string(data), ttl)
}

// GetCachedAnalyses retrieves cached PR analyses
func GetCachedAnalyses(ctx context.Context, cache CacheLayer, repo string, limit int) ([]PRAnalysisRecord, error) {
	if cache == nil {
		return nil, fmt.Errorf("cache not available")
	}

	key := CacheKey("analyses", repo, fmt.Sprintf("limit_%d", limit))
	data, err := cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var analyses []PRAnalysisRecord
	if err := json.Unmarshal([]byte(data), &analyses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached analyses: %w", err)
	}

	return analyses, nil
}

// SetCachedAnalyses caches PR analyses with 3-minute TTL
func SetCachedAnalyses(ctx context.Context, cache CacheLayer, repo string, limit int, analyses []PRAnalysisRecord) error {
	if cache == nil {
		return nil
	}

	data, err := json.Marshal(analyses)
	if err != nil {
		return fmt.Errorf("failed to marshal analyses: %w", err)
	}

	key := CacheKey("analyses", repo, fmt.Sprintf("limit_%d", limit))
	ttl := 3 * time.Minute

	return cache.Set(ctx, key, string(data), ttl)
}

// InvalidateCache removes all cached data for a repository
func InvalidateCache(ctx context.Context, cache CacheLayer, repo string) error {
	if cache == nil {
		return nil
	}

	// In production, use Redis KEYS pattern matching
	// For MVP, we'd need to track keys manually
	slog.Info("cache invalidated", "repo", repo)
	return nil
}

// CacheMiddleware adds caching to handlers
func CacheMiddleware(cache CacheLayer, cacheTTL time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if cache == nil || r.Method != http.MethodGet {
				next(w, r)
				return
			}

			cacheKey := fmt.Sprintf("%s:%s", r.URL.Path, r.URL.RawQuery)
			cached, err := cache.Get(r.Context(), cacheKey)
			if err == nil {
				w.Header().Set("X-Cache", "HIT")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, cached)
				return
			}

			w.Header().Set("X-Cache", "MISS")
			next(w, r)
		}
	}
}
