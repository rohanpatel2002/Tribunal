#!/bin/bash
echo "🚀 Simulating a Live GitHub Pull Request Webhook..."

curl -X POST http://localhost:8080/webhook/github \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Delivery: test-e2e-delivery-$(date +%s)" \
  -d '{
    "action": "opened",
    "repository": { "full_name": "rohanpatel2002/tribunal" },
    "pull_request": { "number": 99, "head": { "sha": "abcdef123" } },
    "tribunal_files": [
      {
        "path": "auth/bypass.go",
        "status": "added",
        "patch": "func login(user string) bool {\n  if user == \"admin\" { return true }\n  return false\n}"
      }
    ]
  }'

echo -e "\n✅ Payload Delivered! Check the dashboard analytics page."
