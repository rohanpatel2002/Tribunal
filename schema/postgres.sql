-- TRIBUNAL MVP schema

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS pr_analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository VARCHAR(255) NOT NULL,
    pr_number INT NOT NULL,
    recommendation VARCHAR(32) NOT NULL,
    total_files INT NOT NULL,
    ai_generated INT NOT NULL,
    critical INT NOT NULL,
    high INT NOT NULL,
    medium INT NOT NULL,
    low INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pr_analyses_repo_pr
    ON pr_analyses(repository, pr_number);

CREATE TABLE IF NOT EXISTS file_analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pr_analysis_id UUID NOT NULL REFERENCES pr_analyses(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    ai_score NUMERIC(4,2) NOT NULL,
    is_ai_generated BOOLEAN NOT NULL,
    confidence NUMERIC(4,2) NOT NULL,
    style_signal NUMERIC(4,2) NOT NULL,
    pattern_signal NUMERIC(4,2) NOT NULL,
    risk_signal NUMERIC(4,2) NOT NULL,
    risk_level VARCHAR(16) NOT NULL,
    summary TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_file_analyses_pr_analysis_id
    ON file_analyses(pr_analysis_id);

CREATE TABLE IF NOT EXISTS processed_webhooks (
    delivery_id VARCHAR(255) PRIMARY KEY,
    repository VARCHAR(255) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
