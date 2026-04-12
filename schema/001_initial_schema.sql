-- Tribunal MVP - Initial Schema Migration (001)
-- Date: April 12, 2026
-- Purpose: Set up core audit, policy, and security event tables

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS uuid-ossp;

-- ============================================================================
-- CORE AUDIT TABLES
-- ============================================================================

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository VARCHAR(255) NOT NULL,
    pr_number INTEGER NOT NULL,
    recommendation VARCHAR(32) NOT NULL CHECK (recommendation IN ('APPROVE', 'BLOCK', 'REVIEW_REQUIRED')),
    total_files INTEGER NOT NULL DEFAULT 0,
    ai_generated INTEGER NOT NULL DEFAULT 0,
    critical INTEGER NOT NULL DEFAULT 0,
    high INTEGER NOT NULL DEFAULT 0,
    medium INTEGER NOT NULL DEFAULT 0,
    low INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_repository ON audit_logs(repository);
CREATE INDEX IF NOT EXISTS idx_audit_logs_pr_number ON audit_logs(repository, pr_number);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- ============================================================================
-- FILE ANALYSIS DETAILS
-- ============================================================================

CREATE TABLE IF NOT EXISTS file_analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    audit_log_id UUID NOT NULL REFERENCES audit_logs(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    ai_score NUMERIC(5,4) NOT NULL CHECK (ai_score BETWEEN 0 AND 1),
    is_ai_generated BOOLEAN NOT NULL,
    confidence NUMERIC(5,4) NOT NULL CHECK (confidence BETWEEN 0 AND 1),
    risk_level VARCHAR(16) NOT NULL CHECK (risk_level IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    summary TEXT,
    suggested_fix TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_file_analyses_audit_log_id ON file_analyses(audit_log_id);
CREATE INDEX IF NOT EXISTS idx_file_analyses_file_path ON file_analyses(file_path);
CREATE INDEX IF NOT EXISTS idx_file_analyses_risk_level ON file_analyses(risk_level);

-- ============================================================================
-- SECURITY POLICIES
-- ============================================================================

CREATE TABLE IF NOT EXISTS security_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository VARCHAR(255) NOT NULL,
    policy_name VARCHAR(255) NOT NULL,
    policy_type VARCHAR(64) NOT NULL CHECK (policy_type IN ('AI_DETECTION', 'VULNERABILITY_SCAN', 'CODE_STYLE', 'COMPLIANCE')),
    description TEXT,
    rules JSONB NOT NULL DEFAULT '{}'::JSONB,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    severity_threshold VARCHAR(16) NOT NULL DEFAULT 'MEDIUM' CHECK (severity_threshold IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    UNIQUE(repository, policy_name)
);

CREATE INDEX IF NOT EXISTS idx_security_policies_repository ON security_policies(repository);
CREATE INDEX IF NOT EXISTS idx_security_policies_is_active ON security_policies(is_active);
CREATE INDEX IF NOT EXISTS idx_security_policies_type ON security_policies(policy_type);

-- ============================================================================
-- SECURITY EVENTS (Audit Trail)
-- ============================================================================

CREATE TABLE IF NOT EXISTS security_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(64) NOT NULL CHECK (event_type IN ('POLICY_CREATED', 'POLICY_UPDATED', 'POLICY_DELETED', 'ANALYSIS_COMPLETED', 'RISK_DETECTED', 'API_KEY_ROTATED', 'AUTH_FAILED')),
    repository VARCHAR(255),
    actor VARCHAR(255) NOT NULL,
    action_details JSONB,
    severity VARCHAR(16) NOT NULL DEFAULT 'INFO' CHECK (severity IN ('INFO', 'WARNING', 'ERROR', 'CRITICAL')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_security_events_event_type ON security_events(event_type);
CREATE INDEX IF NOT EXISTS idx_security_events_repository ON security_events(repository);
CREATE INDEX IF NOT EXISTS idx_security_events_actor ON security_events(actor);
CREATE INDEX IF NOT EXISTS idx_security_events_created_at ON security_events(created_at DESC);

-- ============================================================================
-- API KEY MANAGEMENT
-- ============================================================================

CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_name VARCHAR(255) NOT NULL,
    repository VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by VARCHAR(255),
    CONSTRAINT valid_expiry CHECK (expires_at IS NULL OR expires_at > created_at)
);

CREATE INDEX IF NOT EXISTS idx_api_keys_repository ON api_keys(repository);
CREATE INDEX IF NOT EXISTS idx_api_keys_is_active ON api_keys(is_active);
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys(expires_at);

-- ============================================================================
-- WEBHOOK IDEMPOTENCY TRACKING
-- ============================================================================

CREATE TABLE IF NOT EXISTS processed_webhooks (
    delivery_id VARCHAR(255) PRIMARY KEY,
    repository VARCHAR(255) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_processed_webhooks_repository ON processed_webhooks(repository);
CREATE INDEX IF NOT EXISTS idx_processed_webhooks_processed_at ON processed_webhooks(processed_at);

-- ============================================================================
-- SUMMARY VIEW FOR DASHBOARD
-- ============================================================================

CREATE VIEW IF NOT EXISTS audit_summary_view AS
SELECT
    repository,
    COUNT(DISTINCT pr_number) as total_prs,
    SUM(total_files) as total_files,
    COUNT(CASE WHEN ai_generated > 0 THEN 1 END) as ai_generated_prs,
    SUM(critical) as critical_risks,
    SUM(high) as high_risks,
    ROUND(AVG(COALESCE(ai_generated::NUMERIC / NULLIF(total_files, 0), 0))::NUMERIC, 4) as average_ai_score,
    MAX(created_at) as last_analyzed
FROM audit_logs
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY repository;

-- ============================================================================
-- TRIGGERS FOR AUDIT TRAIL
-- ============================================================================

CREATE OR REPLACE FUNCTION log_policy_change()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO security_events (event_type, repository, actor, action_details, severity)
    VALUES (
        CASE 
            WHEN TG_OP = 'INSERT' THEN 'POLICY_CREATED'
            WHEN TG_OP = 'UPDATE' THEN 'POLICY_UPDATED'
            WHEN TG_OP = 'DELETE' THEN 'POLICY_DELETED'
        END,
        NEW.repository,
        COALESCE(NEW.created_by, 'system'),
        jsonb_build_object(
            'policy_id', COALESCE(NEW.id, OLD.id)::text,
            'policy_name', COALESCE(NEW.policy_name, OLD.policy_name),
            'changes', jsonb_build_object(
                'is_active', NEW.is_active,
                'severity_threshold', NEW.severity_threshold
            )
        ),
        'INFO'
    );
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER security_policies_audit_trigger
AFTER INSERT OR UPDATE OR DELETE ON security_policies
FOR EACH ROW
EXECUTE FUNCTION log_policy_change();

-- ============================================================================
-- PERMISSIONS & CONSTRAINTS
-- ============================================================================

-- Ensure data consistency
ALTER TABLE audit_logs ADD CONSTRAINT check_audit_files
    CHECK (total_files >= ai_generated AND ai_generated >= 0);
ALTER TABLE audit_logs ADD CONSTRAINT check_audit_risks
    CHECK ((critical + high + medium + low) <= total_files);
