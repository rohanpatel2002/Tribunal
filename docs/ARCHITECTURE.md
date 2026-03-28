# TRIBUNAL Architecture

**Complete System Design & Technical Deep Dive**

## System Overview

TRIBUNAL is a distributed, microservice-based system that operates at the intersection of code analysis, semantic understanding, and context synthesis. It specializes in detecting AI-generated code and identifying semantic context blindness — where code is syntactically correct but operationally dangerous.

```
┌──────────────────────────────────────────────────────────────────────┐
│                        External Platforms                            │
│               (GitHub, GitLab, Gitea, Bitbucket)                    │
└────────────┬────────────────────────┬────────────────────────┬───────┘
             │                        │                        │
         Webhooks              PR Comments                 Notifications
             │                        │                        │
             ▼                        ▼                        ▼
    ┌─────────────────────────────────────────────────────────────┐
    │         TRIBUNAL: Distributed Analysis Platform            │
    │                                                              │
    │  ┌──────────────────────────────────────────────────────┐  │
    │  │ Layer 1: PR Ingestion & Detection (Go Service)      │  │
    │  │ • Webhook listener (scaled, concurrent)             │  │
    │  │ • File extraction & parsing                         │  │
    │  │ • 3-signal AI authorship detection                 │  │
    │  │ • Change classification (new/modified/deleted)      │  │
    │  │ Latency: ~2ms per file, ~50ms per PR               │  │
    │  └──────────────────────────────────────────────────────┘  │
    │                          │                                  │
    │                          ▼                                  │
    │  ┌──────────────────────────────────────────────────────┐  │
    │  │ Layer 2: Context Aggregation (Go Service)           │  │
    │  │ • Query service topology graph (cached)             │  │
    │  │ • Load incident history for changed services        │  │
    │  │ • Extract API documentation & constraints           │  │
    │  │ • Map cross-service dependencies                    │  │
    │  │ • Identify related runbooks & playbooks             │  │
    │  │ Latency: ~100ms (with caching)                      │  │
    │  └──────────────────────────────────────────────────────┘  │
    │                          │                                  │
    │                          ▼                                  │
    │  ┌──────────────────────────────────────────────────────┐  │
    │  │ Data Layer: PostgreSQL (Context Graph)              │  │
    │  │ • Service topology (nodes, edges, attributes)       │  │
    │  │ • Incident-to-code correlations                     │  │
    │  │ • Scale constraints & operational limits            │  │
    │  │ • Deployment history & rollouts                     │  │
    │  │ • AI authorship patterns & fingerprints             │  │
    │  │ • Briefing history & human decisions                │  │
    │  └──────────────────────────────────────────────────────┘  │
    │                          │                                  │
    │                          ▼                                  │
    │  ┌──────────────────────────────────────────────────────┐  │
    │  │ Layer 3: Semantic Analysis (Python Service)         │  │
    │  │ • Claude API integration for code understanding      │  │
    │  │ • Context blindness detection                       │  │
    │  │ • Risk classification & severity scoring             │  │
    │  │ • Generate remediation suggestions                  │  │
    │  │ • Correlate with historical patterns                │  │
    │  │ Latency: ~5s per high-risk section (parallel)       │  │
    │  └──────────────────────────────────────────────────────┘  │
    │                          │                                  │
    │                          ▼                                  │
    │  ┌──────────────────────────────────────────────────────┐  │
    │  │ Layer 4: Briefing Generation (Python Service)       │  │
    │  │ • Synthesize context documentation                  │  │
    │  │ • Generate markdown briefings                       │  │
    │  │ • Create evidence links to incidents                │  │
    │  │ • Build risk heatmap data                           │  │
    │  │ • Generate diffs highlighting risky sections        │  │
    │  │ Latency: ~1s per briefing (batched)                 │  │
    │  └──────────────────────────────────────────────────────┘  │
    │                          │                                  │
    │                          ▼                                  │
    │  ┌──────────────────────────────────────────────────────┐  │
    │  │ Layer 5: PR Overlay & Presentation (TypeScript)     │  │
    │  │ • Create check runs & annotations                   │  │
    │  │ • Inline code comments with briefings               │  │
    │  │ • Risk heatmap visualization                        │  │
    │  │ • Summary stats & recommendations                   │  │
    │  │ • Integration with GitHub/GitLab UX                 │  │
    │  │ Latency: <100ms (mostly formatting)                 │  │
    │  └──────────────────────────────────────────────────────┘  │
    │                          │                                  │
    └──────────────────────────┼──────────────────────────────────┘
                               │
                    ┌──────────▼──────────┐
                    │   PR Review Page    │
                    │   (Human Reviewer)  │
                    │  Full Context in    │
                    │   30 Seconds        │
                    └─────────────────────┘
```

## Component Architecture

### 1. GO Service: PR Interceptor & Detector

**Purpose**: Real-time webhook listening, file analysis, AI detection

**Responsibilities**:
- Listen for PR webhooks from GitHub, GitLab, Gitea
- Extract changed files and diffs
- Identify AI-generated code (3-signal analysis)
- Queue analysis jobs for downstream services
- Rate limiting & webhook validation

**3-Signal AI Detection**:

```
Signal 1: Style Fingerprinting
├─ Variable naming patterns (AI uses verbose names: "processUserAuthenticationRequest")
├─ Whitespace normalization (AI: consistent 2-space indents)
├─ Comment density (AI: often fewer comments relative to code)
├─ Bracket placement style (Go: specific conventions)
└─ Signal score: 0-1.0

Signal 2: Timing Analysis
├─ Commit timestamp + file change size correlation
├─ AI often commits in specific time windows
├─ Multiple large changes in rapid succession
├─ File entropy (randomness) analysis
└─ Signal score: 0-1.0

Signal 3: Pattern Matching
├─ Known Copilot/Claude marker patterns
├─ Specific error handling structures
├─ Boilerplate detection
├─ Import organization patterns
└─ Signal score: 0-1.0

Final Score = (Signal1 * 0.3) + (Signal2 * 0.4) + (Signal3 * 0.3)
Threshold: > 0.65 = AI-generated
```

**API Contract**:
```go
type PRAnalysisRequest struct {
    Repository   string
    PullNumber   int
    Owner        string
    ChangedFiles []struct {
        Path     string
        Status   string // added, modified, deleted
        Patch    string // diff content
    }
}

type AIDetectionResult struct {
    FileScans []struct {
        Path              string
        AIScore          float64 // 0-1.0
        SignalBreakdown  map[string]float64
        IsAIGenerated    bool
        Confidence       float64
    }
}
```

**Technologies**:
- Go 1.21 (concurrent, fast)
- Echo Web Framework (high-performance HTTP)
- In-memory caching (detection patterns)
- Worker pool for parallel file processing

---

### 2. GO Service: Context Graph Builder

**Purpose**: Aggregate all operational context about changed services

**Responsibilities**:
- Query service registry (Consul, Kubernetes, custom)
- Load incident history from Datadog/PagerDuty APIs
- Extract API contracts & idempotency information
- Map dependency graph
- Identify cascade risks
- Cache for performance

**Context Data Structure**:
```
Service Context = {
    ServiceMetadata: {
        Name, Owner, SLA, Tier
        CriticalityScore (1-5)
        LastIncident, MTTR
    },
    
    Operationalmetadata: {
        DatabaseTables: [
            {
                Name, Rows, IsSharded, DowntimeTolerance, LockingOperations
            }
        ],
        APIs: [
            {
                Endpoint, IsIdempotent, RateLimits, Timeout
            }
        ],
        ConfigFlags: [
            {
                Name, Scope (1-N services), RolloutStrategy
            }
        ]
    },
    
    RiskHistory: {
        IncidentPatterns: [
            {
                Pattern, Frequency, Services Affected, RootCause
            }
        ],
        FailedDeployments: [ ...],
        RaceConditions: [ ...],
    },
    
    Dependencies: {
        DirectDependencies: [Service],
        TransitiveDependencies: [Service],
        DataFlows: [
            {From, To, DataType, SLA}
        ]
    }
}
```

**Performance Notes**:
- Service topology: Cached (15-min TTL)
- Incident history: Cached (1-hour TTL)
- Full context assembly: ~100ms with warm cache

---

### 3. Python Service: Semantic Analyzer

**Purpose**: Identify context blindness using Claude API

**Responsibilities**:
- Call Claude API with code + context
- Identify semantic violations
- Generate risk scores
- Classify risk severity
- Link to incident patterns

**Analysis Prompts** (simplified):

**For Database Operations**:
```
Analyze this code change:
[CODE SNIPPET]

Against this context:
- Table: Users, Rows: 2.5B, ZeroDowntime: TRUE
- Current Schema Version: 5
- Rolling back from: Column X to Column Y

Question: Will this operation lock the table? 
If yes, for how long?
```

**For API Calls**:
```
Analyze this retry logic:
[CODE SNIPPET]

Against this context:
- Target API: PaymentService.charge()
- Idempotent: FALSE
- Timeout: 30s
- Typical latency: 2-5s

Question: Can this cause duplicate charges?
If yes, under what conditions?
```

**Risk Classification**:
```
CRITICAL: 
- Affects critical path (P0 service)
- High probability of incident
- Customer-facing impact immediate

HIGH:
- Affects important service (P1)
- Medium probability of incident
- Internal impact, may cascade

MEDIUM:
- Low-probability risk
- Limited scope or blast radius
- Mitigations exist

LOW:
- Edge case scenarios
- Theoretical risks
- Unlikely in practice
```

---

### 4. Python Service: Briefing Generator

**Purpose**: Create human-readable context documents

**Responsibilities**:
- Synthesize analysis results
- Create markdown briefings
- Link evidence (incidents, docs)
- Generate diffs with annotations
- Build risk heatmap data

**Briefing Template**:

```markdown
# Context Briefing: [PR Title]

## Risk Summary
- **Overall Risk**: HIGH
- **Critical Sections**: 2
- **High-Risk Changes**: 1
- **Recommended Action**: Request changes

## Critical Issues

### Issue 1: Scale Blindness in Database Migration
**Location**: src/migration/v45_add_active_status.sql

**What the AI Wrote**:
```sql
ALTER TABLE users ADD COLUMN active_status VARCHAR(50) DEFAULT 'pending';
```

**Context it Missed**:
- users table has 2,147,483,647 rows
- Zero-downtime requirement (HA SLA)
- Current locks: ~4 hours

**Why This Matters**:
ALTER COLUMN ... DEFAULT blocks the entire table
On production, this would:
1. Halt all login/checkout flows
2. Trigger all-hands incident
3. Revenue loss: ~$500k/hour

**Recommendation**:
Use online migration tool:
- Add column in background
- Run batch script with small chunks
- Deploy in stages (1% → 10% → 100%)

**See Also**:
- [Incident #2847](incidents/2847): Similar migration, 3-hour outage
- [Runbook: SafeMigrations](https://internal.docs/safe-migrations)

---

## Summary Statistics

| Metric | Value |
|--------|-------|
| Files Analyzed | 7 |
| AI-Generated | 4 |
| Critical Risks | 2 |
| High Risks | 3 |
| Estimated Review Time | 6 min (vs 30 min manual) |
| Required Signature | Engineering Manager |

---

## Recommendation

✋ **REQUEST CHANGES** — Critical issues require design review

---

Generated by TRIBUNAL
```

---

### 5. TypeScript Service: PR Overlay UI

**Purpose**: Present findings inline on GitHub/GitLab

**Responsibilities**:
- Create GitHub check runs
- Add inline code annotations
- Display briefing summaries
- Build risk heatmaps
- Handle platform-specific UX

**Check Run Output**:
```
┌─ TRIBUNAL Code Review ──────────────────────────────────┐
│                                                          │
│ Status: ⚠️  REQUESTED CHANGES                            │
│                                                          │
│ Critical Issues: 2  │ High-Risk: 3  │ Medium: 1         │
│                                                          │
│ ▓▓░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ │
│ 5 of 50 lines analyzed                                   │
│                                                          │
│ 🔴 CRITICAL: Scale blindness in v45 migration          │
│    src/migration/v45_add_active_status.sql:1            │
│                                                          │
│ 🟠 HIGH: Config cascade without rollout plan           │
│    src/config/feature_flags.yaml:47                     │
│                                                          │
│ Review Full Briefing >>>                                │
└──────────────────────────────────────────────────────────┘
```

**Inline Annotation Example**:
```
src/db/migration.go Line 23

  ALTER TABLE users 
  ADD COLUMN active_status VARCHAR(50) DEFAULT 'pending';

🚨 Context Alert:
  • users table: 2.5B rows
  • Zero-downtime required
  • ADD COLUMN with DEFAULT locks table ~4 hours

Similar to Incident #2847 (3-hour outage)
See: SafeMigrations Runbook | Full Briefing
```

---

## Data Flow

### Complete Request → Response Flow

```
1. PR Created (GitHub)
   └─> Webhook fired with PR details

2. Go Interceptor Service
   ├─> Extract files and diffs
   ├─> Run 3-signal AI detection
   ├─> Filter: Only process AI-flagged files
   └─> Queue analysis job with metadata

3. Go Context Graph Builder
   ├─> Query service topology cache
   ├─> Load incident history (last 1 year)
   ├─> Extract API contract details
   ├─> Map dependency graph
   ├─> Cache full context for analyzer
   └─> Pass to Python service

4. Python Semantic Analyzer
   ├─> For each high-risk file:
   │   ├─> Call Claude with code + context
   │   ├─> Parse response for semantic issues
   │   ├─> Classify risk severity
   │   └─> Store results (no token spillover)
   └─> Generate risk scores

5. Python Briefing Generator
   ├─> Synthesize all analysis
   ├─> Create markdown briefing
   ├─> Link evidence
   ├─> Generate heatmap data
   └─> Store in PostgreSQL

6. TypeScript PR Overlay
   ├─> Query briefing from database
   ├─> Create GitHub check run
   ├─> Add inline annotations
   ├─> Generate summary
   └─> Post to PR

7. Human Review
   ├─> Engineer sees PR with full context
   ├─> Reads 30-second summary
   ├─> Clicks through for details
   ├─> Makes informed decision
   └─> Approves or requests changes

Total Latency: ~15-30 seconds
```

---

## Database Schema

### Core Tables

```sql
-- Service Topology
CREATE TABLE services (
    id UUID PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    tier INT, -- 1=critical, 5=internal
    owner_team VARCHAR(255),
    sla_minutes INT,
    last_incident TIMESTAMP
);

CREATE TABLE service_dependencies (
    from_service_id UUID REFERENCES services(id),
    to_service_id UUID REFERENCES services(id),
    data_type VARCHAR(100),
    critical BOOLEAN
);

-- Operational Constraints
CREATE TABLE database_tables (
    id UUID PRIMARY KEY,
    service_id UUID REFERENCES services(id),
    name VARCHAR(255),
    row_count BIGINT,
    zero_downtime_tolerance BOOLEAN,
    last_migration TIMESTAMP
);

CREATE TABLE api_endpoints (
    id UUID PRIMARY KEY,
    service_id UUID REFERENCES services(id),
    path VARCHAR(255),
    is_idempotent BOOLEAN,
    timeout_ms INT,
    rate_limit_rps INT
);

-- Incident History
CREATE TABLE incidents (
    id UUID PRIMARY KEY,
    service_id UUID REFERENCES services(id),
    occurred_at TIMESTAMP,
    root_cause VARCHAR(500),
    pattern_category VARCHAR(100),
    duration_minutes INT
);

-- Analysis Results
CREATE TABLE pr_analyses (
    id UUID PRIMARY KEY,
    repo_name VARCHAR(255),
    pr_number INT,
    created_at TIMESTAMP,
    ai_detected_files INT,
    critical_risks INT,
    high_risks INT
);

CREATE TABLE briefings (
    id UUID PRIMARY KEY,
    analysis_id UUID REFERENCES pr_analyses(id),
    content TEXT, -- markdown
    generated_at TIMESTAMP,
    human_decision VARCHAR(50), -- approved, requested_changes
    decision_timestamp TIMESTAMP
);
```

---

## Performance & Scalability

### Latency Breakdown

| Component | Latency | Notes |
|-----------|---------|-------|
| Webhook ingestion | 2ms | Immediate |
| AI detection (per file) | 2ms | Parallel |
| Context aggregation | 100ms | Cached topology |
| Claude API call (per section) | 3-5s | Parallel, batched |
| Briefing generation | 1s | Parallel |
| UI rendering | 100ms | Async |
| **Total** | **~15-30s** | Highly parallelized |

### Scaling Strategies

**Horizontal Scaling**:
- Go services: Deploy as StatelessK8s pods
- Python services: Use worker pool with task queue (RabbitMQ/Kafka)
- TypeScript: Stateless, can scale independently

**Caching**:
- Service topology: 15-min Redis TTL
- Incident history: 1-hour Redis TTL
- Detection patterns: In-memory LRU cache (Go)

**Optimization**:
- Only analyze files flagged as AI-generated (80% reduction)
- Batch Claude API calls (reduce rate limit issues)
- Parallel analysis for multiple files
- Skip analysis if < 5 LOC changed

---

## Security Considerations

### Data Sensitivity

**PII/Secrets**:
- Code diffs may contain secrets
- Briefings must never expose secrets
- Claude API: Code only, not full context with credentials
- All data at-rest encrypted

**Access Control**:
- Only repo contributors see their PRs
- Platform admins see full activity
- Incident history: Team-based access

**Compliance**:
- GDPR: Briefings purged after 90 days
- SOC2: Audit trail of all decisions
- GitHub: Works with GitHub Enterprise (self-hosted option)

---

## Integration Points

### External Services

**GitHub/GitLab/Gitea APIs**:
- Webhook subscriptions
- PR detail queries
- Check runs creation
- Comment posting
- Status updates

**Incident Management** (Datadog, PagerDuty):
- Historical incident data
- Service correlations
- MTTR tracking
- Escalation policies

**Service Registry** (Consul, Kubernetes API):
- Service discovery
- Health checks
- Dependency mapping
- Configuration

**Documentation** (Confluence, Notion):
- Runbook links
- API documentation
- Deployment guides
- Previous decisions

---

## Deployment Architecture

### Multi-Region Setup

```
┌─ US Region (Primary) ──────────────────┐
│ • Go Interceptor x3 (LB)               │
│ • Python Analyzer x5 (Task queue)      │
│ • TypeScript Service x2 (LB)           │
│ • PostgreSQL Primary (Multi-AZ)        │
│ • Redis Cache x2 (HA)                  │
└────────────────────────────────────────┘

┌─ EU Region (Secondary) ────────────────┐
│ • Go Interceptor x2                    │
│ • Python Analyzer x3                   │
│ • TypeScript Service x1                │
│ • PostgreSQL Replica                   │
│ • Redis Replica                        │
└────────────────────────────────────────┘

Data Replication: PostgreSQL Streaming, Redis Sentinel
Failover: Automatic (RTO < 1 min, RPO < 1 sec)
```

### Development vs Production

**Dev**:
- Single Go service (port 8080)
- Single Python service (port 8081)
- TypeScript on localhost:3000
- SQLite for database (portability)

**Production**:
- Kubernetes StatefulSets/Deployments
- PostgreSQL managed service (AWS RDS, Heroku, etc.)
- Redis managed service (AWS ElastiCache)
- Load balancers (Kubernetes ingress)
- CDN for static assets (CloudFront)

---

This architecture is production-ready, scalable, and designed for rapid iteration on detection methods without redesign.
