#!/bin/bash
# Deadlines: time + timezone fields (cross-cutting addition)
# Spec: specs/backend/deadlines-add-time-timezone.yaml

BASE_URL="${BASE_URL:-http://localhost:8080}"

# 201 — submit with a deadline that has both time and timezone
echo "==> 201 submit with deadline time + timezone"
curl -s -X POST "$BASE_URL/api/v1/events/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "International Conference on Model-Driven Engineering",
    "slug": "MODELS2026TZ",
    "country": "Brazil",
    "city": "Recife",
    "latitude": -8.0476,
    "longitude": -34.8770,
    "start_date": "2026-09-21",
    "end_date": "2026-09-25",
    "website_url": "https://models2026.example.org",
    "domain": "computer_science",
    "submitter": {"name": "Ana Silva", "email": "ana@example.com"},
    "deadlines": [
      {"type": "paper", "description": "Research track full paper", "date": "2026-08-22", "time": "23:59", "timezone": "AoE"}
    ]
  }' | jq .

# 201 — add a deadline with only time (timezone omitted)
echo "==> 201 add deadline with time only"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Beatriz Costa", "email": "beatriz@example.com"},
    "deadlines": [
      {"type": "camera_ready", "description": "Research track camera-ready", "date": "2026-10-01", "time": "18:00"}
    ]
  }' | jq .

# 201 — add a deadline with only timezone (time omitted)
echo "==> 201 add deadline with timezone only"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Beatriz Costa", "email": "beatriz@example.com"},
    "deadlines": [
      {"type": "notification", "description": "Industry track notification", "date": "2026-09-12", "timezone": "UTC-3"}
    ]
  }' | jq .

# 400 — time not zero-padded
echo "==> 400 invalid time (not zero-padded)"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Beatriz Costa", "email": "beatriz@example.com"},
    "deadlines": [
      {"type": "paper", "description": "Research track full paper", "date": "2026-08-22", "time": "9:00"}
    ]
  }' | jq .

# 400 — time out of range (24:00)
echo "==> 400 invalid time (24:00)"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Beatriz Costa", "email": "beatriz@example.com"},
    "deadlines": [
      {"type": "paper", "description": "Research track full paper", "date": "2026-08-22", "time": "24:00"}
    ]
  }' | jq .

# 400 — time out of range (23:60)
echo "==> 400 invalid time (23:60)"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Beatriz Costa", "email": "beatriz@example.com"},
    "deadlines": [
      {"type": "paper", "description": "Research track full paper", "date": "2026-08-22", "time": "23:60"}
    ]
  }' | jq .

# 400 — empty string timezone
echo "==> 400 empty timezone string"
curl -s -X POST "$BASE_URL/api/v1/events/1/deadlines" \
  -H "Content-Type: application/json" \
  -d '{
    "submitter": {"name": "Beatriz Costa", "email": "beatriz@example.com"},
    "deadlines": [
      {"type": "paper", "description": "Research track full paper", "date": "2026-08-22", "timezone": ""}
    ]
  }' | jq .

# 200 — list events: pre-migration deadlines show time=null, timezone=null
echo "==> 200 list events (check time/timezone null on pre-migration deadlines)"
curl -s "$BASE_URL/api/v1/events?year=2026" | jq '.data[].deadlines'
