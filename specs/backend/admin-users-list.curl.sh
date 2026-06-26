#!/bin/bash
# Admin: List Users
# Spec: specs/backend/admin-users-list.yaml

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN="${ADMIN_TOKEN:-}"

# 200 — all users, no filters (default pagination)
curl -s -X GET "$BASE_URL/api/v1/admin/users" \
  -H "Cookie: token=$TOKEN" | jq .

# 200 — filter by single role
curl -s -X GET "$BASE_URL/api/v1/admin/users?roles=admin" \
  -H "Cookie: token=$TOKEN" | jq .

# 200 — filter by multiple roles (OR logic)
curl -s -X GET "$BASE_URL/api/v1/admin/users?roles=admin,moderator" \
  -H "Cookie: token=$TOKEN" | jq .

# 200 — search by partial name or email
curl -s -X GET "$BASE_URL/api/v1/admin/users?search=john" \
  -H "Cookie: token=$TOKEN" | jq .

# 200 — only locked users
curl -s -X GET "$BASE_URL/api/v1/admin/users?locked=true" \
  -H "Cookie: token=$TOKEN" | jq .

# 200 — include soft-deleted users
curl -s -X GET "$BASE_URL/api/v1/admin/users?include_deleted=true" \
  -H "Cookie: token=$TOKEN" | jq .

# 200 — pagination: page 2, 10 per page
curl -s -X GET "$BASE_URL/api/v1/admin/users?page=2&page_size=10" \
  -H "Cookie: token=$TOKEN" | jq .

# 200 — pagination off (all results)
curl -s -X GET "$BASE_URL/api/v1/admin/users?pagination=off" \
  -H "Cookie: token=$TOKEN" | jq .

# 200 — combined filters: locked moderators matching search
curl -s -X GET "$BASE_URL/api/v1/admin/users?roles=moderator&locked=true&search=alice" \
  -H "Cookie: token=$TOKEN" | jq .

# 400 — unknown role value
curl -s -X GET "$BASE_URL/api/v1/admin/users?roles=superuser" \
  -H "Cookie: token=$TOKEN" | jq .

# 400 — invalid page
curl -s -X GET "$BASE_URL/api/v1/admin/users?page=0" \
  -H "Cookie: token=$TOKEN" | jq .

# 400 — page_size exceeds max
curl -s -X GET "$BASE_URL/api/v1/admin/users?page_size=101" \
  -H "Cookie: token=$TOKEN" | jq .

# 401 — no authentication
curl -s -X GET "$BASE_URL/api/v1/admin/users" | jq .

# 403 — moderator attempting access (replace with a moderator token)
curl -s -X GET "$BASE_URL/api/v1/admin/users" \
  -H "Cookie: token=${MODERATOR_TOKEN:-}" | jq .
