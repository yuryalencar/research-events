#!/bin/bash
# Events: Supersede a deadline on an approved event
# Spec: specs/backend/events-deadlines-supersede.yaml

BASE_URL="${BASE_URL:-http://localhost:8080}"

# 200 — supersede an active deadline with a new date/time/timezone (replace 1 and 10 with a real approved event id and one of its active deadline ids)
# Response "deadlines" includes the new active deadline AND the just-superseded one
# (is_active=false, superseded_by_id=<new id>).
echo "==> 200 supersede a deadline with date, time, and timezone"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines/10/supersede" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"},
    "date": "2026-11-01",
    "time": "23:59",
    "timezone": "AoE"
  }' | jq .

# 200 — supersede with date only (time/timezone omitted → new deadline has time=null, timezone=null
# even if the old deadline had values — never inherited)
echo "==> 200 supersede a deadline with date only (time/timezone not inherited)"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines/11/supersede" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"},
    "date": "2026-11-15"
  }' | jq .

# 400 — missing submitter name
echo "==> 400 missing submitter name"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines/10/supersede" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "", "email": "carlos@example.com"},
    "date": "2026-11-01"
  }' | jq .

# 400 — invalid submitter email
echo "==> 400 invalid submitter email"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines/10/supersede" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "not-an-email"},
    "date": "2026-11-01"
  }' | jq .

# 400 — missing date
echo "==> 400 missing date"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines/10/supersede" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"},
    "date": ""
  }' | jq .

# 400 — unparsable date
echo "==> 400 unparsable date"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines/10/supersede" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"},
    "date": "01/11/2026"
  }' | jq .

# 400 — invalid time format (not zero-padded / out of range)
echo "==> 400 invalid time format"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines/10/supersede" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"},
    "date": "2026-11-01",
    "time": "24:00"
  }' | jq .

# 400 — explicit empty timezone
echo "==> 400 empty timezone"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines/10/supersede" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"},
    "date": "2026-11-01",
    "timezone": ""
  }' | jq .

# 400 — malformed JSON
echo "==> 400 malformed JSON"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines/10/supersede" \
  -H "Content-Type: application/json" \
  -d '{"submitter": {"name": "Carlos Souza", "email": "carlos@example.com"}' | jq .

# 404 — non-numeric eventId
echo "==> 404 non-numeric eventId"
curl -s -X POST "$BASE_URL/api/v1/events/abc/deadlines/10/supersede" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"},
    "date": "2026-11-01"
  }' | jq .

# 404 — event does not exist
echo "==> 404 event not found"
curl -s -X POST "$BASE_URL/api/v1/events/999999/deadlines/10/supersede" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"},
    "date": "2026-11-01"
  }' | jq .

# 409 — event exists but is not approved (replace 2 with a pending/rejected event id)
echo "==> 409 event not approved"
curl -s -X POST "$BASE_URL/api/v1/events/2/deadlines/10/supersede" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"},
    "date": "2026-11-01"
  }' | jq .

# 404 — non-numeric deadlineId
echo "==> 404 non-numeric deadlineId"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines/abc/supersede" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"},
    "date": "2026-11-01"
  }' | jq .

# 404 — deadline does not exist
echo "==> 404 deadline not found"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines/999999/supersede" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"},
    "date": "2026-11-01"
  }' | jq .

# 404 — deadline belongs to a different event (replace 3 with another approved event id, 10 belongs to event 1)
echo "==> 404 deadline belongs to a different event"
curl -s -X POST "$BASE_URL/api/v1/events/3/deadlines/10/supersede" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"},
    "date": "2026-11-01"
  }' | jq .

# 409 — deadline already inactive (already cancelled or superseded; replace 10 with such a deadline id)
echo "==> 409 deadline already inactive"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines/10/supersede" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"},
    "date": "2026-11-01"
  }' | jq .

# 429 — rate limit exceeded (requires looping 51+ requests in <60s)
echo "==> 429 rate limit exceeded (requires looping 51+ requests in <60s)"
