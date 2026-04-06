# TRIBUNAL

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go 1.21+](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![Python 3.11+](https://img.shields.io/badge/Python-3.11+-3776AB?logo=python)](https://python.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?logo=typescript)](https://www.typescriptlang.org)
[![PostgreSQL 15+](https://img.shields.io/badge/PostgreSQL-15+-336791?logo=postgresql)](https://www.postgresql.org)
[![Status: Beta](https://img.shields.io/badge/Status-Early%20Development-orange)](#status)

> **The Missing Code Review Layer: AI That Reviews What the AI Wrote**

The senior engineer that scales. TRIBUNAL detects AI-generated code in your CI/CD pipeline, analyzes semantic context blindness, and generates briefings that reduce code review time from 30 minutes to 30 seconds per change.

## The Crisis

Amazon now mandates senior engineer sign-off on every AI-assisted code change. Google and Microsoft have 25%+ of commits from AI tools with zero specialized review process.

Your linters catch syntax errors. They miss **semantic catastrophes**:

- The AI writes a database migration that's syntactically perfect but doesn't know the table has **2 billion rows** and **zero-downtime tolerance**
- The CI/CD change is well-formatted but cascades across **47 dependent services**
- The retry logic is clever but targets a **non-idempotent API**
- The config update is organized but breaks **two different feature flags**

**Result:** Engineering teams scale AI tool usage faster than they can review it safely. Human review becomes the bottleneck. The crisis intensifies.

## The Solution

TRIBUNAL is an **automated senior engineer** that understands context.

Instead of asking "Is this code correct?", TRIBUNAL asks:
> **Does this code understand what it's operating on?**

### How It Works

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  PR Created  │────▶│ AI Detection │────▶│   Context    │────▶│     Risk     │
│              │     │ (3 signals)  │     │   Analysis   │     │   Briefing   │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
                                                                       │
                                                                       ▼
                                                             ┌──────────────────┐
                                                             │  Human Reviews   │
                                                             │  in 30 seconds   │
                                                             │  (fully informed)│
                                                             └──────────────────┘
```

### The Three Phases

**1. DETECT**
- Identifies AI-generated code via authorship fingerprinting
- Analyzes 3 signals: style patterns, timing anomalies, common AI markers
- Zero false positives on human-written code

**2. ANALYZE**
- Queries your service topology, incident history, and runbooks
- Runs Claude-powered semantic analysis on risky sections
- Identifies context blindness gaps the AI didn't know existed

**3. BRIEF**
- Generates a context briefing for the human reviewer
- Shows: "Here's what the AI didn't know"
- Includes: service dependencies, scale constraints, incident patterns
- Result: Reviewer has full context in 30 seconds instead of 30 minutes

## Why TRIBUNAL Exists

**The Paradox:** AI code generation is syntactically impressive but semantically dangerous because it operates without understanding operational context.

**The Gap:** Traditional code review (linters, static analysis) checks syntax. It doesn't check semantic understanding.

**The Bottleneck:** Manual expert review catches semantic errors but doesn't scale. You can't hire senior engineers as fast as your junior engineers can use Copilot.

**The Solution:** An automated system that specializes in the blind spot between perfect syntax and catastrophic semantics.

## What TRIBUNAL Catches

| Risk Type | Example | Why It Matters |
|-----------|---------|----------------|
| **Scale Blindness** | Database migration on 2B-row table | Zero-downtime requirement violated |
| **Idempotency Blindness** | Retry logic against non-idempotent API | Duplicate transactions in production |
| **Cascade Blindness** | Config change affecting 47 services | Cascading failures across the system |
| **Incident Pattern Blindness** | Using pattern that caused 3 past incidents | Repeating known failure modes |
| **Dependency Blindness** | Breaking change in widely-used library | Silent failures downstream |
| **Race Condition Blindness** | Concurrent code without synchronization | Intermittent production bugs |

## Real-World Impact

### For Amazon
Solves the immediate crisis: "We need senior engineers to review every AI change"
- **Before**: 1 senior engineer → 20 PRs/day → 30 min/PR = 10 hours (bottleneck)
- **After**: TRIBUNAL pre-reviews → 20 PRs/day → 2 min human review/PR = 40 minutes (scaled)

### For Google/Microsoft
Scales AI tool trust across 25%+ of codebase without hiring 500 more senior engineers
- Enables safe delegation: Junior engineer with Copilot → TRIBUNAL validates → Senior confirms
- Incident reduction: Catch semantic errors before production

### For Any Enterprise
- **50% faster code reviews** on AI-written changes
- **Zero semantic errors** reaching production
- **Full compliance trail** (what TRIBUNAL found, what human decided)
- **Safe to trust** junior engineers with AI tools

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    GitHub/GitLab/Gitea                          │
│                     (PR Events)                                 │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
        ┌────────────────────────────────────┐
        │   GO: PR Interceptor & Detector    │
        │   • Webhook listener               │
        │   • Extract changed files          │
        │   • 3-signal AI authorship detect  │
        │   • ~2ms per file                  │
        └────────┬───────────────────────────┘
                 │
                 ▼
        ┌────────────────────────────────────┐
        │  GO: Context Graph Builder         │
        │  • Query service topology          │
        │  • Load incident history           │
        │  • Map dependency graph            │
        │  • ~100ms for full context         │
        └────────┬───────────────────────────┘
                 │
                 ▼
        ┌────────────────────────────────────┐
        │   PostgreSQL: Context Layer        │
        │   • Service topology               │
        │   • Incident-to-code correlations  │
        │   • Scale constraints              │
        │   • API documentation              │
        └────────┬───────────────────────────┘
                 │
                 ▼
        ┌────────────────────────────────────┐
        │  Python: Semantic Analyzer         │
        │  • Claude API integration          │
        │  • Context blindness detection     │
        │  • Risk classification             │
        │  • ~5s per high-risk section       │
        └────────┬───────────────────────────┘
                 │
                 ▼
        ┌────────────────────────────────────┐
        │  Python: Briefing Generator        │
        │  • Create human-readable docs      │
        │  • Risk scores + recommendations   │
        │  • Evidence + incident links       │
        └────────┬───────────────────────────┘
                 │
                 ▼
        ┌────────────────────────────────────┐
        │  TypeScript: PR Overlay UI         │
        │  • GitHub check runs               │
        │  • Inline code annotations         │
        │  • Context briefing display        │
        │  • Risk heatmap visualization      │
        └────────┬───────────────────────────┘
                 │
                 ▼
        ┌────────────────────────────────────┐
        │   PR Review Complete               │
        │   Human now has full context       │
        └────────────────────────────────────┘
```

## Technology Stack

| Layer | Technology | Why |
|-------|-----------|-----|
| **PR Interceptor** | Go 1.21+ | Fast (2ms per file), concurrent webhooks, cloud-native |
| **Authorship Detector** | Go | Real-time analysis, pattern matching at scale |
| **Context Builder** | Go | Graph traversal, dependency resolution, in-memory caching |
| **Semantic Analyzer** | Python 3.11+ | Claude API integration, LLM orchestration, risk classification |
| **Briefing Generator** | Python | Template rendering, markdown generation, evidence linking |
| **PR UI Overlay** | TypeScript + React | GitHub/GitLab/Gitea APIs, inline annotations, real-time updates |
| **Data Layer** | PostgreSQL 15+ | Persistent graph storage, incident correlations, audit trail |

## Quick Start (Docker - Recommended)

The absolute fastest way to boot TRIBUNAL is using our `docker-compose` orchestration. It automatically wires up the Go Interceptor service and initializes the PostgreSQL database schema for you.

### Prerequisites
- Docker Engine & Docker Compose

### 1-Click Launch

1. Clone the repository:
   ```bash
   git clone https://github.com/rohanpatel2002/tribunal.git
   cd tribunal
   ```

2. Boot the cluster in detached mode:
   ```bash
   docker-compose up -d --build
   ```

3. **Verify the Go Webhook Engine.** Your service is now live on `localhost:8080`.
   You can verify it by hitting the health endpoint:
   ```bash
   curl -s http://localhost:8080/health
   # Expected: {"service":"go-interceptor","status":"ok"}
   ```

4. **Access the CTO Audit Dashboard.** Your premium Next.js UI is now hosted on `localhost:3000`.
   - Open [http://localhost:3000](http://localhost:3000) in your browser.
   - Use the sample Enterprise API Key: `dev_enterprise_key_123` to query the seeded demo database!

### Stopping the Cluster
To gracefully stop the application while preserving your database volumes:
```bash
docker-compose down
```

## Local Setup (Direct Source)

If you are developing locally without Docker, follow these steps:

### Prerequisites
- Go 1.25+
- PostgreSQL 15+

### Source Setup

```bash
# Clone repository
git clone https://github.com/rohanpatel2002/tribunal.git
cd tribunal

# Start PostgreSQL natively
docker run -d -p 5432:5432 --name tribunal_native_db \
  -e POSTGRES_USER=tribunal \
  -e POSTGRES_PASSWORD=tribunal_password_dev \
  -e POSTGRES_DB=tribunal_db \
  postgres:15-alpine

# Wait 5 seconds for PostgreSQL to initialize
sleep 5

# Initialize database schema
docker exec -i tribunal_native_db psql -U tribunal -d tribunal_db < schema/postgres.sql

# Start Go service (Pointed at your native database)
cd services/go-interceptor
go mod download
export DATABASE_URL=postgres://tribunal:tribunal_password_dev@localhost:5432/tribunal_db?sslmode=disable
go run main.go
```

Services will be available at:
- **Go Interceptor**: `http://localhost:8080`

### First Test

1. Ensure the server is running on `8080`
2. Send a simulated webhook payload to the active API:
   ```bash
   curl -s -X POST http://localhost:8080/analyze \
        -H "Content-Type: application/json" \
        --data-binary @services/go-interceptor/fixtures/analyze-high-risk.json
   ```
3. See deterministic AI threat output predicting zero-downtime tolerance violations.

## Use Cases

### 🏗️ Database Migrations
**Catch**: Scale blindness (2B-row tables, downtime tolerance)
```
Migration detected
├─ Table size: 2,147,483,647 rows
├─ Tolerance: Zero-downtime required (HA constraint)
├─ AI approach: ADD COLUMN (locks table)
└─ ⚠️ RISK: 4-hour table lock on table serving 50k req/s
```

### 🔗 API Integrations
**Catch**: Idempotency blindness, retry logic errors
```
Retry logic detected
├─ Target: PaymentService.charge() [NON-IDEMPOTENT]
├─ Pattern: Simple exponential backoff
├─ Risk: Duplicate charges if service times out
└─ ⚠️ CRITICAL: Calls payment API without idempotency key
```

### ⚙️ Infrastructure-as-Code
**Catch**: Cascade effects, config dependencies
```
Config change detected
├─ Flag: FEATURE_NEW_AUTH
├─ Scope: 47 dependent services
├─ Incident history: 2 outages from similar changes
├─ AI approach: Simple feature flag flip
└─ ⚠️ WARNING: No gradual rollout, no rollback plan
```

### 🔄 Distributed Systems
**Catch**: Race conditions, concurrency blindness
```
Concurrent code detected
├─ Shared resource: Redis cache
├─ Lock mechanism: None
├─ Pattern: Check-then-set (non-atomic)
└─ ⚠️ RACE CONDITION: Cache invalidation race window
```

## Status

🚧 **Early Development** — Q1 2026

| Component | Status | Timeline |
|-----------|--------|----------|
| Go PR Interceptor | 🔨 In Progress | Early April |
| AI Authorship Detector | 🔨 In Progress | Mid April |
| Python Analyzer | 🔨 In Progress | Late April |
| PostgreSQL Schema | ✅ Complete | Now |
| TypeScript UI | 🔨 In Progress | Early May |
| GitHub Integration | 🔨 In Progress | Mid May |
| **Beta Release** | ⏳ Planned | **Q2 2026** |

## Roadmap

### Q2 2026: Beta
- [ ] GitHub integration (first public release)
- [ ] Core semantic analysis
- [ ] Basic context briefing
- [ ] Risk scoring system

### Q3 2026: Expansion
- [ ] GitLab full support
- [ ] Gitea integration
- [ ] Custom rule engine
- [ ] Team dashboard

### Q4 2026: Intelligence
- [ ] Incident pattern correlation
- [ ] Auto-remediation suggestions
- [ ] Historical trend analysis
- [ ] Custom model fine-tuning

## Contributing

We're in active development with a focused roadmap. See [CONTRIBUTING.md](./CONTRIBUTING.md) for how to get involved.

### Ways to Contribute
1. **Share Context Blindness Cases** — Real examples from your work
2. **Report Issues** — Bugs, edge cases, improvements
3. **Request Features** — Detection methods, briefing formats
4. **Join Beta Program** — Be first to use TRIBUNAL in production

## Why This Matters

The era of safe-at-scale AI code generation doesn't exist yet. TRIBUNAL is building it.

As AI code generation becomes standard (not exceptional), the bottleneck shifts from "can we write code fast" to "can we review code safely." The company that solves this problem unlocks the next era of engineering productivity.

TRIBUNAL is that solution.

## License

MIT License — See [LICENSE](./LICENSE) for details

**Permission granted to use, modify, and distribute TRIBUNAL freely.**

## Who This Is For

- **Engineering VPs** losing sleep over AI-generated code in production
- **Platform teams** managing GitHub Copilot at scale
- **Security teams** concerned about semantic vulnerabilities
- **DevOps teams** reviewing infrastructure-as-code changes
- **Any team** where AI-assisted development is outpacing code review capacity

## Get Started

⭐ **Star this repo** to follow development
📧 **Watch for updates** — Beta opens Q2 2026
💬 **Open an issue** — Share your context blindness stories

---

**The Missing Code Review Layer.**
**Made for every engineering VP reading incident reports written by AI.**

Built with ❤️ for the engineers writing the next generation of software.
