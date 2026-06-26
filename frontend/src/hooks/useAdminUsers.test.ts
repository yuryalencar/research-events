import { describe, it, expect, vi, beforeEach } from "vitest"
import { renderHook, waitFor, act } from "@testing-library/react"

import { useAdminUsers } from "./useAdminUsers"
import { listAdminUsers } from "@/lib/api/admin"
import type { AdminUserListItem, ApiMeta } from "@/types/api"

vi.mock("@/lib/api/admin", () => ({
  listAdminUsers: vi.fn(),
}))

// --- Fixtures ---

const meta: ApiMeta = { page: 1, total: 10 }

const makeUser = (overrides: Partial<AdminUserListItem> = {}): AdminUserListItem => ({
  id: 1,
  name: "Alice",
  email: "alice@example.com",
  role: "admin",
  created_at: "2026-01-01T00:00:00Z",
  locked_at: null,
  deleted_at: null,
  ...overrides,
})

// --- useAdminUsers ---

describe("useAdminUsers", () => {
  beforeEach(() => {
    vi.mocked(listAdminUsers).mockReset()
    vi.mocked(listAdminUsers).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })
  })

  describe("initial state", () => {
    // Spec: filter area defaults — search empty, no roles selected, locked off, deleted off
    it("initialises with default draft filter state before any interaction", async () => {
      const { result } = renderHook(() => useAdminUsers())

      expect(result.current.draftFilters).toEqual({
        search: "",
        roles: [],
        locked: false,
        deleted: false,
      })
    })
  })

  describe("draft filter changes", () => {
    // Spec: "Filters are Apply-gated — no fetch is triggered until the Apply button is clicked"
    it("setSearch updates draft search without triggering an additional fetch", async () => {
      const { result } = renderHook(() => useAdminUsers())
      await waitFor(() => expect(result.current.phase).toBe("ready"))
      vi.mocked(listAdminUsers).mockClear()

      act(() => {
        result.current.setSearch("alice")
      })

      expect(result.current.draftFilters.search).toBe("alice")
      expect(listAdminUsers).not.toHaveBeenCalled()
    })

    // Spec: "Role chips: each chip independently toggleable; multiple can be active simultaneously"
    it("toggleRole adds a role to draft; calling again with the same role removes it", async () => {
      const { result } = renderHook(() => useAdminUsers())
      await waitFor(() => expect(result.current.phase).toBe("ready"))

      act(() => { result.current.toggleRole("admin") })
      expect(result.current.draftFilters.roles).toEqual(["admin"])

      act(() => { result.current.toggleRole("moderator") })
      expect(result.current.draftFilters.roles).toEqual(["admin", "moderator"])

      act(() => { result.current.toggleRole("admin") })
      expect(result.current.draftFilters.roles).toEqual(["moderator"])
    })

    // Spec: "Locked / Deleted badges: each independently toggleable"
    it("toggleLocked and toggleDeleted flip their respective draft booleans independently", async () => {
      const { result } = renderHook(() => useAdminUsers())
      await waitFor(() => expect(result.current.phase).toBe("ready"))

      act(() => { result.current.toggleLocked() })
      expect(result.current.draftFilters.locked).toBe(true)
      expect(result.current.draftFilters.deleted).toBe(false)

      act(() => { result.current.toggleDeleted() })
      expect(result.current.draftFilters.deleted).toBe(true)

      act(() => { result.current.toggleLocked() })
      expect(result.current.draftFilters.locked).toBe(false)
    })
  })

  describe("apply", () => {
    // Spec: "Apply resets to page 1"
    it("fetches page 1 with current draft filters when apply is called", async () => {
      const user = makeUser()
      vi.mocked(listAdminUsers).mockResolvedValue({ data: [user], meta })

      const { result } = renderHook(() => useAdminUsers())
      await waitFor(() => expect(result.current.phase).toBe("ready"))
      vi.mocked(listAdminUsers).mockClear()

      act(() => {
        result.current.setSearch("alice")
        result.current.toggleRole("admin")
        result.current.toggleLocked()
      })

      vi.mocked(listAdminUsers).mockResolvedValue({ data: [user], meta: { page: 1, total: 1 } })

      act(() => { result.current.apply() })

      await waitFor(() => expect(result.current.phase).toBe("ready"))

      expect(listAdminUsers).toHaveBeenCalledWith(
        expect.objectContaining({
          search: "alice",
          roles: ["admin"],
          locked: true,
          page: 1,
        }),
      )
      expect(result.current.page).toBe(1)
      expect(result.current.users).toEqual([user])
    })
  })

  describe("reset", () => {
    // Spec: "Reset clears all fields back to defaults (no fetch until Apply)"
    it("resets draft filters to defaults without triggering a fetch", async () => {
      const { result } = renderHook(() => useAdminUsers())
      await waitFor(() => expect(result.current.phase).toBe("ready"))

      act(() => {
        result.current.setSearch("bob")
        result.current.toggleRole("moderator")
        result.current.toggleLocked()
        result.current.toggleDeleted()
      })

      vi.mocked(listAdminUsers).mockClear()

      act(() => { result.current.reset() })

      expect(result.current.draftFilters).toEqual({
        search: "",
        roles: [],
        locked: false,
        deleted: false,
      })
      expect(listAdminUsers).not.toHaveBeenCalled()
    })
  })

  describe("goToPage", () => {
    // Spec: "Prev/Next always use the last submitted filter set, not in-progress edits"
    it("fetches the given page with applied filters, ignoring unapplied draft changes", async () => {
      vi.mocked(listAdminUsers).mockResolvedValue({ data: [], meta: { page: 1, total: 60 } })

      const { result } = renderHook(() => useAdminUsers())
      await waitFor(() => expect(result.current.phase).toBe("ready"))

      // Apply initial filters so applied state is known
      act(() => {
        result.current.setSearch("alice")
        result.current.apply()
      })
      await waitFor(() => expect(result.current.phase).toBe("ready"))

      // Change draft but do NOT apply — goToPage must use applied, not draft
      act(() => { result.current.setSearch("bob") })
      vi.mocked(listAdminUsers).mockClear()
      vi.mocked(listAdminUsers).mockResolvedValue({ data: [], meta: { page: 2, total: 60 } })

      act(() => { result.current.goToPage(2) })

      await waitFor(() => expect(result.current.phase).toBe("ready"))

      expect(listAdminUsers).toHaveBeenCalledWith(
        expect.objectContaining({ search: "alice", page: 2 }), // applied "alice", not draft "bob"
      )
      expect(result.current.page).toBe(2)
    })
  })
})
