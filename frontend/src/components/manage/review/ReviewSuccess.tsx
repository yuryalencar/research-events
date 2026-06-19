"use client"

import type { JSX } from "react"
import { CheckCircleIcon, XCircleIcon } from "lucide-react"
import { useTranslations } from "next-intl"

import { Badge } from "@/components/ui/badge"
import type { ReviewSuccessState } from "@/hooks/useReviewWizard"

// --- Types ---

interface ReviewSuccessProps {
  state: ReviewSuccessState
  onManageDeadlines: () => void
  onBackToDashboard: () => void
}

// --- Component ---

function ReviewSuccess({ state, onManageDeadlines, onBackToDashboard }: ReviewSuccessProps): JSX.Element {
  const t = useTranslations("manage.review.success")
  const isApprove = state.action === "approve"

  return (
    <div className="flex flex-col items-center gap-8 py-12 text-center">
      {isApprove ? (
        <CheckCircleIcon className="size-16 text-green-500" />
      ) : (
        <XCircleIcon className="size-16 text-red-500" />
      )}

      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold">
          {isApprove ? t("approveTitle") : t("rejectTitle")}
        </h1>
        <p className="max-w-md text-muted-foreground">
          {isApprove ? t("approveDescription") : t("rejectDescription")}
        </p>
      </div>

      {/* Summary card */}
      <div className="w-full max-w-md rounded-lg border border-border bg-card p-6 text-left">
        <h2 className="mb-4 font-semibold text-card-foreground">{state.event.name}</h2>
        <dl className="flex flex-col gap-3 text-sm">
          <div className="flex justify-between gap-2">
            <dt className="text-muted-foreground">{t("statusLabel")}</dt>
            <dd>
              <Badge variant="secondary" className={isApprove ? "text-green-700" : "text-red-700"}>
                {isApprove ? "Approved" : "Rejected"}
              </Badge>
            </dd>
          </div>

          {!isApprove && state.reason && (
            <div className="flex flex-col gap-1">
              <dt className="text-muted-foreground">{t("reasonLabel")}</dt>
              <dd className="text-foreground">{state.reason}</dd>
            </div>
          )}
        </dl>
      </div>

      {/* Actions */}
      <div className="flex flex-col items-center gap-3 sm:flex-row">
        {isApprove && (
          <button
            type="button"
            onClick={onManageDeadlines}
            className="rounded-md bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
          >
            {t("manageDeadlinesButton")}
          </button>
        )}
        <button
          type="button"
          onClick={onBackToDashboard}
          className="rounded-md border border-border px-6 py-2.5 text-sm font-medium text-foreground transition-colors hover:bg-muted"
        >
          {t("backToDashboardButton")}
        </button>
      </div>
    </div>
  )
}

// --- Export ---

export { ReviewSuccess }
export type { ReviewSuccessProps }
