"use client"

import type { JSX } from "react"
import Link from "next/link"
import { useTranslations } from "next-intl"

import { useAdminUsers } from "@/hooks/useAdminUsers"
import { UserFilters } from "./UserFilters"
import { UserTableCard } from "./UserTableCard"

// --- Types ---

interface UserListPageProps {
  locale: string
}

// --- Sub-components ---

interface PaginationProps {
  page: number
  totalPages: number
  total: number
  onPrev: () => void
  onNext: () => void
  t: (key: string, values?: Record<string, string | number>) => string
}

function Pagination({ page, totalPages, total, onPrev, onNext, t }: PaginationProps): JSX.Element {
  return (
    <div className="flex items-center justify-between pt-2">
      <p className="text-xs text-muted-foreground">
        {t("list.pageLabel", { page, totalPages, total })}
      </p>
      <div className="flex gap-2">
        <button
          type="button"
          onClick={onPrev}
          disabled={page <= 1}
          className="rounded-md border border-border px-3 py-1.5 text-sm text-foreground transition-colors hover:bg-muted disabled:opacity-50"
        >
          {t("list.prevButton")}
        </button>
        <button
          type="button"
          onClick={onNext}
          disabled={page >= totalPages}
          className="rounded-md border border-border px-3 py-1.5 text-sm text-foreground transition-colors hover:bg-muted disabled:opacity-50"
        >
          {t("list.nextButton")}
        </button>
      </div>
    </div>
  )
}

// --- Component ---

function UserListPage({ locale }: UserListPageProps): JSX.Element {
  const t = useTranslations("manage.users")

  const {
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
    page,
    goToPage,
  } = useAdminUsers()

  const totalPages = meta ? Math.max(1, Math.ceil(meta.total / 20)) : 1

  // Optimistic role/unlock refresh — the list re-fetches are intentionally not triggered
  // on single-card actions; the card updates its own state optimistically. A full re-fetch
  // only happens via Apply or pagination, which is acceptable for an admin tool.
  const noop = () => undefined

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-6">
      {/* Header row */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-col gap-1">
          <Link
            href={`/${locale}/manage`}
            className="text-xs text-muted-foreground hover:text-foreground"
          >
            {t("backToEvents")}
          </Link>
          <h1 className="text-xl font-semibold text-foreground">{t("pageTitle")}</h1>
        </div>
        <Link
          href={`/${locale}/manage/admin/users/register`}
          className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
        >
          {t("registerButton")}
        </Link>
      </div>

      {/* Filters */}
      <UserFilters
        draftFilters={draftFilters}
        setSearch={setSearch}
        toggleRole={toggleRole}
        toggleLocked={toggleLocked}
        toggleDeleted={toggleDeleted}
        apply={apply}
        reset={reset}
      />

      {/* List body */}
      <div className="flex flex-col gap-3">
        {phase === "loading" && (
          <p className="text-sm text-muted-foreground">{t("list.loading")}</p>
        )}

        {phase === "error" && (
          <p className="text-sm text-red-500">{t("list.error")}</p>
        )}

        {phase === "ready" && users.length === 0 && (
          <p className="text-sm text-muted-foreground">{t("list.empty")}</p>
        )}

        {phase === "ready" && users.map((user) => (
          <UserTableCard
            key={user.id}
            user={user}
            onRoleChanged={noop}
            onUnlocked={noop}
          />
        ))}
      </div>

      {/* Pagination — only shown when there is data */}
      {phase === "ready" && meta && meta.total > 0 && (
        <Pagination
          page={page}
          totalPages={totalPages}
          total={meta.total}
          onPrev={() => goToPage(page - 1)}
          onNext={() => goToPage(page + 1)}
          t={t as (key: string, values?: Record<string, string | number>) => string}
        />
      )}
    </div>
  )
}

// --- Export ---

export { UserListPage }
export type { UserListPageProps }
