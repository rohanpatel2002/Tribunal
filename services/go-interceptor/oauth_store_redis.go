package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultGitHubSessionTTL = 30 * 24 * time.Hour

type RedisGitHubOAuthOptions struct {
	SessionTTL    time.Duration
	OAuthStateTTL time.Duration
}

type RedisStoreMetrics struct {
	Operations           uint64     `json:"operations"`
	Errors               uint64     `json:"errors"`
	LastError            string     `json:"lastError,omitempty"`
	LastErrorAt          *time.Time `json:"lastErrorAt,omitempty"`
	LastPingAt           *time.Time `json:"lastPingAt,omitempty"`
	LastPingLatencyMs    int64      `json:"lastPingLatencyMs"`
	LastPingOK           bool       `json:"lastPingOk"`
	SessionTTLSeconds    int64      `json:"sessionTtlSeconds"`
	OAuthStateTTLSeconds int64      `json:"oauthStateTtlSeconds"`
}

type RedisHealthChecker interface {
	Ping(ctx context.Context) (time.Duration, error)
}

type RedisMetricsProvider interface {
	RedisMetrics() RedisStoreMetrics
}

type redisMetrics struct {
	mu                sync.Mutex
	operations        uint64
	errors            uint64
	lastError         string
	lastErrorAt       *time.Time
	lastPingAt        *time.Time
	lastPingLatencyMs int64
	lastPingOK        bool
}

func (m *redisMetrics) recordOp(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.operations++
	if err != nil {
		now := time.Now().UTC()
		m.errors++
		m.lastError = err.Error()
		m.lastErrorAt = &now
	}
}

func (m *redisMetrics) recordPing(latency time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	m.operations++
	m.lastPingAt = &now
	m.lastPingLatencyMs = latency.Milliseconds()
	m.lastPingOK = err == nil
	if err != nil {
		m.errors++
		m.lastError = err.Error()
		m.lastErrorAt = &now
	}
}

func (m *redisMetrics) snapshot(sessionTTL time.Duration, oauthStateTTL time.Duration) RedisStoreMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	return RedisStoreMetrics{
		Operations:           m.operations,
		Errors:               m.errors,
		LastError:            m.lastError,
		LastErrorAt:          m.lastErrorAt,
		LastPingAt:           m.lastPingAt,
		LastPingLatencyMs:    m.lastPingLatencyMs,
		LastPingOK:           m.lastPingOK,
		SessionTTLSeconds:    int64(sessionTTL.Seconds()),
		OAuthStateTTLSeconds: int64(oauthStateTTL.Seconds()),
	}
}

// RedisGitHubOAuthStore stores OAuth state and GitHub connections in Redis.
type RedisGitHubOAuthStore struct {
	client        *redis.Client
	metrics       *redisMetrics
	sessionTTL    time.Duration
	oauthStateTTL time.Duration
}

func NewRedisGitHubOAuthStore(redisURL string, options RedisGitHubOAuthOptions) (*RedisGitHubOAuthStore, error) {
	parsedOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}
	client := redis.NewClient(parsedOptions)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	sessionTTL := options.SessionTTL
	if sessionTTL <= 0 {
		sessionTTL = defaultGitHubSessionTTL
	}

	oauthStateTTL := options.OAuthStateTTL
	if oauthStateTTL < 0 {
		oauthStateTTL = 0
	}

	return &RedisGitHubOAuthStore{
		client:        client,
		metrics:       &redisMetrics{},
		sessionTTL:    sessionTTL,
		oauthStateTTL: oauthStateTTL,
	}, nil
}

func (s *RedisGitHubOAuthStore) SaveOAuthState(ctx context.Context, state string, entry githubOAuthState) error {
	key := redisOAuthStateKey(state)
	value, err := json.Marshal(entry)
	if err != nil {
		s.metrics.recordOp(err)
		return fmt.Errorf("failed to marshal oauth state: %w", err)
	}

	expiresIn := time.Until(entry.ExpiresAt)
	if expiresIn <= 0 {
		expiresIn = 10 * time.Minute
	}
	if s.oauthStateTTL > 0 {
		if expiresIn <= 0 || s.oauthStateTTL < expiresIn {
			expiresIn = s.oauthStateTTL
		}
	}

	if err := s.client.Set(ctx, key, value, expiresIn).Err(); err != nil {
		s.metrics.recordOp(err)
		return fmt.Errorf("failed to save oauth state: %w", err)
	}
	s.metrics.recordOp(nil)
	return nil
}

func (s *RedisGitHubOAuthStore) ConsumeOAuthState(ctx context.Context, state string) (githubOAuthState, bool, error) {
	key := redisOAuthStateKey(state)
	result, err := s.client.GetDel(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			s.metrics.recordOp(nil)
			return githubOAuthState{}, false, nil
		}
		s.metrics.recordOp(err)
		return githubOAuthState{}, false, fmt.Errorf("failed to consume oauth state: %w", err)
	}

	var entry githubOAuthState
	if err := json.Unmarshal([]byte(result), &entry); err != nil {
		s.metrics.recordOp(err)
		return githubOAuthState{}, false, fmt.Errorf("failed to decode oauth state: %w", err)
	}
	s.metrics.recordOp(nil)
	return entry, true, nil
}

func (s *RedisGitHubOAuthStore) CleanupExpiredOAuthStates(ctx context.Context, _ time.Time) error {
	// Redis handles TTL expiration automatically.
	return nil
}

func (s *RedisGitHubOAuthStore) SaveGitHubConnection(ctx context.Context, sessionID string, conn *githubConnection) error {
	if conn == nil {
		s.metrics.recordOp(fmt.Errorf("github connection is required"))
		return fmt.Errorf("github connection is required")
	}

	key := redisGitHubSessionKey(sessionID)
	value, err := json.Marshal(conn)
	if err != nil {
		s.metrics.recordOp(err)
		return fmt.Errorf("failed to marshal github connection: %w", err)
	}

	if err := s.client.Set(ctx, key, value, s.sessionTTL).Err(); err != nil {
		s.metrics.recordOp(err)
		return fmt.Errorf("failed to save github connection: %w", err)
	}
	s.metrics.recordOp(nil)
	return nil
}

func (s *RedisGitHubOAuthStore) GetGitHubConnection(ctx context.Context, sessionID string) (*githubConnection, error) {
	key := redisGitHubSessionKey(sessionID)
	result, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			s.metrics.recordOp(nil)
			return nil, nil
		}
		s.metrics.recordOp(err)
		return nil, fmt.Errorf("failed to fetch github connection: %w", err)
	}

	var conn githubConnection
	if err := json.Unmarshal([]byte(result), &conn); err != nil {
		s.metrics.recordOp(err)
		return nil, fmt.Errorf("failed to decode github connection: %w", err)
	}
	s.metrics.recordOp(nil)
	return &conn, nil
}

func (s *RedisGitHubOAuthStore) DeleteGitHubConnection(ctx context.Context, sessionID string) error {
	key := redisGitHubSessionKey(sessionID)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		s.metrics.recordOp(err)
		return fmt.Errorf("failed to delete github connection: %w", err)
	}
	s.metrics.recordOp(nil)
	return nil
}

func (s *RedisGitHubOAuthStore) Ping(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	err := s.client.Ping(ctx).Err()
	latency := time.Since(start)
	s.metrics.recordPing(latency, err)
	if err != nil {
		return latency, fmt.Errorf("failed to ping redis: %w", err)
	}
	return latency, nil
}

func (s *RedisGitHubOAuthStore) RedisMetrics() RedisStoreMetrics {
	return s.metrics.snapshot(s.sessionTTL, s.oauthStateTTL)
}

func redisOAuthStateKey(state string) string {
	return fmt.Sprintf("tribunal:oauth_state:%s", state)
}

func redisGitHubSessionKey(sessionID string) string {
	return fmt.Sprintf("tribunal:github_session:%s", sessionID)
}
