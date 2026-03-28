# Go Interceptor (MVP)

This service exposes a deterministic, local-first analysis API for TRIBUNAL MVP.

## Endpoints

- `GET /health` — health check
- `POST /analyze` — analyze PR file patches
- `POST /webhook/github` — webhook adapter for GitHub payloads (expects `tribunal_files` in payload)

## Run locally

```bash
cd services/go-interceptor
go run .
```

Server starts at `http://localhost:8080`.

## Example analyze request

```json
{
  "repository": "rohanpatel2002/tribunal",
  "prNumber": 42,
  "files": [
    {
      "path": "db/migrations/202603_add_column.sql",
      "status": "modified",
      "patch": "ALTER TABLE users ADD COLUMN active_status TEXT;"
    }
  ]
}
```

## Notes

- The current detector is heuristic-only and deterministic.
- LLM-based semantic analysis is planned for a subsequent phase.
