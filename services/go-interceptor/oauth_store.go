package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// GitHubOAuthStore persists temporary OAuth state and connected-account metadata.
// Postgres-backed implementations allow multi-instance deployment without losing sessions.
type GitHubOAuthStore interface {
	SaveOAuthState(ctx context.Context, state string, entry githubOAuthState) error
	ConsumeOAuthState(ctx context.Context, state string) (githubOAuthState, bool, error)
	CleanupExpiredOAuthStates(ctx context.Context, now time.Time) error
	SaveGitHubConnection(ctx context.Context, sessionID string, conn *githubConnection) error
	GetGitHubConnection(ctx context.Context, sessionID string) (*githubConnection, error)
	DeleteGitHubConnection(ctx context.Context, sessionID string) error
}

type InMemoryGitHubOAuthStore struct {
	mu          sync.RWMutex
	oauthStates map[string]githubOAuthState
	sessions    map[string]*githubConnection
}

func NewInMemoryGitHubOAuthStore() *InMemoryGitHubOAuthStore {
	return &InMemoryGitHubOAuthStore{
		oauthStates: make(map[string]githubOAuthState),
		sessions:    make(map[string]*githubConnection),
	}
}

func (s *InMemoryGitHubOAuthStore) SaveOAuthState(_ context.Context, state string, entry githubOAuthState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.oauthStates[state] = entry
	return nil
}

func (s *InMemoryGitHubOAuthStore) ConsumeOAuthState(_ context.Context, state string) (githubOAuthState, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.oauthStates[state]
	if ok {
		delete(s.oauthStates, state)
	}
	return entry, ok, nil
}

func (s *InMemoryGitHubOAuthStore) CleanupExpiredOAuthStates(_ context.Context, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range s.oauthStates {
		if now.After(v.ExpiresAt) {
			delete(s.oauthStates, k)
		}
	}
	return nil
}

func (s *InMemoryGitHubOAuthStore) SaveGitHubConnection(_ context.Context, sessionID string, conn *githubConnection) error {
	if conn == nil {
		return fmt.Errorf("github connection is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	copyConn := *conn
	copyConn.Repos = append([]githubRepo(nil), conn.Repos...)
	s.sessions[sessionID] = &copyConn
	return nil
}

func (s *InMemoryGitHubOAuthStore) GetGitHubConnection(_ context.Context, sessionID string) (*githubConnection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn := s.sessions[sessionID]
	if conn == nil {
		return nil, nil
	}
	copyConn := *conn
	copyConn.Repos = append([]githubRepo(nil), conn.Repos...)
	return &copyConn, nil
}

func (s *InMemoryGitHubOAuthStore) DeleteGitHubConnection(_ context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
	return nil
}
