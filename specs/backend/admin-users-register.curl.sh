#!/bin/bash
# Admin: Register User
# Spec: specs/backend/admin-users-register.yaml

BASE_URL="${BASE_URL:-http://localhost:8080}"
COOKIE="${COOKIE:-}"  # set to your admin session cookie, e.g. COOKIE="access_token=<jwt>"

# 201 — register a moderator
curl -s -X POST "$BASE_URL/api/v1/admin/users" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"name":"Jane Doe","email":"jane@example.com","password":"Secret@123","role":"moderator"}' | jq .

# 201 — register another admin
curl -s -X POST "$BASE_URL/api/v1/admin/users" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"name":"John Admin","email":"john@example.com","password":"Secret@123","role":"admin"}' | jq .

# 400 — missing name
curl -s -X POST "$BASE_URL/api/v1/admin/users" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"email":"jane@example.com","password":"Secret@123","role":"moderator"}' | jq .

# 400 — role is "contributor" (not allowed via this endpoint)
curl -s -X POST "$BASE_URL/api/v1/admin/users" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"name":"Jane Doe","email":"jane@example.com","password":"Secret@123","role":"contributor"}' | jq .

# 400 — invalid role value
curl -s -X POST "$BASE_URL/api/v1/admin/users" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"name":"Jane Doe","email":"jane@example.com","password":"Secret@123","role":"superuser"}' | jq .

# 400 — weak password (no special character)
curl -s -X POST "$BASE_URL/api/v1/admin/users" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"name":"Jane Doe","email":"jane@example.com","password":"Password1","role":"moderator"}' | jq .

# 403 — caller is a moderator (not an admin)
curl -s -X POST "$BASE_URL/api/v1/admin/users" \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=<moderator_jwt>" \
  -d '{"name":"Jane Doe","email":"jane@example.com","password":"Secret@123","role":"moderator"}' | jq .

# 409 — email already exists
curl -s -X POST "$BASE_URL/api/v1/admin/users" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"name":"Duplicate","email":"existing@example.com","password":"Secret@123","role":"moderator"}' | jq .
