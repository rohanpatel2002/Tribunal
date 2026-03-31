package main

import (
	"context"
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
			ai_generated, critical, high, medium, low
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id`

	err = tx.QueryRow(ctx, insertPRQuery,
		response.Repository,
		response.PRNumber,
		response.Summary.Recommendation,
		response.Summary.TotalFiles,
		response.Summary.AIGenerated,
		response.Summary.Critical,
		response.Summary.High,
		response.Summary.Medium,
		response.Summary.Low,
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
	for _, result := range response.Results {
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
		&resp.Summary.Recommendation,
		&resp.Summary.TotalFiles,
		&resp.Summary.AIGenerated,
		&resp.Summary.Critical,
		&resp.Summary.High,
		&resp.Summary.Medium,
		&resp.Summary.Low,
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
		resp.Results = append(resp.Results, file)
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
