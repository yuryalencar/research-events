#!/bin/bash
# Auth: Refresh Token
# Spec: specs/backend/auth-refresh-token.yaml
# Note: run auth-login.curl.sh first to populate $COOKIE_JAR

BASE_URL="${BASE_URL:-http://localhost:8080}"
COOKIE_JAR="${COOKIE_JAR:-/tmp/cookies.txt}"

# 200 — valid refresh token (uses cookies saved by auth-login.curl.sh)
echo "==> 200 valid refresh (rotates both tokens)"
curl -sb "$COOKIE_JAR" -sc "$COOKIE_JAR" -X POST "$BASE_URL/api/v1/auth/refresh-token" | jq .

# 401 REFRESH_TOKEN_MISSING — no cookie provided
echo "==> 401 REFRESH_TOKEN_MISSING"
curl -s -X POST "$BASE_URL/api/v1/auth/refresh-token" | jq .

# 401 REFRESH_TOKEN_INVALID — tampered/garbage token value
echo "==> 401 REFRESH_TOKEN_INVALID (tampered)"
curl -s -X POST "$BASE_URL/api/v1/auth/refresh-token" \
  --cookie "refresh_token=thisisnotavalidtoken" | jq .

# 401 REFRESH_TOKEN_REUSE — use the old refresh token after rotation
# (run after the 200 case above, which rotated the token in $COOKIE_JAR)
# To test manually: save the old cookie value before running the 200 case, then replay it:
# curl -s -X POST "$BASE_URL/api/v1/auth/refresh-token" \
#   --cookie "refresh_token=<old_token_value>" | jq .
echo "==> 401 REFRESH_TOKEN_REUSE: replay old token manually (see comment above)"
