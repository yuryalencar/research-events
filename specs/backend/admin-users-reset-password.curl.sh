#!/bin/bash
# Admin: Reset User Password
# Spec: specs/backend/admin-users-reset-password.yaml

BASE_URL="${BASE_URL:-http://localhost:8080}"

# 200 — successful password reset (replace TARGET_USER_ID and cookie value)
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/2/password" \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=<your-admin-jwt>" \
  -d '{"new_password":"NewPass@1","new_password_confirmation":"NewPass@1"}' | jq .

# 400 — passwords do not match
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/2/password" \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=<your-admin-jwt>" \
  -d '{"new_password":"NewPass@1","new_password_confirmation":"Different@2"}' | jq .

# 400 — missing field
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/2/password" \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=<your-admin-jwt>" \
  -d '{"new_password":"NewPass@1"}' | jq .

# 422 — password too weak (missing special character)
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/2/password" \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=<your-admin-jwt>" \
  -d '{"new_password":"Weakpass1","new_password_confirmation":"Weakpass1"}' | jq .

# 422 — admin targeting own account
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/1/password" \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=<your-admin-jwt>" \
  -d '{"new_password":"NewPass@1","new_password_confirmation":"NewPass@1"}' | jq .

# 404 — user not found
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/99999/password" \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=<your-admin-jwt>" \
  -d '{"new_password":"NewPass@1","new_password_confirmation":"NewPass@1"}' | jq .

# 401 — unauthenticated (no cookie)
curl -s -X PATCH "$BASE_URL/api/v1/admin/users/2/password" \
  -H "Content-Type: application/json" \
  -d '{"new_password":"NewPass@1","new_password_confirmation":"NewPass@1"}' | jq .
