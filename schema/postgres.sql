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
    context_briefing TEXT,
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

CREATE TABLE IF NOT EXISTS github_oauth_states (
    state TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_github_oauth_states_expires_at
    ON github_oauth_states(expires_at);

CREATE TABLE IF NOT EXISTS github_connections (
    session_id TEXT PRIMARY KEY,
    login TEXT NOT NULL,
    name TEXT,
    avatar_url TEXT,
    repos JSONB NOT NULL,
    connected_at TIMESTAMPTZ NOT NULL,
    access_token TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ENTERPRISE SAAS SCHEMA: Multi-Tenant Organizations and Subscriptions
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    api_key VARCHAR(128) UNIQUE DEFAULT gen_random_uuid()::varchar,
    subscription_tier VARCHAR(32) NOT NULL DEFAULT 'FREE', -- 'FREE', 'TEAM', 'ENTERPRISE'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS repositories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    full_name VARCHAR(255) UNIQUE NOT NULL, -- e.g., 'rohanpatel2002/tribunal'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_repositories_org_id ON repositories(organization_id);
