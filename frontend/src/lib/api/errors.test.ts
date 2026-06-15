import { describe, it, expect, vi, beforeEach } from "vitest"
import { toast } from "sonner"

import { errorMessageKey, handleApiError } from "./errors"
import { ApiError } from "./client"

vi.mock("sonner", () => ({
  toast: { error: vi.fn() },
}))

const KNOWN_CODES = [
  "VALIDATION_ERROR",
  "UNAUTHORIZED",
  "TOKEN_MISSING",
  "TOKEN_EXPIRED",
  "TOKEN_INVALID",
  "INVALID_CREDENTIALS",
  "ACCOUNT_LOCKED",
  "REFRESH_TOKEN_MISSING",
  "REFRESH_TOKEN_INVALID",
  "REFRESH_TOKEN_REUSE",
  "FORBIDDEN",
  "CANNOT_REVIEW_OWN_EVENT",
  "CANNOT_UNLOCK_SELF",
  "EVENT_NOT_FOUND",
  "DEADLINE_NOT_FOUND",
  "USER_NOT_FOUND",
  "EVENT_NOT_APPROVED",
  "EVENT_ALREADY_SUBMITTED",
  "DEADLINE_ALREADY_INACTIVE",
  "SLUG_ALREADY_EXISTS",
  "USER_NOT_LOCKED",
  "RATE_LIMIT_EXCEEDED",
  "INTERNAL_ERROR",
  "NETWORK_ERROR",
]

describe("errorMessageKey", () => {
  it.each(KNOWN_CODES)("maps %s to errors.%s", (code) => {
    expect(errorMessageKey(code)).toBe(`errors.${code}`)
  })

  it("maps an unknown code to errors.UNKNOWN", () => {
    expect(errorMessageKey("SOME_NEW_CODE_NOT_YET_MAPPED")).toBe("errors.UNKNOWN")
  })
})

describe("handleApiError", () => {
  beforeEach(() => {
    vi.mocked(toast.error).mockClear()
  })

  const t = (key: string): string => `translated:${key}`

  it("shows the translated message for the ApiError's code", () => {
    const error = new ApiError("EVENT_NOT_FOUND", 404, "event not found")

    handleApiError(error, t)

    expect(toast.error).toHaveBeenCalledWith("translated:errors.EVENT_NOT_FOUND")
  })

  it("always shows the generic VALIDATION_ERROR message, never the raw backend message", () => {
    const error = new ApiError("VALIDATION_ERROR", 400, "slug must match ^[A-Za-z0-9_-]+$")

    handleApiError(error, t)

    expect(toast.error).toHaveBeenCalledWith("translated:errors.VALIDATION_ERROR")
    expect(toast.error).not.toHaveBeenCalledWith(expect.stringContaining("slug must match"))
  })

  it("shows the unknown-error message when the error is not an ApiError", () => {
    handleApiError(new Error("something exploded"), t)

    expect(toast.error).toHaveBeenCalledWith("translated:errors.UNKNOWN")
  })
})
