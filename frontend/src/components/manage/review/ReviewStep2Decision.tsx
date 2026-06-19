"use client"

import type { JSX } from "react"
import { useState } from "react"
import { useTranslations } from "next-intl"

import { StatusBadge } from "./ReviewStep1Details"
import { ApproveModal } from "./ApproveModal"
import { RejectModal } from "./RejectModal"
import type { EventStatus } from "@/types/api"

// --- Types ---

interface ReviewStep2DecisionProps {
  status: EventStatus
  isSubmitting: boolean
  bannerError: string | null
  onBack: () => void
  onApprove: (note: string) => Promise<void>
  onReject: (reason: string) => Promise<void>
  onReviewDeadlines: () => void
}

// --- Component ---

function ReviewStep2Decision({
  status,
  isSubmitting,
  bannerError,
  onBack,
  onApprove,
  onReject,
  onReviewDeadlines,
}: ReviewStep2DecisionProps): JSX.Element {
  const t = useTranslations("manage.review")
  const [approveOpen, setApproveOpen] = useState(false)
  const [rejectOpen, setRejectOpen] = useState(false)

  const isApproved = status === "approved"

  async function handleApprove(note: string): Promise<void> {
    await onApprove(note)
    setApproveOpen(false)
  }

  async function handleReject(reason: string): Promise<void> {
    await onReject(reason)
    setRejectOpen(false)
  }

  return (
    <div className="flex flex-col gap-8">
      {/* Current status */}
      <div className="flex flex-col gap-2">
        <p className="text-sm text-muted-foreground">{t("currentStatusLabel")}</p>
        <StatusBadge status={status} label={t(`statusBadge.${status}`)} />
      </div>

      {/* Decision buttons — Reject left, Approve right */}
      <div className="flex flex-col gap-4 sm:flex-row">
        <button
          type="button"
          onClick={() => setRejectOpen(true)}
          disabled={isSubmitting}
          className="flex-1 cursor-pointer rounded-md bg-red-600 px-6 py-3 text-sm font-semibold text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {t("rejectButton")}
        </button>
        <button
          type="button"
          onClick={() => setApproveOpen(true)}
          disabled={isSubmitting}
          className="flex-1 cursor-pointer rounded-md bg-green-600 px-6 py-3 text-sm font-semibold text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {t("approveButton")}
        </button>
      </div>

      {/* Banner error */}
      {bannerError && (
        <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {bannerError}
        </div>
      )}

      {/* Footer — Back (left, muted outline) + Review Deadlines (right, primary outline) */}
      <div className="flex flex-col gap-1 border-t border-border pt-4">
        <div className="flex items-center justify-between gap-4">
          <button
            type="button"
            onClick={onBack}
            className="cursor-pointer rounded-md border border-border px-4 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted"
          >
            {t("backButton")}
          </button>
          <button
            type="button"
            onClick={onReviewDeadlines}
            disabled={!isApproved}
            className="cursor-pointer rounded-md border border-primary px-5 py-2 text-sm font-medium text-primary transition-opacity hover:opacity-80 disabled:cursor-not-allowed disabled:opacity-40"
          >
            {t("reviewDeadlinesButton")}
          </button>
        </div>
        {!isApproved && (
          <p className="text-right text-xs text-muted-foreground">
            {t("reviewDeadlinesDisabled")}
          </p>
        )}
      </div>

      {/* Modals */}
      <ApproveModal
        isOpen={approveOpen}
        isSubmitting={isSubmitting}
        onConfirm={handleApprove}
        onClose={() => setApproveOpen(false)}
      />
      <RejectModal
        isOpen={rejectOpen}
        isSubmitting={isSubmitting}
        onConfirm={handleReject}
        onClose={() => setRejectOpen(false)}
      />
    </div>
  )
}

// --- Export ---

export { ReviewStep2Decision }
export type { ReviewStep2DecisionProps }
