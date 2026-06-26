# Admin Users List Endpoint

**Date:** 2026-06-26
**Feature:** `GET /api/v1/admin/users`
**Spec:** `specs/backend/admin-users-list.yaml`

---

## Goal

Add a paginated endpoint for admins to list platform users, with filters for role (cumulative OR), name/email search, lock status, and soft-deleted inclusion.

---

## Decisions made

- **`deleted_at` added mid-feature** — original spec omitted it, but the user correctly identified that returning soft-deleted users via `include_deleted=true` is useless without knowing which ones are deleted. Added `deleted_at: string | null` to the response shape and updated the spec.
- **`gorm.DeletedAt` → `*time.Time` mapping** — `gorm.DeletedAt` is a `sql.NullTime` wrapper; `toUserListItemResponse` checks `.Valid` before taking `&u.DeletedAt.Time` to get a clean nullable pointer for JSON.
- **No rate limiter on GET** — listing is a read-only operation, not a brute-force target. The other admin write endpoints (Register, ChangeRole) have the rate limiter; List does not.
- **`ToListUsersFilter` lives in `service/`** — keeps the service layer free of repository coupling in tests. Handler calls it inline: `service.ToListUsersFilter(input)`.
- **`listUsersMux` / `listUsersReq` naming** — `listReq` was already taken by `event_list_test.go` in the same `handler_test` package; user helpers were suffixed with `Users` to avoid the collision.
- **Body double-read bug** — two handler tests initially called `responseCode(t, rec)` then decoded `rec.Body` a second time, getting `EOF`. Fixed by decoding once and asserting `body["code"]` directly.

---

## What was built

| File | Change |
|---|---|
| `specs/backend/admin-users-list.yaml` | New spec (updated mid-session to add `deleted_at`) |
| `specs/backend/admin-users-list.curl.sh` | curl examples for all response codes |
| `internal/service/user_list.go` | `ValidateListUsersQuery` + `ToListUsersFilter` + helpers |
| `internal/service/user_list_test.go` | 24 pure function tests — no mocks |
| `internal/repository/user.go` | `ListUsersFilter` type + `List` method + interface entry |
| `internal/repository/mocks/mock_user.go` | Regenerated |
| `internal/repository/user_test.go` | 14 real-DB tests via transaction rollback |
| `internal/handler/admin_user.go` | `List` handler + `toRawListUsersQuery` + `userListItemResponse` + `toUserListItemResponse` |
| `internal/handler/admin_user_test.go` | 10 handler tests via gomock |
| `cmd/api/server.go` | `GET /api/v1/admin/users` route wired |
| `cmd/api/server_test.go` | Route registration test |

---

## Test counts added this session

| Package | Tests added |
|---|---|
| `service` | 24 |
| `repository` | 14 |
| `handler` | 10 |
| `cmd/api` | 1 |

All 49 new tests pass. Full suite (`go test ./...`) green with zero failures.

---

## State at end

Feature complete and merged into `main`. No open TODOs.

The `GET /api/v1/admin/users` endpoint is live with:
- Role filter (`?roles=admin,moderator` — OR logic)
- Name/email search (`?search=alice` — ILIKE)
- Lock filter (`?locked=true|false`)
- Soft-delete inclusion (`?include_deleted=true`)
- Pagination (`?page=N&page_size=N&pagination=on|off`)
- Response: `id`, `name`, `email`, `role`, `created_at`, `locked_at`, `deleted_at` — no password or token fields
