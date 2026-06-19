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

interface ApproveModalProps {
  isOpen: boolean
  isSubmitting: boolean
  onConfirm: (note: string) => Promise<void>
  onClose: () => void
}

// --- Component ---

function ApproveModal({ isOpen, isSubmitting, onConfirm, onClose }: ApproveModalProps): JSX.Element {
  const t = useTranslations("manage.review.approveModal")
  const [note, setNote] = useState("")

  async function handleConfirm(): Promise<void> {
    await onConfirm(note)
  }

  function handleClose(): void {
    setNote("")
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
          <label className="text-sm font-medium text-foreground">{t("noteLabel")}</label>
          <textarea
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder={t("notePlaceholder")}
            rows={3}
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
            disabled={isSubmitting}
            className="rounded-md bg-green-600 px-4 py-2 text-sm font-medium text-white transition-opacity hover:opacity-90 disabled:opacity-50"
          >
            {isSubmitting ? "…" : t("confirmButton")}
          </button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

// --- Export ---

export { ApproveModal }
export type { ApproveModalProps }
