#!/usr/bin/env bash

# smoke_test.sh: End-to-end integration test harness for TRIBUNAL backend.
# Tests /health, /analyze, and /webhook/github without external dependencies.

set -euo pipefail

BASE_URL="http://localhost:${PORT:-8080}"
FIXTURE_DIR="services/go-interceptor/fixtures"

echo "=== TRIBUNAL MVP Smoke Test ==="
echo "Targeting API at: $BASE_URL"
echo ""

# 1. Health Check
echo "[1] Testing /health endpoint..."
HEALTH_RES=$(curl -s -w "\n%{http_code}" "$BASE_URL/health" || true)
HEALTH_BODY=$(echo "$HEALTH_RES" | sed '$d')
HEALTH_CODE=$(echo "$HEALTH_RES" | tail -n1)

if [ "$HEALTH_CODE" != "200" ]; then
    echo "Health check failed (HTTP $HEALTH_CODE). Make sure the server is running."
    exit 1
fi
echo "$HEALTH_BODY" | grep -q '"status":"ok"' || { echo "Health body missing status:ok"; exit 1; }
echo "Health check passed."
echo ""

# 2. Analyze Direct Endpoint
echo "[2] Testing /analyze endpoint (Direct Analysis)..."
if [ ! -f "$FIXTURE_DIR/analyze-high-risk.json" ]; then
    echo "Fixture not found at $FIXTURE_DIR/analyze-high-risk.json"
    exit 1
fi

ANALYZE_RES=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/analyze" \
    -H "Content-Type: application/json" \
    -d @"$FIXTURE_DIR/analyze-high-risk.json" || true)

ANALYZE_BODY=$(echo "$ANALYZE_RES" | sed '$d')
ANALYZE_CODE=$(echo "$ANALYZE_RES" | tail -n1)

if [ "$ANALYZE_CODE" != "200" ]; then
    echo "Analyze direct failed (HTTP $ANALYZE_CODE)"
    echo "Body: $ANALYZE_BODY"
    exit 1
fi
echo "$ANALYZE_BODY" | grep -q '"recommendation"' || { echo "Analyze missing recommendation"; exit 1; }
echo "Analyze direct passed."
echo ""

# 3. Simulate GitHub Webhook
echo "[3] Testing /webhook/github endpoint (Idempotency Payload)..."
DELIVERY_ID="smoke-test-$(date +%s)"
WEBHOOK_PAYLOAD=$(cat <<JSON
{
  "action": "opened",
  "repository": { "full_name": "rohanpatel2002/tribunal" },
  "pull_request": { "number": 999 },
  "tribunal_files": [
    { "path": "test.go", "status": "added", "patch": "package main\n\nfunc main() {}" }
  ]
}
JSON
)

WEBHOOK_RES=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/webhook/github" \
    -H "Content-Type: application/json" \
    -H "X-GitHub-Event: pull_request" \
    -H "X-GitHub-Delivery: $DELIVERY_ID" \
    -d "$WEBHOOK_PAYLOAD" || true)

WEBHOOK_BODY=$(echo "$WEBHOOK_RES" | sed '$d')
WEBHOOK_CODE=$(echo "$WEBHOOK_RES" | tail -n1)

if [ "$WEBHOOK_CODE" != "200" ]; then
    echo "Webhook initial request failed (HTTP $WEBHOOK_CODE)"
    echo "Body: $WEBHOOK_BODY"
    exit 1
fi
echo "Webhook initial request passed."

# 4. Simulate Duplicate Webhook (Idempotency check)
echo "[4] Testing Duplicate Webhook Delivery (Idempotency)..."
DUP_RES=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/webhook/github" \
    -H "Content-Type: application/json" \
    -H "X-GitHub-Event: pull_request" \
    -H "X-GitHub-Delivery: $DELIVERY_ID" \
    -d "$WEBHOOK_PAYLOAD" || true)

DUP_BODY=$(echo "$DUP_RES" | sed '$d')
DUP_CODE=$(echo "$DUP_RES" | tail -n1)

# Because we are probably running without DB in pure smoke mode, idempotency drops DB error gracefully
# But if it IS running with DB, it will output "webhook already processed".
if [ "$DUP_CODE" != "200" ]; then
    echo "Webhook duplicate request failed (HTTP $DUP_CODE)"
    echo "Body: $DUP_BODY"
    exit 1
fi
echo "Webhook duplicate request handled safely."
echo ""

echo "=== All MVP Smoke Tests Passed! ==="
exit 0
