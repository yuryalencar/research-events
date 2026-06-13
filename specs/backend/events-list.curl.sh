#!/bin/bash
# Events: List
# Spec: specs/backend/events-list.yaml

BASE_URL="${BASE_URL:-http://localhost:8080}"

# 200 — no filters (defaults: status=approved, year=current year, page=1, page_size=20)
echo "==> 200 no filters"
curl -s -X GET "$BASE_URL/api/v1/events" | jq .

# 200 — filter by status=pending (year still defaults to current year)
echo "==> 200 status=pending"
curl -s -X GET "$BASE_URL/api/v1/events?status=pending" | jq .

# 200 — filter by year alone
echo "==> 200 year=2025"
curl -s -X GET "$BASE_URL/api/v1/events?year=2025" | jq .

# 200 — combine domain, country (case-insensitive), tier, first_deadline_month and bbox (AND)
echo "==> 200 combined filters"
curl -s -X GET "$BASE_URL/api/v1/events?year=2026&domain=computer_science&country=brazil&status=approved&tier=A&first_deadline_month=6&bbox=-180,-90,180,90" | jq .

# 200 — pagination: page 2, page_size 10
echo "==> 200 pagination page=2&page_size=10"
curl -s -X GET "$BASE_URL/api/v1/events?page=2&page_size=10" | jq .

# 200 — pagination=off returns every matching row; meta.page=1
echo "==> 200 pagination=off"
curl -s -X GET "$BASE_URL/api/v1/events?pagination=off" | jq .

# 200 — tier=unranked returns events submitted without a tier
echo "==> 200 tier=unranked"
curl -s -X GET "$BASE_URL/api/v1/events?tier=unranked" | jq .

# 200 — no events match filters → data: [] and meta.total: 0
echo "==> 200 empty result set"
curl -s -X GET "$BASE_URL/api/v1/events?year=1900" | jq .

# 400 — ?year=abc (non-numeric)
echo "==> 400 invalid year"
curl -s -X GET "$BASE_URL/api/v1/events?year=abc" | jq .

# 400 — ?status=foo (not in enum)
echo "==> 400 invalid status"
curl -s -X GET "$BASE_URL/api/v1/events?status=foo" | jq .

# 400 — ?domain=foo (not in enum)
echo "==> 400 invalid domain"
curl -s -X GET "$BASE_URL/api/v1/events?domain=medicine" | jq .

# 400 — ?tier=foo (not in enum)
echo "==> 400 invalid tier"
curl -s -X GET "$BASE_URL/api/v1/events?tier=S" | jq .

# 400 — ?first_deadline_month out of range
echo "==> 400 first_deadline_month=13"
curl -s -X GET "$BASE_URL/api/v1/events?first_deadline_month=13" | jq .

# 400 — ?bbox wrong number of values
echo "==> 400 bbox wrong count"
curl -s -X GET "$BASE_URL/api/v1/events?bbox=1,2,3" | jq .

# 400 — ?bbox with minLng >= maxLng
echo "==> 400 bbox inverted range"
curl -s -X GET "$BASE_URL/api/v1/events?bbox=10,-5,-10,5" | jq .

# 400 — ?bbox with a value out of range (lat=95)
echo "==> 400 bbox out of range"
curl -s -X GET "$BASE_URL/api/v1/events?bbox=-10,-5,10,95" | jq .

# 400 — ?page=0 (must be positive)
echo "==> 400 page=0"
curl -s -X GET "$BASE_URL/api/v1/events?page=0" | jq .

# 400 — ?page_size=101 (must be <= 100)
echo "==> 400 page_size=101"
curl -s -X GET "$BASE_URL/api/v1/events?page_size=101" | jq .

# 400 — ?pagination=maybe (must be "on" or "off")
echo "==> 400 invalid pagination value"
curl -s -X GET "$BASE_URL/api/v1/events?pagination=maybe" | jq .

# 429 — rate limit exceeded (requires looping 31+ requests in <60s, burst 30)
echo "==> 429 rate limit exceeded (requires looping 31+ requests in <60s)"
