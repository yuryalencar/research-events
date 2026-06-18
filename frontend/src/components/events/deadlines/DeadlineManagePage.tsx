"use client"

import type { JSX } from "react"
import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import { useTranslations, useLocale } from "next-intl"

import { useDeadlineManage } from "@/hooks/useDeadlineManage"
import { DeadlineCard } from "@/components/events/deadlines/DeadlineCard"
import { AddDeadlineCard } from "@/components/events/deadlines/AddDeadlineCard"
import { DeadlineManageSuccess } from "@/components/events/deadlines/DeadlineManageSuccess"
import type { EventListItem } from "@/types/api"

// --- Constants ---

const SESSION_KEY = "deadline_management_event"

const inputClass =
  "rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"

// --- Sub-component: content (receives event, calls hook) ---

function DeadlineManageContent({ event }: { event: EventListItem }): JSX.Element {
  const t = useTranslations("deadlines.manage")
  const locale = useLocale()
  const router = useRouter()

  const {
    deadlineStates,
    newDeadlines,
    contributor,
    errors,
    apiError,
    pageState,
    hasChanges,
    successSummary,
    startSupersede,
    revertSupersede,
    cancelDeadlineLocal,
    revertCancel,
    updateSupersede,
    addNewDeadline,
    removeNewDeadline,
    updateNewDeadline,
    updateContributor,
    submitChanges,
  } = useDeadlineManage(event)

  if (pageState === "success" && successSummary) {
    return <DeadlineManageSuccess event={event} summary={successSummary} />
  }

  const isSubmitting = pageState === "submitting"
  const activeDeadlines = event.deadlines.filter((d) => d.is_active)

  return (
    <main className="mx-auto max-w-3xl px-4 py-8">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold">{t("pageTitle")}</h1>
        <p className="mt-1 text-muted-foreground">{event.name}</p>
      </div>

      {/* Existing deadlines */}
      <section className="mb-8">
        <h2 className="mb-3 text-lg font-semibold">{t("existingDeadlinesTitle")}</h2>
        {activeDeadlines.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("noDeadlines")}</p>
        ) : (
          <div className="flex flex-col gap-3">
            {activeDeadlines.map((d) => (
              <DeadlineCard
                key={d.id}
                deadline={d}
                state={deadlineStates[d.id]}
                errors={errors}
                onStartSupersede={() => startSupersede(d.id)}
                onRevertSupersede={() => revertSupersede(d.id)}
                onCancelDeadline={() => cancelDeadlineLocal(d.id)}
                onRevertCancel={() => revertCancel(d.id)}
                onUpdateSupersede={(field, value) => updateSupersede(d.id, field, value)}
              />
            ))}
          </div>
        )}
      </section>

      {/* New deadlines */}
      <section className="mb-8">
        <h2 className="mb-3 text-lg font-semibold">{t("newDeadlinesTitle")}</h2>
        {newDeadlines.length > 0 && (
          <div className="mb-3 flex flex-col gap-3">
            {newDeadlines.map((nd) => (
              <AddDeadlineCard
                key={nd.localId}
                deadline={nd}
                errors={errors}
                onUpdate={(field, value) => updateNewDeadline(nd.localId, field, value)}
                onRemove={() => removeNewDeadline(nd.localId)}
              />
            ))}
          </div>
        )}
        <button
          type="button"
          onClick={addNewDeadline}
          className="text-sm text-primary hover:underline"
        >
          {t("addDeadlineButton")}
        </button>
      </section>

      {/* Contributor */}
      <section className="mb-8 rounded-lg border border-border p-6">
        <h2 className="mb-4 text-lg font-semibold">{t("contributorTitle")}</h2>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium">{t("contributorNameLabel")}</label>
            <input
              type="text"
              value={contributor.name}
              onChange={(ev) => updateContributor("name", ev.target.value)}
              placeholder={t("contributorNamePlaceholder")}
              className={inputClass}
            />
            {errors["contributor.name"] && (
              <p className="text-xs text-red-500">{errors["contributor.name"]}</p>
            )}
          </div>
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium">{t("contributorEmailLabel")}</label>
            <input
              type="email"
              value={contributor.email}
              onChange={(ev) => updateContributor("email", ev.target.value)}
              placeholder={t("contributorEmailPlaceholder")}
              className={inputClass}
            />
            {errors["contributor.email"] && (
              <p className="text-xs text-red-500">{errors["contributor.email"]}</p>
            )}
          </div>
        </div>
      </section>

      {/* API error */}
      {apiError && (
        <div className="mb-6 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {apiError}
        </div>
      )}

      <div className="flex justify-between border-t border-border pt-4">
        <button
          type="button"
          onClick={() => router.push(`/${locale}`)}
          className="rounded-md border border-border px-4 py-2.5 text-sm font-medium text-foreground transition-colors hover:bg-muted"
        >
          ← {t("backButton")}
        </button>
        <div className="flex items-center gap-3">
          {!hasChanges && !isSubmitting && (
            <p className="text-sm text-muted-foreground">{t("noChangesHint")}</p>
          )}
          <button
            type="button"
            onClick={submitChanges}
            disabled={isSubmitting || !hasChanges}
            className="rounded-md bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {isSubmitting ? t("submittingButton") : t("submitButton")}
          </button>
        </div>
      </div>
    </main>
  )
}

// --- Component: loader (reads sessionStorage, guards redirect) ---

function DeadlineManagePage(): JSX.Element {
  const t = useTranslations("deadlines.manage")
  const router = useRouter()
  const locale = useLocale()
  const [event, setEvent] = useState<EventListItem | null>(null)
  const [loaded, setLoaded] = useState(false)

  useEffect(() => {
    const raw = sessionStorage.getItem(SESSION_KEY)
    if (!raw) {
      router.replace(`/${locale}`)
      return
    }
    try {
      setEvent(JSON.parse(raw) as EventListItem)
    } catch {
      router.replace(`/${locale}`)
    }
    setLoaded(true)
  }, [router, locale])

  if (!loaded || !event) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-muted-foreground">{t("loading")}</p>
      </main>
    )
  }

  return <DeadlineManageContent event={event} />
}

// --- Export ---

export { DeadlineManagePage }
