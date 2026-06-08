#!/bin/bash
# Server Bootstrap + Health Check
# Spec: specs/backend/server-bootstrap.yaml

BASE_URL="${BASE_URL:-http://localhost:8080}"

# 200 — all checks healthy
curl -s -X GET "$BASE_URL/health" | jq .

# 503 — simulate unhealthy: stop Postgres and re-run
# docker compose stop db
# curl -s -X GET "$BASE_URL/health" | jq .

# 405 — wrong method
curl -s -X POST "$BASE_URL/health" | jq .

# CORS preflight — OPTIONS returns 204 with Allow-Origin header
curl -s -o /dev/null -w "HTTP %{http_code}\n" \
  -X OPTIONS "$BASE_URL/health" \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET"

# CORS — Access-Control-Allow-Origin present on normal GET
curl -s -I -X GET "$BASE_URL/health" \
  -H "Origin: http://localhost:3000" | grep -i "access-control"
