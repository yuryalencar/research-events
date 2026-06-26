import { useState, useRef, useEffect, useCallback } from "react"

import { listAdminUsers } from "@/lib/api/admin"
import type { AdminUserListItem, ApiMeta, ListAdminUsersParams } from "@/types/api"

// --- Types ---

interface AdminUserDraftFilters {
  search: string
  roles: string[]
  locked: boolean
  deleted: boolean
}

type AdminUsersPhase = "loading" | "ready" | "error"

interface FetchParams extends AdminUserDraftFilters {
  page: number
}

interface UseAdminUsersReturn {
  users: AdminUserListItem[]
  meta: ApiMeta | null
  phase: AdminUsersPhase
  draftFilters: AdminUserDraftFilters
  setSearch: (v: string) => void
  toggleRole: (role: string) => void
  toggleLocked: () => void
  toggleDeleted: () => void
  apply: () => void
  reset: () => void
  page: number
  goToPage: (page: number) => void
}

// --- Defaults ---

const defaultFilters: AdminUserDraftFilters = {
  search: "",
  roles: [],
  locked: false,
  deleted: false,
}

// --- Hook ---

// useAdminUsers manages filter, pagination, and fetch state for the admin user list.
// Draft filters are Apply-gated: changing a filter never triggers a fetch on its own.
// apply() commits the draft and fetches page 1. goToPage() always uses the last applied
// filters, so in-progress draft edits can't interfere with pagination.
function useAdminUsers(): UseAdminUsersReturn {
  // 1. State
  const [draftFilters, setDraftFilters] = useState<AdminUserDraftFilters>({ ...defaultFilters })
  const [fetchParams, setFetchParams] = useState<FetchParams>({ ...defaultFilters, page: 1 })
  const [users, setUsers] = useState<AdminUserListItem[]>([])
  const [meta, setMeta] = useState<ApiMeta | null>(null)
  const [phase, setPhase] = useState<AdminUsersPhase>("loading")

  // Refs let stable useCallback closures read the latest filter values without
  // being added to dependency arrays (which would recreate callbacks on every render).
  const draftFiltersRef = useRef<AdminUserDraftFilters>({ ...defaultFilters })
  const appliedFiltersRef = useRef<AdminUserDraftFilters>({ ...defaultFilters })

  // 2. Fetch effect — fires on mount and whenever fetchParams changes.
  useEffect(() => {
    let isMounted = true
    setPhase("loading")

    const params: ListAdminUsersParams = {
      page: fetchParams.page,
      page_size: 20,
      pagination: "on",
    }
    if (fetchParams.search) params.search = fetchParams.search
    if (fetchParams.roles.length > 0) params.roles = fetchParams.roles
    if (fetchParams.locked) params.locked = true
    if (fetchParams.deleted) params.include_deleted = true

    listAdminUsers(params)
      .then((result) => {
        if (!isMounted) return
        setUsers(result.data)
        setMeta(result.meta)
        setPhase("ready")
      })
      .catch(() => {
        if (!isMounted) return
        setUsers([])
        setMeta(null)
        setPhase("error")
      })

    return () => {
      isMounted = false
    }
  }, [fetchParams])

  // 3. Handlers — all stable via useCallback with no deps; read latest through refs.

  const setSearch = useCallback((v: string) => {
    const next = { ...draftFiltersRef.current, search: v }
    draftFiltersRef.current = next
    setDraftFilters(next)
  }, [])

  const toggleRole = useCallback((role: string) => {
    const current = draftFiltersRef.current
    const roles = current.roles.includes(role)
      ? current.roles.filter((r) => r !== role)
      : [...current.roles, role]
    const next = { ...current, roles }
    draftFiltersRef.current = next
    setDraftFilters(next)
  }, [])

  const toggleLocked = useCallback(() => {
    const next = { ...draftFiltersRef.current, locked: !draftFiltersRef.current.locked }
    draftFiltersRef.current = next
    setDraftFilters(next)
  }, [])

  const toggleDeleted = useCallback(() => {
    const next = { ...draftFiltersRef.current, deleted: !draftFiltersRef.current.deleted }
    draftFiltersRef.current = next
    setDraftFilters(next)
  }, [])

  // apply commits the current draft to the applied set and re-fetches page 1.
  const apply = useCallback(() => {
    const filters = draftFiltersRef.current
    appliedFiltersRef.current = filters
    setFetchParams({ ...filters, page: 1 })
  }, [])

  // reset clears draft fields to defaults without triggering a fetch — the user must
  // click Apply to actually refetch with the cleared filters.
  const reset = useCallback(() => {
    const next = { ...defaultFilters }
    draftFiltersRef.current = next
    setDraftFilters(next)
  }, [])

  // goToPage fetches a different page using the already-applied filters, not the
  // in-progress draft — so draft edits can't change what Prev/Next navigates through.
  const goToPage = useCallback((page: number) => {
    setFetchParams({ ...appliedFiltersRef.current, page })
  }, [])

  // 4. Return
  return {
    users,
    meta,
    phase,
    draftFilters,
    setSearch,
    toggleRole,
    toggleLocked,
    toggleDeleted,
    apply,
    reset,
    page: fetchParams.page,
    goToPage,
  }
}

// --- Export ---

export { useAdminUsers }
export type { UseAdminUsersReturn, AdminUserDraftFilters, AdminUsersPhase }
