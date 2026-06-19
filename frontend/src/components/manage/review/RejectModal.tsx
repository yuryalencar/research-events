"use client"

import type { JSX } from "react"
import { useState } from "react"
import { useTranslations } from "next-intl"

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"

// --- Types ---

interface RejectModalProps {
  isOpen: boolean
  isSubmitting: boolean
  onConfirm: (reason: string) => Promise<void>
  onClose: () => void
}

// --- Component ---

function RejectModal({ isOpen, isSubmitting, onConfirm, onClose }: RejectModalProps): JSX.Element {
  const t = useTranslations("manage.review.rejectModal")
  const [reason, setReason] = useState("")

  const isReasonValid = reason.trim().length > 0

  async function handleConfirm(): Promise<void> {
    if (!isReasonValid) return
    await onConfirm(reason)
  }

  function handleClose(): void {
    setReason("")
    onClose()
  }

  return (
    <Dialog open={isOpen} onOpenChange={(open) => { if (!open) handleClose() }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("title")}</DialogTitle>
          <DialogDescription>{t("description")}</DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-3 px-4 pb-2">
          <label className="text-sm font-medium text-foreground">{t("reasonLabel")}</label>
          <textarea
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder={t("reasonPlaceholder")}
            rows={4}
            className="rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary resize-none"
          />
        </div>

        <div className="flex justify-end gap-3 border-t border-border px-4 py-3">
          <button
            type="button"
            onClick={handleClose}
            disabled={isSubmitting}
            className="rounded-md border border-border px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-muted disabled:opacity-50"
          >
            {t("cancelButton")}
          </button>
          <button
            type="button"
            onClick={handleConfirm}
            disabled={isSubmitting || !isReasonValid}
            className="rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white transition-opacity hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSubmitting ? "…" : t("confirmButton")}
          </button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

// --- Export ---

export { RejectModal }
export type { RejectModalProps }
