# INTEGRATION GUIDE

**Set Up TRIBUNAL with GitHub, GitLab, or Gitea**

## GitHub Integration

### Prerequisites

- GitHub organization or personal repo with admin access
- TRIBUNAL instance running (self-hosted or SaaS)
- GitHub Personal Access Token (PAT) with `repo` and `workflow` scopes

### Step 1: Create GitHub App

**Recommended Approach**: GitHub Apps (vs. webhooks) provide better security and rate limiting.

1. Go to `https://github.com/settings/apps`
2. Click "New GitHub App"
3. Fill in:
   - **App name**: `Tribunal`
   - **Homepage URL**: `https://tribunal.your-domain.com`
   - **Webhook URL**: `https://tribunal.your-domain.com/webhooks/github`
   - **Webhook secret**: Generate a secure random string, save it

4. **Permissions**:
   - Read access:
     - Commit commits
     - Pull requests
     - Repository metadata
   - Write access:
     - Checks
     - Pull request reviews
     - Commit statuses

5. **Subscribe to events**:
   - Pull request
   - Push

6. **Installation**: Click "Create GitHub App"
7. **Private key**: Generate and download, save securely

### Step 2: Configure TRIBUNAL

Update your `.env`:

```bash
# GitHub Configuration
GITHUB_APP_ID=<your-app-id>
GITHUB_APP_PRIVATE_KEY=<path-to-pem-file>
GITHUB_WEBHOOK_SECRET=<your-webhook-secret>

# Webhook configuration
WEBHOOK_PORT=8080
WEBHOOK_PATH=/webhooks/github
```

### Step 3: Install on Repositories

1. Go to app settings → "Install App"
2. Select organization
3. Choose "All repositories" or specific repos
4. Authorize

### Step 4: Test

Create a test PR with AI-generated code. You should see:
- ✅ Webhook fires
- ✅ TRIBUNAL analyzes
- ✅ Check run appears on PR
- ✅ Annotations visible

---

## GitLab Integration

### Prerequisites

- GitLab instance (self-hosted or gitlab.com)
- Admin access
- TRIBUNAL instance accessible via HTTPS

### Step 1: Create Group/Project Webhook

For **Group-level reviews** (recommended):

1. Go to Group → Settings → Webhooks
2. **URL**: `https://tribunal.your-domain.com/webhooks/gitlab`
3. **Secret token**: Generate a secure random string
4. **Trigger events**:
   - Merge request events
   - Push events
5. SSL verification: ✅ (recommended)
6. Click "Add webhook"

For **Project-level reviews**:

1. Go to Project → Settings → Webhooks
2. Same configuration as above
3. Click "Add webhook"

### Step 2: Create GitLab User Token

For TRIBUNAL to post comments/reviews:

1. Go to Admin Panel → Users
2. Create service account: `tribunal-bot`
3. Go to User Settings → Access Tokens
4. Create token with scopes:
   - `api` (full API access)
   - `read_api` (read repository)
   - `write_repository` (post comments)
5. Save token securely

### Step 3: Configure TRIBUNAL

Update `.env`:

```bash
# GitLab Configuration
GITLAB_INSTANCE_URL=https://gitlab.com  # or your self-hosted URL
GITLAB_WEBHOOK_SECRET=<your-webhook-secret>
GITLAB_BOT_TOKEN=<your-bot-token>

# Webhook configuration
WEBHOOK_PORT=8080
WEBHOOK_PATH=/webhooks/gitlab
```

### Step 4: Test

Create a test merge request with AI-generated code. You should see:
- ✅ Webhook fires
- ✅ TRIBUNAL analyzes
- ✅ Bot posts comments on MR
- ✅ Annotations visible in diff

---

## Gitea Integration

### Prerequisites

- Gitea instance running (v1.20+)
- Repository admin access
- TRIBUNAL accessible via HTTPS

### Step 1: Create Webhook

1. Go to Repository → Settings → Webhooks
2. **Payload URL**: `https://tribunal.your-domain.com/webhooks/gitea`
3. **Content type**: `application/json`
4. **Secret**: Generate secure random string
5. **Events**:
   - Pull request
   - Push
6. **Active**: ✅
7. Click "Add webhook"

### Step 2: Create API Token

For TRIBUNAL to post reviews:

1. Go to Settings → Applications → OAuth2 Applications
2. Create application: `Tribunal`
3. Redirect URI: `https://tribunal.your-domain.com/auth/gitea/callback`
4. Save Client ID and Secret

### Step 3: Configure TRIBUNAL

Update `.env`:

```bash
# Gitea Configuration
GITEA_INSTANCE_URL=https://gitea.your-domain.com
GITEA_WEBHOOK_SECRET=<your-webhook-secret>
GITEA_CLIENT_ID=<client-id>
GITEA_CLIENT_SECRET=<client-secret>

# Webhook configuration
WEBHOOK_PORT=8080
WEBHOOK_PATH=/webhooks/gitea
```

### Step 4: Test

Create a test PR with AI-generated code. TRIBUNAL should analyze and post comments.

---

## Self-Hosted Deployment

### Docker Compose Setup

```yaml
version: '3.8'

services:
  tribunal-go:
    image: tribunal:latest-go
    ports:
      - "8080:8080"
    environment:
      - GITHUB_APP_ID=${GITHUB_APP_ID}
      - GITHUB_APP_PRIVATE_KEY=/secrets/github.pem
      - WEBHOOK_SECRET=${WEBHOOK_SECRET}
      - PYTHON_SERVICE_URL=http://tribunal-python:8081
    volumes:
      - /path/to/github.pem:/secrets/github.pem:ro
    depends_on:
      - postgres
      - redis

  tribunal-python:
    image: tribunal:latest-python
    ports:
      - "8081:8081"
    environment:
      - CLAUDE_API_KEY=${CLAUDE_API_KEY}
      - DATABASE_URL=postgresql://user:pass@postgres/tribunal
    depends_on:
      - postgres

  tribunal-ts:
    image: tribunal:latest-ts
    ports:
      - "3000:3000"
    environment:
      - GO_SERVICE_URL=http://tribunal-go:8080
      - PYTHON_SERVICE_URL=http://tribunal-python:8081

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_PASSWORD=tribunal_password
      - POSTGRES_DB=tribunal
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./schema/postgres.sql:/docker-entrypoint-initdb.d/schema.sql
    ports:
      - "5432:5432"

  redis:
    image: redis:7
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

Deploy:
```bash
docker-compose up -d
```

---

## Environment Configuration

### Required Variables

```bash
# Platform (github, gitlab, gitea)
PLATFORM=github

# Database
DATABASE_URL=postgresql://user:password@localhost/tribunal

# Redis
REDIS_URL=redis://localhost:6379

# API Keys
CLAUDE_API_KEY=sk-...

# GitHub
GITHUB_APP_ID=123456
GITHUB_APP_PRIVATE_KEY=/path/to/private-key.pem
GITHUB_WEBHOOK_SECRET=whsec_...

# OR GitLab
GITLAB_INSTANCE_URL=https://gitlab.com
GITLAB_BOT_TOKEN=glpat-...
GITLAB_WEBHOOK_SECRET=...

# OR Gitea
GITEA_INSTANCE_URL=https://gitea.example.com
GITEA_CLIENT_ID=...
GITEA_CLIENT_SECRET=...
```

### Optional Variables

```bash
# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Performance
MAX_CONCURRENT_ANALYSES=10
ANALYSIS_TIMEOUT_SECONDS=30

# Features
ENABLE_INCIDENT_CORRELATION=true
ENABLE_CUSTOM_RULES=false

# Notifications
SLACK_WEBHOOK_URL=https://hooks.slack.com/...
EMAIL_ALERTS=true
```

---

## Webhook Payloads

### GitHub Webhook Example

```json
{
  "action": "opened",
  "pull_request": {
    "id": 1,
    "number": 123,
    "title": "Add user authentication",
    "head": {
      "ref": "feature/auth",
      "sha": "abc123def456"
    },
    "base": {
      "ref": "main",
      "sha": "xyz789"
    }
  },
  "repository": {
    "name": "myapp",
    "full_name": "org/myapp",
    "owner": {
      "login": "org"
    }
  }
}
```

### GitLab Webhook Example

```json
{
  "object_kind": "merge_request",
  "action": "open",
  "object_attributes": {
    "id": 1,
    "iid": 123,
    "title": "Add user authentication",
    "source_branch": "feature/auth",
    "target_branch": "main"
  },
  "project": {
    "name": "myapp",
    "path_with_namespace": "org/myapp",
    "owner": {
      "username": "org"
    }
  }
}
```

---

## Troubleshooting

### Webhook Not Firing

1. **Check webhook delivery**: Platform settings → Webhooks → Recent Deliveries
2. **Verify URL is accessible**: `curl https://tribunal.your-domain.com/health`
3. **Check secret matches**: Ensure `WEBHOOK_SECRET` env var matches platform config
4. **Check logs**: `docker logs tribunal-go`

### Analyses Not Running

1. **Check Python service**: `curl http://localhost:8081/health`
2. **Check database**: `psql -U user -d tribunal -c "SELECT 1"`
3. **Check logs**: `docker logs tribunal-python`
4. **Check Claude API key**: Verify `CLAUDE_API_KEY` is valid

### Check Run Not Appearing on PR

1. **Verify permissions**: GitHub App needs "Checks" write permission
2. **Check logs**: Look for permission errors in `tribunal-ts` logs
3. **Test API call**: `curl -H "Authorization: Bearer ..." https://api.github.com/repos/owner/repo/check-runs`

---

## Security Best Practices

### Webhook Security

✅ **Always use HTTPS** for webhook URLs
✅ **Verify webhook signatures** before processing
✅ **Use secrets** for all webhook URLs
✅ **Rate limit** webhook processing (prevent DoS)
✅ **Log all** webhook delivery attempts

### API Token Security

✅ **Store tokens in environment variables**, not code
✅ **Use short-lived tokens** where possible
✅ **Rotate tokens** every 90 days
✅ **Monitor token usage** for suspicious activity
✅ **Use different tokens** for different environments (dev/prod)

### Data Security

✅ **Encrypt database at-rest**
✅ **Use TLS for all connections**
✅ **Don't log sensitive data** (API keys, code diffs)
✅ **Implement access controls** on briefing data
✅ **Audit trail** for all decisions

---

## Support

For integration issues:

1. Check [Troubleshooting](#troubleshooting) section
2. Review logs: `docker logs <service-name>`
3. Check webhook delivery history in platform settings
4. Open a GitHub issue with:
   - Platform (GitHub/GitLab/Gitea)
   - Error messages from logs
   - Steps to reproduce

---

**Ready to integrate?** Start with GitHub, then expand to GitLab/Gitea.
