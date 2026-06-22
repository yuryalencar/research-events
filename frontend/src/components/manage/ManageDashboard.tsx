"use client"

import type { JSX } from "react"
import { useLocale, useTranslations } from "next-intl"
import { useRouter } from "next/navigation"
import { ChevronDown } from "lucide-react"

import { logout } from "@/lib/api/auth"
import { useReviewEvents } from "@/hooks/useReviewEvents"
import { ManageHeader } from "./ManageHeader"
import { EventReviewCard } from "./EventReviewCard"
import type { EventStatus, EventTier } from "@/types/api"

// --- Types ---

interface SessionUser {
  id: number
  name: string
  role: "admin" | "moderator"
  email: string
}

interface ManageDashboardProps {
  user: SessionUser
}

// Tier options shown in the filter dropdown.
const TIER_OPTIONS: { label: string; value: EventTier | undefined }[] = [
  { label: "tierAll", value: undefined },
  { label: "A*", value: "A*" },
  { label: "A", value: "A" },
  { label: "B", value: "B" },
  { label: "C", value: "C" },
]

const STATUS_OPTIONS: { label: string; value: EventStatus }[] = [
  { label: "statusPending", value: "pending" },
  { label: "statusApproved", value: "approved" },
  { label: "statusRejected", value: "rejected" },
]

// --- Component ---

function ManageDashboard({ user }: ManageDashboardProps): JSX.Element {
  const t = useTranslations("manage.reviewDashboard")
  const locale = useLocale()
  const router = useRouter()

  const currentYear = new Date().getFullYear()
  const {
    events,
    meta,
    phase,
    draftFilters,
    setDraftStatus,
    setDraftTier,
    setDraftYear,
    apply,
    page,
    goToPage,
  } = useReviewEvents(currentYear)

  async function handleSignOut(): Promise<void> {
    try {
      await logout()
    } finally {
      localStorage.removeItem("manage_user")
      router.replace(`/${locale}/manage`)
    }
  }

  const totalPages = meta ? Math.ceil(meta.total / 30) : 1
  const total = meta?.total ?? 0

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <ManageHeader userName={user.name} userRole={user.role} onSignOut={handleSignOut} />

      <main className="pb-4 pt-9 sm:pt-11">
        <div className="mx-auto flex w-full max-w-4xl flex-col gap-4 px-4 sm:px-6">
        {/* Filter row */}
        <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:gap-4">
          {/* Status */}
          <div className="flex flex-col gap-1">
            <label className="text-xs font-medium text-muted-foreground">{t("statusLabel")}</label>
            <div className="relative">
              <select
                value={draftFilters.status}
                onChange={(e) => setDraftStatus(e.target.value as EventStatus)}
                className="w-full appearance-none rounded-md border border-border bg-background px-3 py-2 pr-9 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              >
                {STATUS_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {t(o.label as Parameters<typeof t>[0])}
                  </option>
                ))}
              </select>
              <ChevronDown size={14} className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
            </div>
          </div>

          {/* Tier */}
          <div className="flex flex-col gap-1">
            <label className="text-xs font-medium text-muted-foreground">{t("tierLabel")}</label>
            <div className="relative">
              <select
                value={draftFilters.tier ?? ""}
                onChange={(e) => setDraftTier((e.target.value as EventTier) || undefined)}
                className="w-full appearance-none rounded-md border border-border bg-background px-3 py-2 pr-9 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              >
                {TIER_OPTIONS.map((o) => (
                  <option key={o.value ?? "all"} value={o.value ?? ""}>
                    {o.label === "tierAll" ? t("tierAll") : o.label}
                  </option>
                ))}
              </select>
              <ChevronDown size={14} className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
            </div>
          </div>

          {/* Year */}
          <div className="flex flex-col gap-1">
            <label className="text-xs font-medium text-muted-foreground">{t("yearLabel")}</label>
            <div className="flex items-center gap-1">
              <input
                type="number"
                value={draftFilters.year ?? ""}
                placeholder={t("allYears")}
                onChange={(e) => {
                  const v = parseInt(e.target.value, 10)
                  setDraftYear(isNaN(v) ? undefined : v)
                }}
                className="w-24 rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              />
              {draftFilters.year !== undefined && (
                <button
                  type="button"
                  onClick={() => setDraftYear(undefined)}
                  className="flex h-8 w-8 items-center justify-center rounded-md border border-border text-sm text-muted-foreground transition hover:bg-muted"
                >
                  ×
                </button>
              )}
            </div>
          </div>

          {/* Apply */}
          <button
            onClick={apply}
            className="rounded-md bg-primary px-5 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
          >
            {t("applyButton")}
          </button>
        </div>

        {/* Event list */}
        {phase === "loading" && (
          <p className="text-sm text-muted-foreground">{t("loading")}</p>
        )}

        {phase === "error" && (
          <p className="text-sm text-red-500">{t("errorLoad")}</p>
        )}

        {phase === "ready" && events.length === 0 && (
          <p className="text-sm text-muted-foreground">{t("empty")}</p>
        )}

        {phase === "ready" && events.length > 0 && (
          <div className="flex flex-col gap-2">
            {events.map((event) => (
              <EventReviewCard
                key={event.id}
                event={event}
                role={user.role}
                userId={user.id}
                locale={locale}
              />
            ))}
          </div>
        )}

        {/* Pagination */}
        {phase === "ready" && (
          <div className="flex flex-col items-center gap-2 sm:flex-row sm:justify-center">
            <button
              onClick={() => goToPage(page - 1)}
              disabled={page <= 1}
              className="rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground transition-colors hover:bg-muted disabled:cursor-not-allowed disabled:opacity-40"
            >
              {t("prevButton")}
            </button>

            <span className="text-sm text-muted-foreground">
              {t("pageLabel", { page, totalPages: totalPages || 1, total })}
            </span>

            <button
              onClick={() => goToPage(page + 1)}
              disabled={page >= totalPages}
              className="rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground transition-colors hover:bg-muted disabled:cursor-not-allowed disabled:opacity-40"
            >
              {t("nextButton")}
            </button>
          </div>
        )}
        </div>
      </main>
    </div>
  )
}

// --- Export ---

export { ManageDashboard }
export type { ManageDashboardProps, SessionUser }
