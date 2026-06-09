#!/bin/bash
# Auth: Login
# Spec: specs/backend/auth-login.yaml

BASE_URL="${BASE_URL:-http://localhost:8080}"
COOKIE_JAR="${COOKIE_JAR:-/tmp/cookies.txt}"

# 200 — successful admin login (saves cookies for use by other scripts)
echo "==> 200 admin login"
curl -sc "$COOKIE_JAR" -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"secret"}' | jq .

# 200 — successful moderator login
echo "==> 200 moderator login"
curl -sc "$COOKIE_JAR" -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"moderator@example.com","password":"secret"}' | jq .

# 400 — missing email
echo "==> 400 missing email"
curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"password":"secret"}' | jq .

# 400 — missing password
echo "==> 400 missing password"
curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com"}' | jq .

# 401 — wrong password (same response shape as non-existent email — never reveals existence)
echo "==> 401 wrong password"
curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"wrongpassword"}' | jq .

# 401 — non-existent email (same shape as wrong password)
echo "==> 401 non-existent email"
curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"ghost@example.com","password":"secret"}' | jq .

# 403 — contributor attempting login
echo "==> 403 contributor login"
curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"contributor@example.com","password":"secret"}' | jq .

# 423 — locked account (run after 5 failed attempts on the same account)
echo "==> 423 locked account"
curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"locked@example.com","password":"anypassword"}' | jq .
