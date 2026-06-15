import { describe, it, expect, vi } from "vitest"

import { apiPrivateRequest } from "./client"
import { reviewEvent, unlockUser } from "./admin"
import type { ReviewEventInput, ReviewEventResult, UnlockUserResult } from "@/types/api"

vi.mock("./client", () => ({
  apiPrivateRequest: vi.fn(),
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
