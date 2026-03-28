# Context Briefing System

**How TRIBUNAL Generates Context Documents That Save 20 Minutes Per Code Review**

## Overview

The context briefing is TRIBUNAL's core deliverable — a markdown document that shows the human reviewer exactly what the AI didn't know it didn't know.

Instead of spending 30 minutes hunting through incident reports, API docs, and service dependencies, the human reviewer reads a 2-minute briefing that includes:
- What the AI wrote (code summary)
- What context it missed (with evidence)
- Why it matters (impact analysis)
- What should have caught it (learning feedback)

---

## Briefing Structure

Every briefing follows this standard format:

```
# Context Briefing: {PR Title}

## Executive Summary
- Overall risk level
- Number of issues found
- Recommended action

## Critical Findings
[Detailed analysis of high-risk changes]

## Analysis Details
[Service topology, incident history, dependencies]

## Evidence & References
[Links to incidents, documentation, runbooks]

## Recommendation
[Specific action for reviewer]
```

---

## Real-World Example 1: Database Migration

### Context

PR: "Add active_status column to users table"

**Code snippet**:
```sql
ALTER TABLE users ADD COLUMN active_status VARCHAR(50) DEFAULT 'pending';
```

### Generated Briefing

```markdown
# Context Briefing: Add active_status column to users table

## Executive Summary
🚨 **RISK LEVEL: CRITICAL**
- Overall risk: 9/10
- Issues found: 1 critical
- Recommendation: **REQUEST CHANGES** 

---

## Critical Issues

### Issue 1: Scale Blindness in Database Migration
**Location**: db/migrations/v45_users_active_status.sql:1-2

**What the AI wrote**:
```sql
ALTER TABLE users ADD COLUMN active_status VARCHAR(50) DEFAULT 'pending';
```

**What context it missed**:
The users table is one of our most critical resources:
- **Row count**: 2,147,483,647 (2.1 billion)
- **Downtime tolerance**: ZERO (HA SLA: 99.99%)
- **Lock behavior**: `ADD COLUMN ... DEFAULT` blocks entire table during migration
- **Typical lock duration**: 3-4 hours at this scale

**Why this matters**:
- ALTER TABLE acquires exclusive lock on entire table
- Lock holds for duration of column write (default value applied to all rows)
- During lock:
  - All SELECT queries fail
  - All INSERT/UPDATE/DELETE operations fail
  - Login service (reads users) goes down
  - Checkout service (updates users) goes down
  - Authentication service blocks
- **Business impact**: $500k/hour revenue loss
- **Customer impact**: All users locked out

**Similar incident**:
- **Incident #2847**: "Users table migration caused 3-hour outage"
  - Date: 2024-02-15
  - Duration: 3h 22m
  - Impact: All authentication failed
  - Root cause: Schema change with exclusive lock
  - Resolution time: Manual intervention, full migration restart

**Related incidents in past 12 months**:
- #2847 (Feb 2024): 3h outage
- #1923 (Nov 2023): 1h 45m outage
- #1201 (Aug 2023): 2h 10m outage
All caused by similar migration patterns.

---

## What Should Have Caught This

### 1. Service Metadata (Should be Known)
The AI should have known:
- ✗ users table size: 2.1B rows
- ✗ users table SLA: 99.99% uptime
- ✗ users table criticality: P0 (top tier)
- ✗ Recent incidents: 3 in 12 months

### 2. PostgreSQL Constraints (Should be Documented)
From `docs/safe_migrations.md`:
> "Adding a column with a DEFAULT value will lock the entire table during value application. For tables > 1M rows, use the online migration pattern with `ALTER TABLE ... ADD COLUMN ... DEFAULT NULL; UPDATE CONCURRENTLY; ADD CONSTRAINT`"

### 3. Incident Pattern (Should be Correlated)
All 3 outages in past 12 months have same root cause:
- Schema changes on large tables
- Without online migration strategy

---

## Recommended Solution

**Safe migration pattern** (from runbook):

```sql
-- Step 1: Add column with NULL default (no lock)
ALTER TABLE users ADD COLUMN active_status VARCHAR(50);

-- Step 2: Create default in application code
-- (When retrieving user, if active_status is NULL, use 'pending')

-- Step 3: Batch update in chunks (prevents lock)
WITH batch AS (
  SELECT user_id FROM users 
  WHERE active_status IS NULL 
  LIMIT 100000
)
UPDATE users SET active_status = 'pending' 
WHERE user_id IN (SELECT user_id FROM batch);

-- Run Step 3 multiple times over several hours

-- Step 4: Add constraint (when 99.9% populated)
ALTER TABLE users ADD CONSTRAINT active_status_not_null 
  NOT NULL;
```

**Why this works**:
- Step 1: No lock (NULL default is instant)
- Step 2: Backwards compatible (app handles NULL)
- Step 3: Batch updates prevent long locks
- Step 4: Incremental, can be rolled back

**Timeline**:
- Deploy Step 1: 30 seconds
- Run Step 3 batches: 2-4 hours (can be paused/resumed)
- Deploy Step 4: 30 seconds
- Total downtime: 0 minutes

---

## See Also

- **Runbook**: [Safe Database Migrations](https://internal.wiki/runbooks/safe_migrations)
- **Template**: [Migration Safety Checklist](https://internal.wiki/templates/migration_checklist)
- **Incident #2847**: [Root Cause Analysis](https://internal.incidents/2847)
- **Architecture Docs**: [PostgreSQL at Scale](https://internal.docs/architecture/postgres_scale)

---

## Recommendation

✋ **REQUEST CHANGES**

**Reason**: 
This migration will cause a critical outage (estimated 3+ hours of complete service downtime). The safe migration pattern should be followed instead.

**Action items for engineer**:
1. [ ] Review safe migration pattern (link above)
2. [ ] Rewrite migration using `ALTER ... ADD COLUMN ... DEFAULT NULL` pattern
3. [ ] Implement batch update process (can run over 2-4 hours)
4. [ ] Add migration safety checklist (link above)
5. [ ] Re-submit for review

**Approver note**: Once migration follows safe pattern, this should be quick to approve.

---

**Generated by TRIBUNAL**
`Analyzed: 2026-03-28 14:32:15 UTC`
`AI Detection Score: 0.87 (high confidence AI-generated)`
`Analysis Time: 4.2 seconds`

---
```

---

## Real-World Example 2: Idempotency Blindness

### Context

PR: "Add retry logic for payment processing"

**Code snippet**:
```go
func chargeUserWithRetry(ctx context.Context, userID string, amount float64) error {
    var lastErr error
    for attempt := 0; attempt < 3; attempt++ {
        err := paymentService.Charge(userID, amount)
        if err == nil {
            return nil
        }
        lastErr = err
        time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
    }
    return lastErr
}
```

### Generated Briefing

```markdown
# Context Briefing: Add retry logic for payment processing

## Executive Summary
🚨 **RISK LEVEL: CRITICAL**
- Overall risk: 9.5/10  
- Issues found: 1 critical (duplicate charges)
- Recommendation: **BLOCK MERGE**

---

## Critical Issues

### Issue 1: Idempotency Blindness in Payment Retry Logic
**Location**: payment/client.go:42-52

**What the AI wrote**:
Simple exponential backoff retry logic that calls `paymentService.Charge()` up to 3 times.

**What context it missed**:
The PaymentService.Charge() endpoint is **NON-IDEMPOTENT**.

From API documentation:
```
POST /charge
- Required header: Idempotency-Key (UUID)
- If missing: NOT idempotent (may charge twice if called twice)
- If present: idempotent (duplicate calls return same result)
```

**Why this matters**:
- Network timeout occurs after Charge() succeeds but before response returns
- Retry logic calls Charge() again (thinking first call failed)
- Without Idempotency-Key, system charges user TWICE
- **Business impact**: Duplicate charges → chargebacks → customer refunds → trust loss

**Real incident**:
- **Incident #1847**: "Double-charged 847 customers"
  - Date: 2024-01-10
  - Root cause: Retry logic without idempotency key
  - Duration: 2 hours to detect
  - Impact: 847 customer refunds ($24,000 total)
  - Resolution: Manual refund + API key validation added

---

## What Should Have Caught This

### 1. API Contract (Should be Documented)
From `apis/payment_service.yaml`:
```yaml
/charge:
  post:
    parameters:
      - name: Idempotency-Key
        required: true
        description: "UUID. REQUIRED for safety. Absence = non-idempotent."
    responses:
      200:
        description: "Successfully charged. Safe to retry."
```

### 2. Incident Pattern (Should be Correlated)
From incident database:
- **Incident #1847** (Jan 2024): Double charges, retry logic without key
- **Incident #2121** (Aug 2023): Triple charges, similar pattern
- Pattern: 100% of double-charge incidents in past 2 years caused by missing Idempotency-Key

### 3. Code Pattern (Should be Flagged)
This is a known unsafe pattern:
- Retry logic without idempotency key
- Non-idempotent API
- Pattern matches 5 previous fixes in codebase

---

## Recommended Solution

```go
import "github.com/google/uuid"

func chargeUserWithRetry(ctx context.Context, userID string, amount float64) error {
    // Generate unique idempotency key ONCE per charge attempt
    idempotencyKey := uuid.New().String()
    
    var lastErr error
    for attempt := 0; attempt < 3; attempt++ {
        // Pass idempotency key to service
        err := paymentService.ChargeWithIdempotency(
            ctx,
            userID, 
            amount,
            idempotencyKey,  // ← Same key for all retries
        )
        if err == nil {
            return nil
        }
        
        // Only retry on network errors, not validation errors
        if !isNetworkError(err) {
            return err
        }
        
        lastErr = err
        time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
    }
    return lastErr
}
```

**Why this works**:
- Generate idempotency key ONCE at start
- Use same key for all retries
- PaymentService guarantees same response if idempotency key matches
- No duplicate charges even with multiple retries
- Network timeouts handled safely

---

## See Also

- **API Docs**: [Payment Service API](https://internal.docs/apis/payment_service)
- **Incident #1847**: [Root Cause Analysis](https://internal.incidents/1847)
- **Best Practice**: [Retry Patterns with Non-Idempotent APIs](https://internal.docs/best-practices/retry_patterns)
- **Code Example**: [Safe Payment Retry](https://github.com/org/payments/blob/main/client.go#L123)

---

## Recommendation

🛑 **BLOCK MERGE** 

This code will cause duplicate charges in production. This is a known incident pattern (happened 2x in 12 months) with significant financial impact.

**Must fix before merge**:
1. [ ] Add Idempotency-Key header
2. [ ] Generate key ONCE per charge attempt
3. [ ] Pass same key to all retries
4. [ ] Test with network timeout simulation

---

**Generated by TRIBUNAL**
`AI Detection Score: 0.92 (very high confidence)`
`Analysis Time: 3.7 seconds`

---
```

---

## Briefing Generation Process

### Step 1: Parse Code Changes
```
Input: Diff from PR
├─ Extract changed functions
├─ Identify high-risk patterns
└─ Queue for analysis
```

### Step 2: Query Context
```
For each risky section:
├─ Load service topology (changed services)
├─ Query incident history (last 12 months)
├─ Extract API contracts (OpenAPI, Swagger)
├─ Find related code patterns
└─ Search documentation
```

### Step 3: Claude Analysis
```
Claude API call with:
├─ Code snippet
├─ Service context
├─ Incident patterns
├─ API constraints
└─ Documentation references

Claude responds with:
├─ Risk identification
├─ Context blindness analysis
├─ Recommended solution
└─ Evidence links
```

### Step 4: Synthesize Briefing
```
Combine:
├─ Code analysis from Claude
├─ Incident correlations from database
├─ Related documentation links
├─ Similar patterns from codebase
└─ Severity scoring

Output: Markdown briefing
```

### Step 5: Post to PR
```
TypeScript service:
├─ Creates GitHub check run
├─ Posts briefing as comment
├─ Adds inline annotations
└─ Updates summary
```

---

## Performance Metrics

| Metric | Value |
|--------|-------|
| Time to generate briefing | 4-8 seconds |
| Number of evidence items per briefing | 5-15 |
| Average briefing length | 2-4 KB markdown |
| Human reading time | 2-5 minutes |
| Time saved vs. manual research | 15-25 minutes |

---

This system ensures every code review is informed, fast, and backed by evidence.
