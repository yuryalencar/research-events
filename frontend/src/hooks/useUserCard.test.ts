import { describe, it, expect, vi } from "vitest"
import { renderHook, act } from "@testing-library/react"

import { useUserCard } from "./useUserCard"
import type { AdminUserListItem } from "@/types/api"

vi.mock("@/lib/api/admin", () => ({
  changeUserRole: vi.fn(),
  resetUserPassword: vi.fn(),
  unlockUser: vi.fn(),
}))

// --- Fixtures ---

const makeUser = (overrides: Partial<AdminUserListItem> = {}): AdminUserListItem => ({
  id: 1,
  name: "Alice",
  email: "alice@example.com",
  role: "moderator",
  created_at: "2026-01-01T00:00:00Z",
  locked_at: null,
  deleted_at: null,
  ...overrides,
})

const noop = () => {}

// --- useUserCard ---

describe("useUserCard", () => {
  describe("expand toggle", () => {
    // Spec: collapsed card by default, chevron toggle
    it("initialises with isExpanded false", () => {
      const { result } = renderHook(() => useUserCard(makeUser(), noop, noop))
      expect(result.current.isExpanded).toBe(false)
    })

    it("toggleExpanded flips isExpanded on each call", () => {
      const { result } = renderHook(() => useUserCard(makeUser(), noop, noop))

      act(() => { result.current.toggleExpanded() })
      expect(result.current.isExpanded).toBe(true)

      act(() => { result.current.toggleExpanded() })
      expect(result.current.isExpanded).toBe(false)
    })
  })

  describe("role section", () => {
    // Spec: "Dropdown pre-selected with current role"
    it("selectedRole initialises to the user's current role", () => {
      const { result } = renderHook(() => useUserCard(makeUser({ role: "moderator" }), noop, noop))
      expect(result.current.selectedRole).toBe("moderator")
    })

    // Spec: "Apply Role button disabled when dropdown value equals the user's current role"
    it("canApplyRole is false when selectedRole equals the current role", () => {
      const { result } = renderHook(() => useUserCard(makeUser({ role: "moderator" }), noop, noop))
      expect(result.current.canApplyRole).toBe(false)
    })

    it("canApplyRole is true when selectedRole differs from the current role", () => {
      const { result } = renderHook(() => useUserCard(makeUser({ role: "moderator" }), noop, noop))

      act(() => { result.current.setSelectedRole("contributor") })

      expect(result.current.canApplyRole).toBe(true)
    })
  })

  describe("password section", () => {
    // Spec: "Apply Password disabled until: new password non-empty AND meets complexity AND confirm matches"
    it("canApplyPassword is false when both password fields are empty", () => {
      const { result } = renderHook(() => useUserCard(makeUser(), noop, noop))
      expect(result.current.canApplyPassword).toBe(false)
    })

    it("canApplyPassword is false when password is set but complexity not fully met", () => {
      const { result } = renderHook(() => useUserCard(makeUser(), noop, noop))

      act(() => {
        result.current.setNewPassword("short")
        result.current.setConfirmPassword("short")
      })

      expect(result.current.canApplyPassword).toBe(false)
    })

    it("canApplyPassword is false when confirm does not match new password", () => {
      const { result } = renderHook(() => useUserCard(makeUser(), noop, noop))

      act(() => {
        result.current.setNewPassword("Secure@1")
        result.current.setConfirmPassword("Different@2")
      })

      expect(result.current.canApplyPassword).toBe(false)
    })

    it("canApplyPassword is true when complexity is met and passwords match", () => {
      const { result } = renderHook(() => useUserCard(makeUser(), noop, noop))

      act(() => {
        result.current.setNewPassword("Secure@1")
        result.current.setConfirmPassword("Secure@1")
      })

      expect(result.current.canApplyPassword).toBe(true)
    })

    // Spec: "If confirm is filled but doesn't match → inline error below confirm field"
    it("passwordMismatch is true when confirm is filled but differs from new password", () => {
      const { result } = renderHook(() => useUserCard(makeUser(), noop, noop))

      act(() => {
        result.current.setNewPassword("Secure@1")
        result.current.setConfirmPassword("Wrong@99")
      })

      expect(result.current.passwordMismatch).toBe(true)
    })

    it("passwordMismatch is false when confirm is empty", () => {
      const { result } = renderHook(() => useUserCard(makeUser(), noop, noop))

      act(() => { result.current.setNewPassword("Secure@1") })

      expect(result.current.passwordMismatch).toBe(false)
    })
  })
})
