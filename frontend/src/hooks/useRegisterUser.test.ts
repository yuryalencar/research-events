import { describe, it, expect, vi } from "vitest"
import { renderHook, act } from "@testing-library/react"

import { useRegisterUser } from "./useRegisterUser"

vi.mock("@/lib/api/admin", () => ({
  registerAdminUser: vi.fn(),
}))

const noop = () => {}

// --- useRegisterUser ---

describe("useRegisterUser", () => {
  describe("initial state", () => {
    // Spec: form starts blank with Register User button disabled
    it("initialises with all fields empty and canSubmit false", () => {
      const { result } = renderHook(() => useRegisterUser(noop))

      expect(result.current.name).toBe("")
      expect(result.current.email).toBe("")
      expect(result.current.role).toBe("")
      expect(result.current.password).toBe("")
      expect(result.current.confirmPassword).toBe("")
      expect(result.current.canSubmit).toBe(false)
      expect(result.current.phase).toBe("idle")
      expect(result.current.error).toBeNull()
      expect(result.current.registeredUser).toBeNull()
    })
  })

  describe("canSubmit", () => {
    // Spec: "Register User button is disabled until all fields are valid and passwords match"
    it("is false when any required field is still empty", () => {
      const { result } = renderHook(() => useRegisterUser(noop))

      // Fill everything except role
      act(() => {
        result.current.setName("Alice")
        result.current.setEmail("alice@example.com")
        result.current.setPassword("Secure@1")
        result.current.setConfirmPassword("Secure@1")
        // role intentionally left empty
      })

      expect(result.current.canSubmit).toBe(false)
    })

    // Spec: complexity checklist — all four rules must be met
    it("is false when the password does not meet complexity requirements", () => {
      const { result } = renderHook(() => useRegisterUser(noop))

      act(() => {
        result.current.setName("Alice")
        result.current.setEmail("alice@example.com")
        result.current.setRole("moderator")
        result.current.setPassword("weakpass") // no uppercase, no special char
        result.current.setConfirmPassword("weakpass")
      })

      expect(result.current.canSubmit).toBe(false)
    })

    // Spec: "Confirm password must match password"
    it("is false when confirm password does not match password", () => {
      const { result } = renderHook(() => useRegisterUser(noop))

      act(() => {
        result.current.setName("Alice")
        result.current.setEmail("alice@example.com")
        result.current.setRole("moderator")
        result.current.setPassword("Secure@1")
        result.current.setConfirmPassword("Different@2")
      })

      expect(result.current.canSubmit).toBe(false)
    })

    // Happy path — all gates pass
    it("is true when all fields are valid and passwords match", () => {
      const { result } = renderHook(() => useRegisterUser(noop))

      act(() => {
        result.current.setName("Alice")
        result.current.setEmail("alice@example.com")
        result.current.setRole("moderator")
        result.current.setPassword("Secure@1")
        result.current.setConfirmPassword("Secure@1")
      })

      expect(result.current.canSubmit).toBe(true)
    })
  })

  describe("passwordMismatch", () => {
    // Spec: mismatch error only shows once confirm field has been touched
    it("is false when confirm password is still empty", () => {
      const { result } = renderHook(() => useRegisterUser(noop))

      act(() => { result.current.setPassword("Secure@1") })

      expect(result.current.passwordMismatch).toBe(false)
    })
  })

  describe("reset", () => {
    // Spec: "'Register another user' resets all form state (fields, errors, complexity)"
    it("clears all form fields and resets phase to idle", () => {
      const { result } = renderHook(() => useRegisterUser(noop))

      act(() => {
        result.current.setName("Alice")
        result.current.setEmail("alice@example.com")
        result.current.setRole("moderator")
        result.current.setPassword("Secure@1")
        result.current.setConfirmPassword("Secure@1")
      })

      act(() => { result.current.reset() })

      expect(result.current.name).toBe("")
      expect(result.current.email).toBe("")
      expect(result.current.role).toBe("")
      expect(result.current.password).toBe("")
      expect(result.current.confirmPassword).toBe("")
      expect(result.current.canSubmit).toBe(false)
      expect(result.current.phase).toBe("idle")
      expect(result.current.error).toBeNull()
    })
  })
})
