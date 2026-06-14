#!/bin/bash
# Events: Cancel a deadline on an approved event
# Spec: specs/backend/events-deadlines-cancel.yaml

BASE_URL="${BASE_URL:-http://localhost:8080}"

# 200 — cancel an active deadline (replace 1 and 10 with a real approved event id and one of its active deadline ids)
echo "==> 200 cancel an active deadline"
curl -s -X PATCH "$BASE_URL/api/v1/events/1/deadlines/10/cancel" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"}
  }' | jq .

# 200 — cancel the last remaining active deadline (response "deadlines" is an empty array)
echo "==> 200 cancel the last active deadline"
curl -s -X PATCH "$BASE_URL/api/v1/events/1/deadlines/11/cancel" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"}
  }' | jq .

# 400 — missing submitter name
echo "==> 400 missing submitter name"
curl -s -X PATCH "$BASE_URL/api/v1/events/1/deadlines/10/cancel" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "", "email": "carlos@example.com"}
  }' | jq .

# 400 — invalid submitter email
echo "==> 400 invalid submitter email"
curl -s -X PATCH "$BASE_URL/api/v1/events/1/deadlines/10/cancel" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "not-an-email"}
  }' | jq .

# 400 — malformed JSON
echo "==> 400 malformed JSON"
curl -s -X PATCH "$BASE_URL/api/v1/events/1/deadlines/10/cancel" \
  -H "Content-Type: application/json" \
  -d '{"submitter": {"name": "Carlos Souza", "email": "carlos@example.com"}' | jq .

# 404 — non-numeric eventId
echo "==> 404 non-numeric eventId"
curl -s -X PATCH "$BASE_URL/api/v1/events/abc/deadlines/10/cancel" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"}
  }' | jq .

# 404 — event does not exist
echo "==> 404 event not found"
curl -s -X PATCH "$BASE_URL/api/v1/events/999999/deadlines/10/cancel" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"}
  }' | jq .

# 409 — event exists but is not approved (replace 2 with a pending/rejected event id)
echo "==> 409 event not approved"
curl -s -X PATCH "$BASE_URL/api/v1/events/2/deadlines/10/cancel" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"}
  }' | jq .

# 404 — non-numeric deadlineId
echo "==> 404 non-numeric deadlineId"
curl -s -X PATCH "$BASE_URL/api/v1/events/1/deadlines/abc/cancel" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"}
  }' | jq .

# 404 — deadline does not exist
echo "==> 404 deadline not found"
curl -s -X PATCH "$BASE_URL/api/v1/events/1/deadlines/999999/cancel" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"}
  }' | jq .

# 404 — deadline belongs to a different event (replace 3 with another approved event id, 10 belongs to event 1)
echo "==> 404 deadline belongs to a different event"
curl -s -X PATCH "$BASE_URL/api/v1/events/3/deadlines/10/cancel" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"}
  }' | jq .

# 409 — deadline already inactive (already cancelled or superseded; replace 10 with such a deadline id)
echo "==> 409 deadline already inactive"
curl -s -X PATCH "$BASE_URL/api/v1/events/1/deadlines/10/cancel" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Carlos Souza", "email": "carlos@example.com"}
  }' | jq .

# 429 — rate limit exceeded (requires looping 51+ requests in <60s)
echo "==> 429 rate limit exceeded (requires looping 51+ requests in <60s)"
