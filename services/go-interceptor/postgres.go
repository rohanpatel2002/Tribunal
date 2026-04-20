package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements the Repository interface for PostgreSQL
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository establishes a connection pool and returns the repository
func NewPostgresRepository(ctx context.Context, connString string) (*PostgresRepository, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database url: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Ping database to verify connection is alive
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &PostgresRepository{pool: pool}, nil
}

// Close gracefully closes all connections in the pool
func (r *PostgresRepository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

// SaveAnalysis inserts the PR analysis and its file results atomically within a transaction.
func (r *PostgresRepository) SaveAnalysis(ctx context.Context, response *AnalyzeResponse) error {
	// Begin transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert into pr_analyses and capture generated UUID
	var prAnalysisID string
	insertPRQuery := `
		INSERT INTO pr_analyses (
			repository, pr_number, recommendation, total_files, 
			ai_generated, critical, high, medium, low, context_briefing
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id`

	err = tx.QueryRow(ctx, insertPRQuery,
		response.Repository,
		response.PRNumber,
		response.Recommendation,
		response.TotalFiles,
		response.AIGenerated,
		response.Critical,
		response.High,
		response.Medium,
		response.Low,
		response.ContextBriefing,
	).Scan(&prAnalysisID)

	if err != nil {
		return fmt.Errorf("failed to insert pr_analysis: %w", err)
	}

	// Prepare batched insert for files for peak performance
	insertFileQuery := `
		INSERT INTO file_analyses (
			pr_analysis_id, path, ai_score, is_ai_generated, confidence,
			style_signal, pattern_signal, risk_signal, risk_level, summary
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)`

	batch := &pgx.Batch{}
	for _, result := range response.Files {
		batch.Queue(insertFileQuery,
			prAnalysisID,
			result.Path,
			result.AIScore,
			result.IsAIGenerated,
			result.Confidence,
			result.Signals.Style,
			result.Signals.Pattern,
			result.Signals.Risk,
			result.RiskLevel,
			result.Summary,
		)
	}

	// Send batch to Postgres
	br := tx.SendBatch(ctx, batch)
	_, err = br.Exec() // Run the batch but we don't strictly need the return tag yet

	// Must close batch before committing transaction
	if closeErr := br.Close(); closeErr != nil {
		return fmt.Errorf("failed to close batch operation: %w", closeErr)
	}

	if err != nil {
		return fmt.Errorf("failed to insert file_analyses: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetAnalysisByPR retrieves a PR analysis report by matching the repo and PR number
func (r *PostgresRepository) GetAnalysisByPR(ctx context.Context, repository string, prNumber int) (*AnalyzeResponse, error) {
	queryPR := `
		SELECT id, recommendation, total_files, ai_generated, critical, high, medium, low
		FROM pr_analyses
		WHERE repository = $1 AND pr_number = $2
		ORDER BY created_at DESC
		LIMIT 1`

	resp := &AnalyzeResponse{
		Repository: repository,
		PRNumber:   prNumber,
	}

	var prAnalysisID string
	err := r.pool.QueryRow(ctx, queryPR, repository, prNumber).Scan(
		&prAnalysisID,
		&resp.Recommendation,
		&resp.TotalFiles,
		&resp.AIGenerated,
		&resp.Critical,
		&resp.High,
		&resp.Medium,
		&resp.Low,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrAnalysisNotFound
		}
		return nil, fmt.Errorf("failed to get PR analysis summary: %w", err)
	}

	// Fetch granular file-level analyses
	queryFiles := `
		SELECT path, ai_score, is_ai_generated, confidence,
			   style_signal, pattern_signal, risk_signal, risk_level, summary
		FROM file_analyses
		WHERE pr_analysis_id = $1`

	rows, err := r.pool.Query(ctx, queryFiles, prAnalysisID)
	if err != nil {
		return nil, fmt.Errorf("failed to query file analyses: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var file FileAnalysis
		err := rows.Scan(
			&file.Path,
			&file.AIScore,
			&file.IsAIGenerated,
			&file.Confidence,
			&file.Signals.Style,
			&file.Signals.Pattern,
			&file.Signals.Risk,
			&file.RiskLevel,
			&file.Summary,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file analysis row: %w", err)
		}
		resp.Files = append(resp.Files, file)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating file analyses: %w", err)
	}

	return resp, nil
}

// MarkWebhookProcessed records the delivery ID to ensure idempotent processing.
func (r *PostgresRepository) MarkWebhookProcessed(ctx context.Context, deliveryID string, repoFullName string) (bool, error) {
	query := `
		INSERT INTO processed_webhooks (delivery_id, repository)
		VALUES ($1, $2)
		ON CONFLICT (delivery_id) DO NOTHING
	`
	commandTag, err := r.pool.Exec(ctx, query, deliveryID, repoFullName)
	if err != nil {
		return false, fmt.Errorf("failed to mark webhook processed: %w", err)
	}

	// If RowsAffected is 0, the delivery_id was already present (due to ON CONFLICT)
	return commandTag.RowsAffected() > 0, nil
}

// GetRepositoryAuditSummary computes aggregate analytics for a given repository payload.
func (r *PostgresRepository) GetRepositoryAuditSummary(ctx context.Context, repository string) (*AuditSummary, error) {
	summary := &AuditSummary{Repository: repository}
	query := `
		SELECT
				COALESCE(SUM(total_files), 0),
				COALESCE(SUM(ai_generated), 0),
				COALESCE(SUM(critical), 0),
				COALESCE(SUM(high), 0),
				COUNT(id),
				COALESCE(AVG(CASE WHEN total_files > 0 THEN ai_generated::float / total_files ELSE 0 END), 0)
		FROM pr_analyses
		WHERE LOWER(repository) = LOWER($1)
		`

	err := r.pool.QueryRow(ctx, query, repository).Scan(
		&summary.TotalFiles,
		&summary.AIGeneratedPRs,
		&summary.CriticalRisks,
		&summary.HighRisks,
		&summary.TotalPRs,
		&summary.AverageAIScore,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to aggregate audit metrics: %w", err)
	}

	return summary, nil
}

// GetSubscriptionTier maps the full repo name to an organization and queries its SaaS tier. Returns "FREE" by default if entirely unmapped.
func (r *PostgresRepository) GetSubscriptionTier(ctx context.Context, repoFullName string) (string, error) {
	query := `
        SELECT o.subscription_tier
        FROM organizations o
        JOIN repositories r ON r.organization_id = o.id
        WHERE r.full_name = $1
        LIMIT 1
        `

	var tier string
	err := r.pool.QueryRow(ctx, query, repoFullName).Scan(&tier)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "FREE", nil // Default logic: All untracked repos get the Free heuristic tier
		}
		return "FREE", fmt.Errorf("could not lookup subscription tier: %v", err)
	}

	return tier, nil
}

// GetRecentAnalyses grabs the most recent PR results to populate the enterprise audit log table.
func (r *PostgresRepository) GetRecentAnalyses(ctx context.Context, limit int, repository string) ([]PRAnalysisRecord, error) {
	query := `
		SELECT id::text, repository, pr_number, recommendation, total_files, ai_generated, critical, high, medium, low, context_briefing, created_at
		FROM pr_analyses
		WHERE LOWER(repository) = LOWER($1)
		ORDER BY created_at DESC
		LIMIT $2
		`
	rows, err := r.pool.Query(ctx, query, repository, limit)
	if err != nil {
		return nil, fmt.Errorf("failed querying recent PR analyses: %w", err)
	}
	defer rows.Close()

	var results []PRAnalysisRecord
	for rows.Next() {
		var rec PRAnalysisRecord
		var contextBriefing sql.NullString
		if err := rows.Scan(
			&rec.ID, &rec.Repository, &rec.PRNumber, &rec.Recommendation,
			&rec.TotalFiles, &rec.AIGenerated, &rec.Critical, &rec.High,
			&rec.Medium, &rec.Low, &contextBriefing, &rec.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning PR logs row: %w", err)
		}
		if contextBriefing.Valid {
			rec.ContextBriefing = contextBriefing.String
		}
		results = append(results, rec)
	}
	return results, rows.Err()
}

// SaveSecurityPolicy persists a security policy to the database.
func (r *PostgresRepository) SaveSecurityPolicy(ctx context.Context, policy *SecurityPolicy) error {
	query := `
		INSERT INTO security_policies (repository, policy_name, policy_type, description, rules, is_active, severity_threshold, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (repository, policy_name) DO UPDATE SET
			policy_type = EXCLUDED.policy_type,
			description = EXCLUDED.description,
			rules = EXCLUDED.rules,
			is_active = EXCLUDED.is_active,
			severity_threshold = EXCLUDED.severity_threshold,
			updated_at = NOW()
	`

	rulesJSON, err := json.Marshal(policy.Rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	_, err = r.pool.Exec(ctx, query,
		policy.Repository,
		policy.PolicyName,
		policy.PolicyType,
		policy.Description,
		rulesJSON,
		policy.IsActive,
		policy.SeverityThreshold,
		policy.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to save security policy: %w", err)
	}

	// Log the event
	eventQuery := `
		INSERT INTO security_events (event_type, repository, actor, action_details, severity, created_at)
		VALUES ('POLICY_CREATED', $1, $2, $3, 'INFO', NOW())
	`
	actionDetails := map[string]string{"policy_name": policy.PolicyName}
	actionJSON, _ := json.Marshal(actionDetails)

	_, _ = r.pool.Exec(ctx, eventQuery, policy.Repository, policy.CreatedBy, actionJSON)

	return nil
}

// GetSecurityPolicies retrieves all active policies for a repository.
func (r *PostgresRepository) GetSecurityPolicies(ctx context.Context, repository string) ([]SecurityPolicy, error) {
	query := `
		SELECT id::text, repository, policy_name, policy_type, description, rules, is_active, severity_threshold, created_at
		FROM security_policies
		WHERE repository = $1 AND is_active = TRUE
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, repository)
	if err != nil {
		return nil, fmt.Errorf("failed to query security policies: %w", err)
	}
	defer rows.Close()

	var policies []SecurityPolicy
	for rows.Next() {
		var policy SecurityPolicy
		var rulesJSON []byte

		err := rows.Scan(
			&policy.ID,
			&policy.Repository,
			&policy.PolicyName,
			&policy.PolicyType,
			&policy.Description,
			&rulesJSON,
			&policy.IsActive,
			&policy.SeverityThreshold,
			&policy.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		err = json.Unmarshal(rulesJSON, &policy.Rules)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
		}

		policies = append(policies, policy)
	}

	return policies, rows.Err()
}

// DeleteSecurityPolicy deactivates a security policy.
func (r *PostgresRepository) DeleteSecurityPolicy(ctx context.Context, repository string, policyName string, actor string) error {
	query := `
		UPDATE security_policies
		SET is_active = FALSE, updated_at = NOW()
		WHERE repository = $1 AND policy_name = $2
	`

	_, err := r.pool.Exec(ctx, query, repository, policyName)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	// Log event
	eventQuery := `
		INSERT INTO security_events (event_type, repository, actor, action_details, severity, created_at)
		VALUES ('POLICY_DELETED', $1, $2, $3, 'WARNING', NOW())
	`
	actionDetails := map[string]string{"policy_name": policyName}
	actionJSON, _ := json.Marshal(actionDetails)

	_, _ = r.pool.Exec(ctx, eventQuery, repository, actor, actionJSON)

	return nil
}
