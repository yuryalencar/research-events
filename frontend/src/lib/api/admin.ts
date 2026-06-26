import { apiPrivateRequest, apiPrivateRequestWithMeta } from "./client"
import type {
  ReviewEventInput,
  ReviewEventResult,
  UnlockUserResult,
  ListAdminUsersParams,
  AdminUserListItem,
  ApiMeta,
  RegisterAdminUserInput,
  RegisterAdminUserResult,
  ChangeUserRoleResult,
  ResetUserPasswordInput,
  ResetUserPasswordResult,
} from "@/types/api"

// --- Public API ---

// reviewEvent approves or rejects a pending event, optionally editing its
// fields in the same request. Admin/moderator only.
async function reviewEvent(id: number, input: ReviewEventInput): Promise<ReviewEventResult> {
  return apiPrivateRequest<ReviewEventResult>(`/api/v1/admin/events/${id}/review`, {
    method: "PATCH",
    body: JSON.stringify(input),
  })
}

// unlockUser clears a locked account's failed-login state. Admin/moderator
// only; a user cannot unlock themselves (enforced by the backend).
async function unlockUser(id: number): Promise<UnlockUserResult> {
  return apiPrivateRequest<UnlockUserResult>(`/api/v1/admin/users/${id}/unlock`, {
    method: "PATCH",
  })
}

// listAdminUsers fetches a paginated, filtered list of users.
// Undefined params are omitted from the query string entirely.
async function listAdminUsers(params: ListAdminUsersParams = {}): Promise<{ data: AdminUserListItem[]; meta: ApiMeta }> {
  const qs = new URLSearchParams()
  if (params.roles?.length) qs.set("roles", params.roles.join(","))
  if (params.search !== undefined) qs.set("search", params.search)
  if (params.locked !== undefined) qs.set("locked", String(params.locked))
  if (params.include_deleted !== undefined) qs.set("include_deleted", String(params.include_deleted))
  if (params.page !== undefined) qs.set("page", String(params.page))
  if (params.page_size !== undefined) qs.set("page_size", String(params.page_size))
  if (params.pagination !== undefined) qs.set("pagination", params.pagination)

  const query = qs.toString()
  const path = query ? `/api/v1/admin/users?${query}` : "/api/v1/admin/users"
  return apiPrivateRequestWithMeta<AdminUserListItem[]>(path)
}

// registerAdminUser creates a new admin or moderator account. Admin only.
async function registerAdminUser(input: RegisterAdminUserInput): Promise<RegisterAdminUserResult> {
  return apiPrivateRequest<RegisterAdminUserResult>("/api/v1/admin/users", {
    method: "POST",
    body: JSON.stringify(input),
  })
}

// changeUserRole updates a user's role and invalidates their session. Admin only.
async function changeUserRole(id: number, role: string): Promise<ChangeUserRoleResult> {
  return apiPrivateRequest<ChangeUserRoleResult>(`/api/v1/admin/users/${id}/role`, {
    method: "PATCH",
    body: JSON.stringify({ role }),
  })
}

// resetUserPassword sets a new password for any user without requiring the
// current password. Invalidates the target user's session. Admin only.
async function resetUserPassword(id: number, input: ResetUserPasswordInput): Promise<ResetUserPasswordResult> {
  return apiPrivateRequest<ResetUserPasswordResult>(`/api/v1/admin/users/${id}/password`, {
    method: "PATCH",
    body: JSON.stringify(input),
  })
}

// --- Export ---

export { reviewEvent, unlockUser, listAdminUsers, registerAdminUser, changeUserRole, resetUserPassword }
