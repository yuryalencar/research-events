#!/bin/bash
# Users: Update own password
# Spec: specs/backend/users-update-password.yaml

BASE_URL="${BASE_URL:-http://localhost:8080}"

# 200 — password updated successfully (JWT cookie must be present)
curl -s -X PATCH "$BASE_URL/api/v1/users/me/password" \
  -H "Content-Type: application/json" \
  --cookie "token=<your_jwt_here>" \
  -d '{"current_password":"OldPass@1","new_password":"NewPass@2","new_password_confirmation":"NewPass@2"}' | jq .

# 400 — missing required field
curl -s -X PATCH "$BASE_URL/api/v1/users/me/password" \
  -H "Content-Type: application/json" \
  --cookie "token=<your_jwt_here>" \
  -d '{"current_password":"OldPass@1","new_password":"NewPass@2"}' | jq .

# 400 — new_password and new_password_confirmation do not match
curl -s -X PATCH "$BASE_URL/api/v1/users/me/password" \
  -H "Content-Type: application/json" \
  --cookie "token=<your_jwt_here>" \
  -d '{"current_password":"OldPass@1","new_password":"NewPass@2","new_password_confirmation":"Different@3"}' | jq .

# 400 — new_password too short (fewer than 8 characters)
curl -s -X PATCH "$BASE_URL/api/v1/users/me/password" \
  -H "Content-Type: application/json" \
  --cookie "token=<your_jwt_here>" \
  -d '{"current_password":"OldPass@1","new_password":"Ab@1","new_password_confirmation":"Ab@1"}' | jq .

# 400 — new_password missing uppercase letter
curl -s -X PATCH "$BASE_URL/api/v1/users/me/password" \
  -H "Content-Type: application/json" \
  --cookie "token=<your_jwt_here>" \
  -d '{"current_password":"OldPass@1","new_password":"newpass@2","new_password_confirmation":"newpass@2"}' | jq .

# 400 — new_password missing lowercase letter
curl -s -X PATCH "$BASE_URL/api/v1/users/me/password" \
  -H "Content-Type: application/json" \
  --cookie "token=<your_jwt_here>" \
  -d '{"current_password":"OldPass@1","new_password":"NEWPASS@2","new_password_confirmation":"NEWPASS@2"}' | jq .

# 400 — new_password missing special character
curl -s -X PATCH "$BASE_URL/api/v1/users/me/password" \
  -H "Content-Type: application/json" \
  --cookie "token=<your_jwt_here>" \
  -d '{"current_password":"OldPass@1","new_password":"NewPass12","new_password_confirmation":"NewPass12"}' | jq .

# 400 — current_password is incorrect
curl -s -X PATCH "$BASE_URL/api/v1/users/me/password" \
  -H "Content-Type: application/json" \
  --cookie "token=<your_jwt_here>" \
  -d '{"current_password":"WrongPass@9","new_password":"NewPass@2","new_password_confirmation":"NewPass@2"}' | jq .

# 401 — no JWT cookie (unauthenticated)
curl -s -X PATCH "$BASE_URL/api/v1/users/me/password" \
  -H "Content-Type: application/json" \
  -d '{"current_password":"OldPass@1","new_password":"NewPass@2","new_password_confirmation":"NewPass@2"}' | jq .
