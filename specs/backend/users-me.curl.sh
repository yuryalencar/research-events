#!/bin/bash
# Users: Get current session user
# Spec: specs/backend/users-me.yaml

BASE_URL="${BASE_URL:-http://localhost:8080}"

# 200 — valid session (access_token cookie must be set; log in first)
curl -s -X GET "$BASE_URL/api/v1/users/me" \
  --cookie "access_token=<your_token_here>" | jq .

# 401 TOKEN_MISSING — no cookie
curl -s -X GET "$BASE_URL/api/v1/users/me" | jq .

# 401 TOKEN_EXPIRED — present via refresh attempt (set an expired token)
# curl -s -X GET "$BASE_URL/api/v1/users/me" \
#   --cookie "access_token=<expired_token>" | jq .
