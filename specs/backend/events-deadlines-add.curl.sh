#!/bin/bash
# Events: Add deadlines to an approved event
# Spec: specs/backend/events-deadlines-add.yaml

BASE_URL="${BASE_URL:-http://localhost:8080}"

# 201 — add a single deadline to an approved event (replace 1 with a real approved event id)
echo "==> 201 add one deadline"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Beatriz Costa", "email": "beatriz@example.com"},
    "deadlines": [
      {"type": "camera_ready", "description": "Research track camera-ready", "date": "2026-10-01", "is_optional": false}
    ]
  }' | jq .

# 201 — add multiple deadlines in one request (writes a single batch_deadlines_added audit row)
echo "==> 201 add multiple deadlines (batch)"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"},
    "deadlines": [
      {"type": "paper", "description": "Industry track full paper", "date": "2026-08-29", "is_optional": false},
      {"type": "notification", "description": "Industry track notification", "date": "2026-09-12", "is_optional": false}
    ]
  }' | jq .

# 400 — empty deadlines array
echo "==> 400 empty deadlines array"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Beatriz Costa", "email": "beatriz@example.com"},
    "deadlines": []
  }' | jq .

# 400 — deadline entry missing description
echo "==> 400 deadline missing description"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Beatriz Costa", "email": "beatriz@example.com"},
    "deadlines": [{"type": "paper", "date": "2026-08-22"}]
  }' | jq .

# 400 — deadline entry with invalid type
echo "==> 400 deadline invalid type"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Beatriz Costa", "email": "beatriz@example.com"},
    "deadlines": [{"type": "keynote", "description": "Keynote announcement", "date": "2026-08-22"}]
  }' | jq .

# 400 — invalid submitter email
echo "==> 400 invalid submitter email"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Beatriz Costa", "email": "not-an-email"},
    "deadlines": [{"type": "paper", "description": "Research track full paper", "date": "2026-08-22"}]
  }' | jq .

# 400 — malformed JSON
echo "==> 400 malformed JSON"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines" \
  -H "Content-Type: application/json" \
  -d '{"submitter": {"name": "Beatriz Costa", "email": "beatriz@example.com"}' | jq .

# 404 — non-numeric id
echo "==> 404 non-numeric id"
curl -s -X POST "$BASE_URL/api/v1/events/abc/deadlines" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Beatriz Costa", "email": "beatriz@example.com"},
    "deadlines": [{"type": "paper", "description": "Research track full paper", "date": "2026-08-22"}]
  }' | jq .

# 404 — event does not exist
echo "==> 404 event not found"
curl -s -X POST "$BASE_URL/api/v1/events/999999/deadlines" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Beatriz Costa", "email": "beatriz@example.com"},
    "deadlines": [{"type": "paper", "description": "Research track full paper", "date": "2026-08-22"}]
  }' | jq .

# 409 — event exists but is not approved (replace 2 with a pending/rejected event id)
echo "==> 409 event not approved"
curl -s -X POST "$BASE_URL/api/v1/events/2/deadlines" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Beatriz Costa", "email": "beatriz@example.com"},
    "deadlines": [{"type": "paper", "description": "Research track full paper", "date": "2026-08-22"}]
  }' | jq .

# 429 — rate limit exceeded (requires looping 51+ requests in <60s)
echo "==> 429 rate limit exceeded (requires looping 51+ requests in <60s)"
