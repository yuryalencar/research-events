import { describe, it, expect, vi } from "vitest"

import { apiPrivateRequest, apiPrivateRequestWithMeta } from "./client"
import { reviewEvent, unlockUser, listAdminUsers, registerAdminUser, changeUserRole, resetUserPassword } from "./admin"
import type {
  ReviewEventInput,
  ReviewEventResult,
  UnlockUserResult,
  AdminUserListItem,
  RegisterAdminUserInput,
  RegisterAdminUserResult,
  ChangeUserRoleResult,
  ResetUserPasswordResult,
} from "@/types/api"

vi.mock("./client", () => ({
  apiPrivateRequest: vi.fn(),
  apiPrivateRequestWithMeta: vi.fn(),
}))

describe("reviewEvent", () => {
  it("patches /api/v1/admin/events/{id}/review via apiPrivateRequest", async () => {
    const input: ReviewEventInput = { action: "approve" }
    const result = {} as ReviewEventResult
    vi.mocked(apiPrivateRequest).mockResolvedValue(result)

    const data = await reviewEvent(1, input)

    expect(apiPrivateRequest).toHaveBeenCalledWith("/api/v1/admin/events/1/review", {
      method: "PATCH",
      body: JSON.stringify(input),
    })
    expect(data).toEqual(result)
  })
})

describe("unlockUser", () => {
  it("patches /api/v1/admin/users/{id}/unlock via apiPrivateRequest", async () => {
    const result = {} as UnlockUserResult
    vi.mocked(apiPrivateRequest).mockResolvedValue(result)

    const data = await unlockUser(1)

    expect(apiPrivateRequest).toHaveBeenCalledWith("/api/v1/admin/users/1/unlock", {
      method: "PATCH",
    })
    expect(data).toEqual(result)
  })
})

describe("listAdminUsers", () => {
  it("gets /api/v1/admin/users with all provided params serialised into the query string", async () => {
    const result = { data: [] as AdminUserListItem[], meta: { page: 1, total: 0 } }
    vi.mocked(apiPrivateRequestWithMeta).mockResolvedValue(result)

    const data = await listAdminUsers({
      roles: ["admin", "moderator"],
      search: "alice",
      locked: true,
      include_deleted: false,
      page: 2,
      page_size: 10,
    })

    expect(apiPrivateRequestWithMeta).toHaveBeenCalledWith(
      "/api/v1/admin/users?roles=admin%2Cmoderator&search=alice&locked=true&include_deleted=false&page=2&page_size=10",
    )
    expect(data).toEqual(result)
  })

  it("omits params that are not provided from the query string", async () => {
    const result = { data: [] as AdminUserListItem[], meta: { page: 1, total: 0 } }
    vi.mocked(apiPrivateRequestWithMeta).mockResolvedValue(result)

    await listAdminUsers({ search: "bob" })

    expect(apiPrivateRequestWithMeta).toHaveBeenCalledWith(
      "/api/v1/admin/users?search=bob",
    )
  })
})

describe("registerAdminUser", () => {
  it("posts /api/v1/admin/users with the correct body via apiPrivateRequest", async () => {
    const input: RegisterAdminUserInput = {
      name: "Alice",
      email: "alice@example.com",
      password: "Secure@1",
      role: "moderator",
    }
    const result = {} as RegisterAdminUserResult
    vi.mocked(apiPrivateRequest).mockResolvedValue(result)

    const data = await registerAdminUser(input)

    expect(apiPrivateRequest).toHaveBeenCalledWith("/api/v1/admin/users", {
      method: "POST",
      body: JSON.stringify(input),
    })
    expect(data).toEqual(result)
  })
})

describe("changeUserRole", () => {
  it("patches /api/v1/admin/users/{id}/role with the correct body via apiPrivateRequest", async () => {
    const result = {} as ChangeUserRoleResult
    vi.mocked(apiPrivateRequest).mockResolvedValue(result)

    const data = await changeUserRole(5, "moderator")

    expect(apiPrivateRequest).toHaveBeenCalledWith("/api/v1/admin/users/5/role", {
      method: "PATCH",
      body: JSON.stringify({ role: "moderator" }),
    })
    expect(data).toEqual(result)
  })
})

describe("resetUserPassword", () => {
  it("patches /api/v1/admin/users/{id}/password with the correct body via apiPrivateRequest", async () => {
    const result = {} as ResetUserPasswordResult
    vi.mocked(apiPrivateRequest).mockResolvedValue(result)

    const data = await resetUserPassword(5, {
      new_password: "NewPass@1",
      new_password_confirmation: "NewPass@1",
    })

    expect(apiPrivateRequest).toHaveBeenCalledWith("/api/v1/admin/users/5/password", {
      method: "PATCH",
      body: JSON.stringify({
        new_password: "NewPass@1",
        new_password_confirmation: "NewPass@1",
      }),
    })
    expect(data).toEqual(result)
  })
})
