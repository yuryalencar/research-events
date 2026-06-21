import { describe, it, expect, vi, beforeEach } from "vitest"
import { renderHook, act } from "@testing-library/react"

import { useUpdatePassword } from "./useUpdatePassword"
import * as usersApi from "@/lib/api/users"
import * as errorsApi from "@/lib/api/errors"
import { ApiError } from "@/lib/api/client"

vi.mock("@/lib/api/users", () => ({
  updatePassword: vi.fn(),
}))

vi.mock("@/lib/api/errors", () => ({
  handleApiError: vi.fn(),
}))

// Minimal translation stub — returns the key so tests can assert on it.
const t = (key: string): string => key

describe("useUpdatePassword — complexity", () => {
  it("all rules are false when newPassword is empty", () => {
    const { result } = renderHook(() => useUpdatePassword(t))

    expect(result.current.complexity).toEqual({
      minLength: false,
      hasUppercase: false,
      hasLowercase: false,
      hasSpecial: false,
    })
  })

  it("minLength is true at exactly 8 characters", () => {
    const { result } = renderHook(() => useUpdatePassword(t))

    act(() => result.current.setNewPassword("Abcde@1x"))

    expect(result.current.complexity.minLength).toBe(true)
  })

  it("hasUppercase is true when at least one A–Z character is present", () => {
    const { result } = renderHook(() => useUpdatePassword(t))

    act(() => result.current.setNewPassword("abcdefgH"))

    expect(result.current.complexity.hasUppercase).toBe(true)
  })

  it("hasLowercase is true when at least one a–z character is present", () => {
    const { result } = renderHook(() => useUpdatePassword(t))

    act(() => result.current.setNewPassword("ABCDEFGh"))

    expect(result.current.complexity.hasLowercase).toBe(true)
  })

  it("hasSpecial is true when at least one non-alphanumeric character is present", () => {
    const { result } = renderHook(() => useUpdatePassword(t))

    act(() => result.current.setNewPassword("abcdefg@"))

    expect(result.current.complexity.hasSpecial).toBe(true)
  })

  it("all rules are true for a fully valid password", () => {
    const { result } = renderHook(() => useUpdatePassword(t))

    act(() => result.current.setNewPassword("NewPass@1"))

    expect(result.current.complexity).toEqual({
      minLength: true,
      hasUppercase: true,
      hasLowercase: true,
      hasSpecial: true,
    })
  })
})

describe("useUpdatePassword — confirmMismatch", () => {
  it("is false when confirmPassword is empty", () => {
    const { result } = renderHook(() => useUpdatePassword(t))

    act(() => result.current.setNewPassword("NewPass@1"))

    expect(result.current.confirmMismatch).toBe(false)
  })

  it("is false when confirmPassword matches newPassword", () => {
    const { result } = renderHook(() => useUpdatePassword(t))

    act(() => {
      result.current.setNewPassword("NewPass@1")
      result.current.setConfirmPassword("NewPass@1")
    })

    expect(result.current.confirmMismatch).toBe(false)
  })

  it("is true when confirmPassword differs from newPassword", () => {
    const { result } = renderHook(() => useUpdatePassword(t))

    act(() => {
      result.current.setNewPassword("NewPass@1")
      result.current.setConfirmPassword("Different@2")
    })

    expect(result.current.confirmMismatch).toBe(true)
  })
})

describe("useUpdatePassword — canSubmit", () => {
  it("is false when currentPassword is empty", () => {
    const { result } = renderHook(() => useUpdatePassword(t))

    act(() => {
      result.current.setNewPassword("NewPass@1")
      result.current.setConfirmPassword("NewPass@1")
    })

    expect(result.current.canSubmit).toBe(false)
  })

  it("is false when any complexity rule fails", () => {
    const { result } = renderHook(() => useUpdatePassword(t))

    act(() => {
      result.current.setCurrentPassword("OldPass@1")
      result.current.setNewPassword("short") // fails minLength + uppercase + special
      result.current.setConfirmPassword("short")
    })

    expect(result.current.canSubmit).toBe(false)
  })

  it("is false when confirmPassword does not match newPassword", () => {
    const { result } = renderHook(() => useUpdatePassword(t))

    act(() => {
      result.current.setCurrentPassword("OldPass@1")
      result.current.setNewPassword("NewPass@1")
      result.current.setConfirmPassword("Different@2")
    })

    expect(result.current.canSubmit).toBe(false)
  })

  it("is true when all rules pass, confirm matches, and currentPassword is non-empty", () => {
    const { result } = renderHook(() => useUpdatePassword(t))

    act(() => {
      result.current.setCurrentPassword("OldPass@1")
      result.current.setNewPassword("NewPass@1")
      result.current.setConfirmPassword("NewPass@1")
    })

    expect(result.current.canSubmit).toBe(true)
  })
})

describe("useUpdatePassword — submit", () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it("calls updatePassword with correct fields and sets phase to success on success", async () => {
    vi.mocked(usersApi.updatePassword).mockResolvedValue(undefined)

    const { result } = renderHook(() => useUpdatePassword(t))

    act(() => {
      result.current.setCurrentPassword("OldPass@1")
      result.current.setNewPassword("NewPass@1")
      result.current.setConfirmPassword("NewPass@1")
    })

    await act(() => result.current.submit())

    expect(usersApi.updatePassword).toHaveBeenCalledWith({
      current_password: "OldPass@1",
      new_password: "NewPass@1",
      new_password_confirmation: "NewPass@1",
    })
    expect(result.current.phase).toBe("success")
  })

  it("calls handleApiError and keeps phase idle on API error", async () => {
    const error = new ApiError("INVALID_CURRENT_PASSWORD", 400, "current password is incorrect")
    vi.mocked(usersApi.updatePassword).mockRejectedValue(error)

    const { result } = renderHook(() => useUpdatePassword(t))

    act(() => {
      result.current.setCurrentPassword("WrongPass@1")
      result.current.setNewPassword("NewPass@1")
      result.current.setConfirmPassword("NewPass@1")
    })

    await act(() => result.current.submit())

    expect(errorsApi.handleApiError).toHaveBeenCalledWith(error, t)
    expect(result.current.phase).toBe("idle")
  })

  it("does not call updatePassword when canSubmit is false", async () => {
    const { result } = renderHook(() => useUpdatePassword(t))

    // currentPassword is empty — canSubmit is false
    await act(() => result.current.submit())

    expect(usersApi.updatePassword).not.toHaveBeenCalled()
  })

  it("calls onAuthError when an auth error is thrown (TOKEN_MISSING)", async () => {
    const error = new ApiError("TOKEN_MISSING", 401, "no session")
    vi.mocked(usersApi.updatePassword).mockRejectedValue(error)
    const onAuthError = vi.fn()

    const { result } = renderHook(() => useUpdatePassword(t, onAuthError))

    act(() => {
      result.current.setCurrentPassword("OldPass@1")
      result.current.setNewPassword("NewPass@1")
      result.current.setConfirmPassword("NewPass@1")
    })

    await act(() => result.current.submit())

    expect(errorsApi.handleApiError).toHaveBeenCalledWith(error, t)
    expect(onAuthError).toHaveBeenCalledOnce()
    expect(result.current.phase).toBe("idle")
  })

  it("calls onAuthError when an auth error is thrown (TOKEN_EXPIRED)", async () => {
    const error = new ApiError("TOKEN_EXPIRED", 401, "token expired")
    vi.mocked(usersApi.updatePassword).mockRejectedValue(error)
    const onAuthError = vi.fn()

    const { result } = renderHook(() => useUpdatePassword(t, onAuthError))

    act(() => {
      result.current.setCurrentPassword("OldPass@1")
      result.current.setNewPassword("NewPass@1")
      result.current.setConfirmPassword("NewPass@1")
    })

    await act(() => result.current.submit())

    expect(onAuthError).toHaveBeenCalledOnce()
  })

  it("does not call onAuthError for non-auth errors", async () => {
    const error = new ApiError("INVALID_CURRENT_PASSWORD", 422, "wrong password")
    vi.mocked(usersApi.updatePassword).mockRejectedValue(error)
    const onAuthError = vi.fn()

    const { result } = renderHook(() => useUpdatePassword(t, onAuthError))

    act(() => {
      result.current.setCurrentPassword("WrongPass@1")
      result.current.setNewPassword("NewPass@1")
      result.current.setConfirmPassword("NewPass@1")
    })

    await act(() => result.current.submit())

    expect(onAuthError).not.toHaveBeenCalled()
  })
})
