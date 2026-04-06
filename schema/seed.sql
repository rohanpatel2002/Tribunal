-- Seed Data for TRIBUNAL CTO Dashboard Demo
-- Automatically injects demo organizations, repositories, and sample analysis data

INSERT INTO organizations (id, name, api_key, subscription_tier)
VALUES 
    ('00000000-0000-0000-0000-000000000001', 'Acme Corp', 'dev_enterprise_key_123', 'ENTERPRISE')
ON CONFLICT DO NOTHING;

INSERT INTO repositories (id, organization_id, full_name)
VALUES 
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'rohanpatel2002/tribunal')
ON CONFLICT DO NOTHING;

INSERT INTO pr_analyses (repository, pr_number, recommendation, total_files, ai_generated, critical, high, medium, low)
VALUES 
    ('rohanpatel2002/tribunal', 101, 'BLOCK', 12, 12, 1, 2, 0, 9),
    ('rohanpatel2002/tribunal', 102, 'APPROVE', 3, 0, 0, 0, 0, 3),
    ('rohanpatel2002/tribunal', 103, 'REVIEW_REQUIRED', 8, 5, 0, 1, 4, 3),
    ('rohanpatel2002/tribunal', 104, 'BLOCK', 45, 38, 3, 8, 12, 22),
    ('rohanpatel2002/tribunal', 105, 'APPROVE', 2, 0, 0, 0, 0, 2);
