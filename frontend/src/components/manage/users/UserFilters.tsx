"use client"

import type { JSX } from "react"
import { useTranslations } from "next-intl"

import type { AdminUserDraftFilters } from "@/hooks/useAdminUsers"

// --- Types ---

interface UserFiltersProps {
  draftFilters: AdminUserDraftFilters
  setSearch: (v: string) => void
  toggleRole: (role: string) => void
  toggleLocked: () => void
  toggleDeleted: () => void
  apply: () => void
  reset: () => void
}

// --- Sub-components ---

interface ChipProps {
  label: string
  active: boolean
  onClick: () => void
}

function Chip({ label, active, onClick }: ChipProps): JSX.Element {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`rounded-full border px-3 py-1 text-xs font-medium transition-colors ${
        active
          ? "border-primary bg-primary text-primary-foreground"
          : "border-border bg-background text-foreground hover:bg-muted"
      }`}
    >
      {label}
    </button>
  )
}

// --- Component ---

const ROLE_OPTIONS = ["admin", "moderator", "contributor"] as const

function UserFilters({
  draftFilters,
  setSearch,
  toggleRole,
  toggleLocked,
  toggleDeleted,
  apply,
  reset,
}: UserFiltersProps): JSX.Element {
  const t = useTranslations("manage.users")

  return (
    <div className="flex flex-col gap-3">
      {/* Search */}
      <input
        type="text"
        value={draftFilters.search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder={t("filters.searchPlaceholder")}
        className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
      />

      {/* Role chips row */}
      <div className="flex flex-wrap items-center gap-2">
        <span className="text-xs font-medium text-muted-foreground">
          {t("filters.rolesLabel")}:
        </span>
        {ROLE_OPTIONS.map((role) => (
          <Chip
            key={role}
            label={role}
            active={draftFilters.roles.includes(role)}
            onClick={() => toggleRole(role)}
          />
        ))}
      </div>

      {/* Status chips row + actions */}
      <div className="flex flex-wrap items-center gap-2">
        <span className="text-xs font-medium text-muted-foreground">
          {t("filters.statusLabel")}:
        </span>
        <Chip
          label={t("filters.lockedBadge")}
          active={draftFilters.locked}
          onClick={toggleLocked}
        />
        <Chip
          label={t("filters.deletedBadge")}
          active={draftFilters.deleted}
          onClick={toggleDeleted}
        />

        <div className="ml-auto flex gap-2">
          <button
            type="button"
            onClick={reset}
            className="rounded-md border border-border px-3 py-1.5 text-sm text-foreground transition-colors hover:bg-muted"
          >
            {t("filters.reset")}
          </button>
          <button
            type="button"
            onClick={apply}
            className="rounded-md bg-primary px-4 py-1.5 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
          >
            {t("filters.apply")}
          </button>
        </div>
      </div>
    </div>
  )
}

// --- Export ---

export { UserFilters }
export type { UserFiltersProps }
