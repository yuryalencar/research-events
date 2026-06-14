#!/bin/bash
# Admin: Review (approve/reject) an event
# Spec: specs/backend/admin-events-review.yaml
# Note: run auth-login.curl.sh first to populate $COOKIE_JAR with an admin or moderator token

BASE_URL="${BASE_URL:-http://localhost:8080}"
COOKIE_JAR="${COOKIE_JAR:-/tmp/cookies.txt}"

# 200 — approve a pending event, no field edits
echo "==> 200 approve event id=1 (no edits)"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/events/1/review" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve"}' | jq .

# 200 — approve with an optional reviewer reason and field edits (typo fix + recompute year)
echo "==> 200 approve event id=2 with reason and event field edits"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/events/2/review" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "approve",
    "reason": "Fixed typo in conference name and corrected start date",
    "event": {
      "name": "International Conference on Software Engineering",
      "start_date": "2026-05-18"
    }
  }' | jq .

# 200 — reject an event with a required reason
echo "==> 200 reject event id=3 with reason"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/events/3/review" \
  -H "Content-Type: application/json" \
  -d '{"action": "reject", "reason": "Duplicate of an already-approved event"}' | jq .

# 200 — re-review: flip a previously rejected event back to approved
echo "==> 200 re-review event id=3 back to approved"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/events/3/review" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve", "reason": "New evidence — not actually a duplicate"}' | jq .

# 400 — missing/invalid action
echo "==> 400 VALIDATION_ERROR invalid action"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/events/1/review" \
  -H "Content-Type: application/json" \
  -d '{"action": "approveee"}' | jq .

# 400 — reject without a reason
echo "==> 400 VALIDATION_ERROR missing reason on reject"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/events/1/review" \
  -H "Content-Type: application/json" \
  -d '{"action": "reject"}' | jq .

# 400 — invalid edited field (latitude out of range)
echo "==> 400 VALIDATION_ERROR invalid event.latitude"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/events/1/review" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve", "event": {"latitude": 999}}' | jq .

# 400 — edited end_date before start_date
echo "==> 400 VALIDATION_ERROR end_date before start_date"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/events/1/review" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve", "event": {"end_date": "2020-01-01"}}' | jq .

# 400 — malformed JSON
echo "==> 400 VALIDATION_ERROR malformed JSON"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/events/1/review" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve"' | jq .

# 400 — non-numeric :id
echo "==> 400 VALIDATION_ERROR non-integer id"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/events/abc/review" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve"}' | jq .

# 401 — no cookie
echo "==> 401 TOKEN_MISSING"
curl -s -X PATCH "$BASE_URL/api/v1/admin/events/1/review" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve"}' | jq .

# 403 — contributor (or any non-admin/moderator) role
echo "==> 403 FORBIDDEN contributor role"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/events/1/review" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve"}' | jq .

# 403 — moderator reviewing an event they submitted themselves (requires moderator
# cookie in jar and an event whose created_by_id matches that moderator's user id)
echo "==> 403 CANNOT_REVIEW_OWN_EVENT moderator reviewing own submission"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/events/4/review" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve"}' | jq .

# 404 — event not found
echo "==> 404 EVENT_NOT_FOUND"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/events/999999/review" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve"}' | jq .

# 409 — edited slug collides with another pending/approved event's slug
echo "==> 409 SLUG_ALREADY_EXISTS"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/events/1/review" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve", "event": {"slug": "another-events-existing-slug"}}' | jq .
