import { describe, it, expect, vi } from "vitest"

import { apiPrivateRequest } from "./client"
import { updatePassword } from "./users"
import type { UpdatePasswordInput } from "./users"

vi.mock("./client", () => ({
  apiPrivateRequest: vi.fn(),
}))

describe("updatePassword", () => {
  it("patches /api/v1/users/me/password via apiPrivateRequest with the three fields", async () => {
    vi.mocked(apiPrivateRequest).mockResolvedValue(undefined)

    const input: UpdatePasswordInput = {
      current_password: "OldPass@1",
      new_password: "NewPass@2",
      new_password_confirmation: "NewPass@2",
    }

    await updatePassword(input)

    expect(apiPrivateRequest).toHaveBeenCalledWith("/api/v1/users/me/password", {
      method: "PATCH",
      body: JSON.stringify(input),
    })
  })
})
