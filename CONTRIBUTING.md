# Contributing to TRIBUNAL

**Contributing to the Missing Code Review Layer**

Thank you for your interest in making AI-assisted code review safer at scale.

## Current Status

TRIBUNAL is in **active development** (Q1 2026). We're building a focused product with a clear roadmap.

## How to Contribute

### 1. Share Real-World Context Blindness Cases

**Most Valuable**: Examples from your experience where AI-generated code missed context and caused problems (or nearly did).

**Open an Issue**: Use the Context Blindness Case template

**Include**:
- What the AI wrote (code snippet or description)
- What context it missed
- Why it mattered
- What should have caught it

### 2. Report Issues & Bugs

Found something broken or weird?

**Open an Issue**: Use the Bug Report template

**Include**:
- Exact steps to reproduce
- Expected behavior
- Actual behavior
- System info (OS, Go version, Python version, etc.)

### 3. Request Features

Have ideas for detection methods, briefing formats, or integrations?

**Open an Issue**: Use the Feature Request template

**Examples of good feature requests**:
- "Detect AWS permission bypasses in IAM policy changes"
- "Add Slack integration for critical risk notifications"
- "Support custom organization-specific risk rules"

### 4. Join the Beta Program

Want to be first to use TRIBUNAL with your real GitHub repos?

**Expression of Interest**: Open an issue with your:
- Organization size
- Current Copilot usage
- Primary use cases
- Available timeline for feedback

### 5. Code Contributions (Limited)

We're not accepting code contributions yet, but we welcome feedback on architecture and design decisions.

---

## Development Setup

### Prerequisites

- Go 1.21+
- Python 3.11+
- Node.js 18+
- PostgreSQL 15+
- Docker (for local PostgreSQL)

### Quick Start

```bash
# Clone the repo
git clone https://github.com/YOUR_USERNAME/tribunal.git
cd tribunal

# Copy environment
cp .env.example .env

# Start PostgreSQL
docker run -d -p 5432:5432 \
  -e POSTGRES_PASSWORD=tribunal_dev \
  -e POSTGRES_DB=tribunal \
  postgres:15

# Wait for postgres
sleep 10

# Initialize database
psql -h localhost -U postgres -d tribunal -f schema/postgres.sql

# Start Go service (Terminal 1)
cd services/go-interceptor
go mod download
go run main.go

# Start Python service (Terminal 2)
cd services/python-analyzer
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python main.py

# Start TypeScript UI (Terminal 3)
cd services/ts-ui
npm install
npm run dev
```

---

## Community Guidelines

### Be Respectful
- Assume good intent
- Focus on ideas, not people
- Constructive criticism only

### Be Focused
- TRIBUNAL specializes in semantic code review
- Keep issues actionable
- Off-topic discussions moved separately

### Be Helpful
- If you report a bug, help us reproduce it
- If you suggest a feature, explain the use case
- Sharing context blindness cases is incredibly valuable

---

## Security Issues

Do NOT open public issues for security vulnerabilities.

Email: security@tribunal.dev (when available)

---

## Recognition

Contributors are recognized in:
- CHANGELOG.md
- GitHub Contributors section
- Release notes
- Eventually: CONTRIBUTORS.md

---

**Thank you for caring about safe-at-scale AI code review.**

—The TRIBUNAL Team
