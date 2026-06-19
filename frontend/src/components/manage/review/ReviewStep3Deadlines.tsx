"use client"

import type { JSX } from "react"
import { useEffect } from "react"
import { useTranslations } from "next-intl"
import { ExternalLinkIcon } from "lucide-react"

import { useDeadlineManage } from "@/hooks/useDeadlineManage"
import { DeadlineCard } from "@/components/events/deadlines/DeadlineCard"
import { AddDeadlineCard } from "@/components/events/deadlines/AddDeadlineCard"
import { DeadlineManageSuccess } from "@/components/events/deadlines/DeadlineManageSuccess"
import type { EventListItem } from "@/types/api"

// --- Types ---

interface ReviewStep3DeadlinesProps {
  event: EventListItem
  contributor: { name: string; email: string }
  onBack: () => void
  onDone: () => void
}

// --- Component ---

function ReviewStep3Deadlines({
  event,
  contributor,
  onBack,
  onDone,
}: ReviewStep3DeadlinesProps): JSX.Element {
  const t = useTranslations("manage.review.step3")
  const tDeadlines = useTranslations("deadlines.manage")

  const {
    deadlineStates,
    newDeadlines,
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

  // Pre-fill contributor from the reviewer's session — hidden from UI since
  // the admin/moderator doesn't need to fill in their own name and email.
  useEffect(() => {
    updateContributor("name", contributor.name)
    updateContributor("email", contributor.email)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  if (pageState === "success" && successSummary) {
    return <DeadlineManageSuccess event={event} summary={successSummary} />
  }

  const isSubmitting = pageState === "submitting"
  const activeDeadlines = event.deadlines.filter((d) => d.is_active)

  return (
    <div className="flex flex-col gap-8">
      {/* Conference URL */}
      <div className="flex items-center gap-2 rounded-lg border border-border bg-card px-4 py-3">
        <ExternalLinkIcon size={14} className="shrink-0 text-muted-foreground" />
        <a
          href={event.website_url}
          target="_blank"
          rel="noopener noreferrer"
          className="text-sm text-primary underline-offset-4 hover:underline"
        >
          {t("conferenceLink")} — {event.website_url}
        </a>
      </div>

      {/* Existing deadlines */}
      <section>
        <h2 className="mb-3 text-base font-semibold">{tDeadlines("existingDeadlinesTitle")}</h2>
        {activeDeadlines.length === 0 ? (
          <p className="text-sm text-muted-foreground">{tDeadlines("noDeadlines")}</p>
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
      <section>
        <h2 className="mb-3 text-base font-semibold">{tDeadlines("newDeadlinesTitle")}</h2>
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
          {tDeadlines("addDeadlineButton")}
        </button>
      </section>

      {/* API error */}
      {apiError && (
        <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {apiError}
        </div>
      )}

      {/* Footer */}
      <div className="flex justify-between border-t border-border pt-4">
        <button
          type="button"
          onClick={onBack}
          className="cursor-pointer rounded-md border border-border px-4 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted"
        >
          {t("backButton")}
        </button>
        <div className="flex items-center gap-3">
          {hasChanges && (
            <button
              type="button"
              onClick={submitChanges}
              disabled={isSubmitting}
              className="cursor-pointer rounded-md bg-primary px-5 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {isSubmitting ? tDeadlines("submittingButton") : tDeadlines("submitButton")}
            </button>
          )}
          <button
            type="button"
            onClick={onDone}
            className="cursor-pointer rounded-md border border-primary px-5 py-2 text-sm font-medium text-primary transition-colors hover:bg-primary/10"
          >
            {t("doneButton")}
          </button>
        </div>
      </div>
    </div>
  )
}

// --- Export ---

export { ReviewStep3Deadlines }
export type { ReviewStep3DeadlinesProps }
