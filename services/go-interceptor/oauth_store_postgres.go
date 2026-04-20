package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresGitHubOAuthStore persists OAuth state and connected GitHub accounts.
type PostgresGitHubOAuthStore struct {
	pool *pgxpool.Pool
}

func NewPostgresGitHubOAuthStore(repo *PostgresRepository) *PostgresGitHubOAuthStore {
	if repo == nil || repo.pool == nil {
		return nil
	}
	return &PostgresGitHubOAuthStore{pool: repo.pool}
}

func (s *PostgresGitHubOAuthStore) SaveOAuthState(ctx context.Context, state string, entry githubOAuthState) error {
	query := `
		INSERT INTO github_oauth_states (state, session_id, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (state) DO UPDATE SET
			session_id = EXCLUDED.session_id,
			expires_at = EXCLUDED.expires_at
	`
	_, err := s.pool.Exec(ctx, query, state, entry.SessionID, entry.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to save oauth state: %w", err)
	}
	return nil
}

func (s *PostgresGitHubOAuthStore) ConsumeOAuthState(ctx context.Context, state string) (githubOAuthState, bool, error) {
	query := `
		DELETE FROM github_oauth_states
		WHERE state = $1
		RETURNING session_id, expires_at
	`
	var entry githubOAuthState
	err := s.pool.QueryRow(ctx, query, state).Scan(&entry.SessionID, &entry.ExpiresAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return githubOAuthState{}, false, nil
		}
		return githubOAuthState{}, false, fmt.Errorf("failed to consume oauth state: %w", err)
	}
	return entry, true, nil
}

func (s *PostgresGitHubOAuthStore) CleanupExpiredOAuthStates(ctx context.Context, now time.Time) error {
	query := `
		DELETE FROM github_oauth_states
		WHERE expires_at <= $1
	`
	_, err := s.pool.Exec(ctx, query, now)
	if err != nil {
		return fmt.Errorf("failed to cleanup oauth states: %w", err)
	}
	return nil
}

func (s *PostgresGitHubOAuthStore) SaveGitHubConnection(ctx context.Context, sessionID string, conn *githubConnection) error {
	if conn == nil {
		return fmt.Errorf("github connection is required")
	}

	reposJSON, err := json.Marshal(conn.Repos)
	if err != nil {
		return fmt.Errorf("failed to marshal github repos: %w", err)
	}

	query := `
		INSERT INTO github_connections (
			session_id, login, name, avatar_url, repos, connected_at, access_token, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (session_id) DO UPDATE SET
			login = EXCLUDED.login,
			name = EXCLUDED.name,
			avatar_url = EXCLUDED.avatar_url,
			repos = EXCLUDED.repos,
			connected_at = EXCLUDED.connected_at,
			access_token = EXCLUDED.access_token,
			updated_at = NOW()
	`

	_, err = s.pool.Exec(ctx, query,
		sessionID,
		conn.Login,
		conn.Name,
		conn.AvatarURL,
		reposJSON,
		conn.ConnectedAt,
		conn.AccessToken,
	)
	if err != nil {
		return fmt.Errorf("failed to save github connection: %w", err)
	}
	return nil
}

func (s *PostgresGitHubOAuthStore) GetGitHubConnection(ctx context.Context, sessionID string) (*githubConnection, error) {
	query := `
		SELECT login, name, avatar_url, repos, connected_at, access_token
		FROM github_connections
		WHERE session_id = $1
	`

	var (
		conn      githubConnection
		reposJSON []byte
	)

	err := s.pool.QueryRow(ctx, query, sessionID).Scan(
		&conn.Login,
		&conn.Name,
		&conn.AvatarURL,
		&reposJSON,
		&conn.ConnectedAt,
		&conn.AccessToken,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch github connection: %w", err)
	}

	if err := json.Unmarshal(reposJSON, &conn.Repos); err != nil {
		return nil, fmt.Errorf("failed to decode github repos: %w", err)
	}
	return &conn, nil
}

func (s *PostgresGitHubOAuthStore) DeleteGitHubConnection(ctx context.Context, sessionID string) error {
	query := `
		DELETE FROM github_connections
		WHERE session_id = $1
	`
	_, err := s.pool.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete github connection: %w", err)
	}
	return nil
}
