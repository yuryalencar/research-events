"use client"

import type { JSX } from "react"
import { useTranslations } from "next-intl"
import { XIcon } from "lucide-react"

import type { DeadlineFormRow, WizardErrors } from "@/hooks/useSubmitWizard"

// --- Types ---

interface Step3DeadlinesProps {
  deadlines: DeadlineFormRow[]
  errors: WizardErrors
  isSubmitting: boolean
  bannerError: string | null
  onBack: () => void
  onAddDeadline: () => void
  onRemoveDeadline: (id: string) => void
  onUpdateDeadline: (id: string, field: keyof Omit<DeadlineFormRow, "id">, value: unknown) => void
  onSubmit: () => void
  onSkipAndSubmit: () => void
}

// --- Constants ---

const DEADLINE_TYPES = ["abstract", "paper", "notification", "camera_ready", "other"] as const

const inputClass =
  "rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"

// --- Component ---

function Step3Deadlines({
  deadlines,
  isSubmitting,
  bannerError,
  onBack,
  onAddDeadline,
  onRemoveDeadline,
  onUpdateDeadline,
  onSubmit,
  onSkipAndSubmit,
}: Step3DeadlinesProps): JSX.Element {
  const t = useTranslations("submit.step3")
  const tSubmit = useTranslations("submit")

  const hasDeadlineErrors = deadlines.some(
    d => !d.type || !d.description.trim() || !d.date,
  )

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-xl font-semibold">{t("title")}</h2>
        <p className="mt-1 text-sm text-muted-foreground">{t("description")}</p>
      </div>

      {/* Banner error */}
      {bannerError && (
        <div className="rounded-md border border-red-400 bg-red-50 px-4 py-3 text-sm text-red-900 dark:border-red-700 dark:bg-red-950/30 dark:text-red-200">
          {bannerError}
        </div>
      )}

      {/* Empty state */}
      {deadlines.length === 0 && (
        <p className="rounded-md border border-border bg-muted/30 px-4 py-6 text-center text-sm text-muted-foreground">
          {t("emptyState")}
        </p>
      )}

      {/* Deadline rows */}
      {deadlines.length > 0 && (
        <div className="flex flex-col gap-3">
          {deadlines.map(deadline => (
            <DeadlineRow
              key={deadline.id}
              deadline={deadline}
              onUpdate={(field, value) => onUpdateDeadline(deadline.id, field, value)}
              onRemove={() => onRemoveDeadline(deadline.id)}
            />
          ))}
        </div>
      )}

      {/* Add deadline */}
      <button
        type="button"
        onClick={onAddDeadline}
        className="self-start rounded-md border border-dashed border-border px-4 py-2 text-sm text-muted-foreground transition-colors hover:border-primary hover:text-primary"
      >
        {t("addDeadline")}
      </button>

      {/* Navigation */}
      <div className="flex justify-between border-t border-border pt-4">
        <button
          type="button"
          onClick={onBack}
          disabled={isSubmitting}
          className="rounded-md border border-border px-4 py-2.5 text-sm font-medium text-foreground transition-colors hover:bg-muted disabled:opacity-40"
        >
          ← {tSubmit("back")}
        </button>

        <div className="flex gap-3">
          <button
            type="button"
            onClick={onSkipAndSubmit}
            disabled={isSubmitting}
            className="rounded-md border border-border px-4 py-2.5 text-sm font-medium text-foreground transition-colors hover:bg-muted disabled:opacity-40"
          >
            {isSubmitting ? "…" : t("skipButton")}
          </button>
          <button
            type="button"
            onClick={onSubmit}
            disabled={isSubmitting || hasDeadlineErrors}
            className="rounded-md bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-40"
          >
            {isSubmitting ? "…" : t("submitButton")}
          </button>
        </div>
      </div>
    </div>
  )
}

// --- Sub-components ---

interface DeadlineRowProps {
  deadline: DeadlineFormRow
  onUpdate: (field: keyof Omit<DeadlineFormRow, "id">, value: unknown) => void
  onRemove: () => void
}

function DeadlineRow({ deadline, onUpdate, onRemove }: DeadlineRowProps): JSX.Element {
  const t = useTranslations("submit.step3")

  return (
    <div className="grid grid-cols-1 gap-3 rounded-md border border-border p-4 sm:grid-cols-[1fr_1fr_auto_auto_auto]">
      <div className="flex flex-col gap-1">
        <label className="text-xs font-medium text-muted-foreground">{t("typeLabel")}</label>
        <select
          value={deadline.type}
          onChange={e => onUpdate("type", e.target.value)}
          className={inputClass}
        >
          <option value="">{t("typePlaceholder")}</option>
          {DEADLINE_TYPES.map(dt => (
            <option key={dt} value={dt}>
              {t(`deadlineTypes.${dt}`)}
            </option>
          ))}
        </select>
      </div>

      <div className="flex flex-col gap-1">
        <label className="text-xs font-medium text-muted-foreground">{t("descriptionLabel")}</label>
        <input
          type="text"
          value={deadline.description}
          onChange={e => onUpdate("description", e.target.value)}
          placeholder={t("descriptionPlaceholder")}
          className={inputClass}
        />
      </div>

      <div className="flex flex-col gap-1">
        <label className="text-xs font-medium text-muted-foreground">{t("dateLabel")}</label>
        <input
          type="date"
          value={deadline.date}
          onChange={e => onUpdate("date", e.target.value)}
          className={inputClass}
        />
      </div>

      <div className="flex items-end gap-2 pb-0.5">
        <label className="flex items-center gap-2 text-sm text-muted-foreground">
          <input
            type="checkbox"
            checked={deadline.isOptional}
            onChange={e => onUpdate("isOptional", e.target.checked)}
            className="size-4 rounded border-border"
          />
          {t("optionalLabel")}
        </label>
      </div>

      <div className="flex items-end pb-0.5">
        <button
          type="button"
          onClick={onRemove}
          aria-label={t("removeDeadline")}
          className="flex size-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
        >
          <XIcon className="size-4" />
        </button>
      </div>
    </div>
  )
}

// --- Export ---

export { Step3Deadlines }
export type { Step3DeadlinesProps }
