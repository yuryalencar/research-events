#!/bin/bash
# Auth: Logout
# Spec: specs/backend/auth-logout.yaml
# Note: run auth-login.curl.sh first to populate $COOKIE_JAR

BASE_URL="${BASE_URL:-http://localhost:8080}"
COOKIE_JAR="${COOKIE_JAR:-/tmp/cookies.txt}"

# 200 — valid access token (uses cookies saved by auth-login.curl.sh)
echo "==> 200 successful logout"
curl -sb "$COOKIE_JAR" -sc "$COOKIE_JAR" -X POST "$BASE_URL/api/v1/auth/logout" | jq .

# 200 — idempotent: second logout on already-cleared tokens
echo "==> 200 idempotent second logout"
curl -sb "$COOKIE_JAR" -sc "$COOKIE_JAR" -X POST "$BASE_URL/api/v1/auth/logout" | jq .

# 401 TOKEN_MISSING — no cookie
echo "==> 401 TOKEN_MISSING"
curl -s -X POST "$BASE_URL/api/v1/auth/logout" | jq .

# 401 TOKEN_INVALID — tampered signature
echo "==> 401 TOKEN_INVALID (tampered)"
curl -s -X POST "$BASE_URL/api/v1/auth/logout" \
  --cookie "access_token=header.payload.invalidsignature" | jq .

# 200 — expired token with valid signature (graceful logout)
# Requires a manually crafted expired JWT signed with the correct JWT_SECRET.
# Generate one with: go run ./cmd/tools/gen-expired-token (future CLI tool)
# curl -s -X POST "$BASE_URL/api/v1/auth/logout" \
#   --cookie "access_token=<expired_jwt_here>" | jq .
echo "==> 200 expired token (graceful logout): requires manually crafted expired JWT — see comment above"
