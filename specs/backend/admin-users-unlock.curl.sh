#!/bin/bash
# Admin: Unlock User Account
# Spec: specs/backend/admin-users-unlock.yaml
# Note: run auth-login.curl.sh first to populate $COOKIE_JAR with an admin token

BASE_URL="${BASE_URL:-http://localhost:8080}"
COOKIE_JAR="${COOKIE_JAR:-/tmp/cookies.txt}"

# 200 — admin unlocks a locked user
echo "==> 200 unlock user id=2"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/users/2/unlock" | jq .

# 400 — non-integer :id
echo "==> 400 VALIDATION_ERROR non-integer id"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/users/abc/unlock" | jq .

# 403 — moderator attempting unlock (requires moderator cookie in jar)
echo "==> 403 FORBIDDEN moderator attempt"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/users/2/unlock" | jq .

# 404 — user not found
echo "==> 404 USER_NOT_FOUND"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/users/99999/unlock" | jq .

# 409 — user exists but is not locked
echo "==> 409 USER_NOT_LOCKED"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/users/3/unlock" | jq .

# 422 — admin attempting to unlock themselves (replace 1 with the admin's own user ID)
echo "==> 422 CANNOT_UNLOCK_SELF"
curl -sb "$COOKIE_JAR" -X PATCH "$BASE_URL/api/v1/admin/users/1/unlock" | jq .

# 401 TOKEN_MISSING — no cookie
echo "==> 401 TOKEN_MISSING"
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/2/unlock" | jq .
