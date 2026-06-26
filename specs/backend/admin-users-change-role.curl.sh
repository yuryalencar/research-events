#!/bin/bash
# Admin: Change User Role
# Spec: specs/backend/admin-users-change-role.yaml

BASE_URL="${BASE_URL:-http://localhost:8080}"
COOKIE="${COOKIE:-}"  # set to your admin session cookie, e.g. COOKIE="access_token=<jwt>"

# 200 — promote contributor to moderator
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/42/role" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"role":"moderator"}' | jq .

# 200 — promote contributor to admin
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/42/role" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"role":"admin"}' | jq .

# 200 — downgrade moderator to contributor
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/43/role" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"role":"contributor"}' | jq .

# 200 — change another admin's role to moderator
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/44/role" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"role":"moderator"}' | jq .

# 400 — non-integer :id
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/abc/role" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"role":"moderator"}' | jq .

# 400 — invalid role value
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/42/role" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"role":"superuser"}' | jq .

# 403 — caller is a moderator
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/42/role" \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=<moderator_jwt>" \
  -d '{"role":"moderator"}' | jq .

# 404 — user not found
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/99999/role" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"role":"moderator"}' | jq .

# 409 — user already has this role
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/42/role" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"role":"contributor"}' | jq .

# 422 — admin attempting to change own role (replace 1 with the admin's own ID)
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/1/role" \
  -H "Content-Type: application/json" \
  -H "Cookie: $COOKIE" \
  -d '{"role":"moderator"}' | jq .
