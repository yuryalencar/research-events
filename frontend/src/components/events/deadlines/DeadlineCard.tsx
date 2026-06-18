"use client"

import type { JSX } from "react"
import { useTranslations } from "next-intl"
import { PencilIcon, XCircleIcon, UndoIcon } from "lucide-react"

import type { DeadlineResponse } from "@/types/api"
import type { DeadlineState, EditValues } from "@/hooks/useDeadlineManage"

// --- Types ---

interface DeadlineCardProps {
  deadline: DeadlineResponse
  state: DeadlineState
  errors: Record<string, string>
  onStartSupersede: () => void
  onRevertSupersede: () => void
  onCancelDeadline: () => void
  onRevertCancel: () => void
  onUpdateSupersede: (field: keyof EditValues, value: string) => void
}

// --- Constants ---

const inputClass =
  "rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"

// --- Component ---

function DeadlineCard({
  deadline,
  state,
  errors,
  onStartSupersede,
  onRevertSupersede,
  onCancelDeadline,
  onRevertCancel,
  onUpdateSupersede,
}: DeadlineCardProps): JSX.Element {
  const t = useTranslations("deadlines.manage")

  const err = (field: string): string | undefined =>
    errors[`supersede.${deadline.id}.${field}`]

  if (state.mode === "pendingCancel") {
    return (
      <div className="flex items-center justify-between gap-4 rounded-md border border-border p-4 opacity-60">
        <div className="flex min-w-0 flex-col gap-1">
          <div className="flex items-center gap-2">
            <span className="shrink-0 rounded-full bg-muted px-2 py-0.5 text-xs font-medium">
              {t(`deadlineTypes.${deadline.type}`)}
            </span>
            <span className="truncate text-sm line-through text-muted-foreground">
              {deadline.description}
            </span>
            <span className="shrink-0 rounded-full bg-destructive/10 px-2 py-0.5 text-xs font-medium text-destructive">
              {t("pendingCancelBadge")}
            </span>
          </div>
          <p className="text-xs text-muted-foreground line-through">{deadline.date}</p>
        </div>
        <button
          type="button"
          onClick={onRevertCancel}
          className="flex shrink-0 items-center gap-1.5 rounded-md border border-border px-3 py-1.5 text-sm transition-colors hover:bg-muted"
        >
          <UndoIcon className="size-3.5" />
          {t("undoButton")}
        </button>
      </div>
    )
  }

  if (state.mode === "superseding") {
    return (
      <div className="flex flex-col gap-3 rounded-md border border-primary/40 bg-primary/5 p-4">
        {/* Muted header showing what's being superseded */}
        <div className="flex items-center gap-2 text-muted-foreground">
          <span className="shrink-0 rounded-full bg-muted px-2 py-0.5 text-xs font-medium">
            {t(`deadlineTypes.${deadline.type}`)}
          </span>
          <span className="truncate text-sm">{deadline.description}</span>
          <span className="ml-auto shrink-0 text-xs">{deadline.date}</span>
        </div>

        {/* Edit fields */}
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-[1fr_auto_auto_auto]">
          <div className="flex flex-col gap-1">
            <label className="text-xs font-medium text-muted-foreground">{t("dateLabel")}</label>
            <input
              type="date"
              value={state.editValues.date}
              onChange={(ev) => onUpdateSupersede("date", ev.target.value)}
              className={inputClass}
            />
            {err("date") && <p className="mt-1 text-xs text-red-500">{err("date")}</p>}
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs font-medium text-muted-foreground">{t("timeLabel")}</label>
            <input
              type="text"
              value={state.editValues.time}
              onChange={(ev) => onUpdateSupersede("time", ev.target.value)}
              placeholder={t("timePlaceholder")}
              className={inputClass}
            />
            {err("time") && <p className="mt-1 text-xs text-red-500">{err("time")}</p>}
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-xs font-medium text-muted-foreground">{t("timezoneLabel")}</label>
            <input
              type="text"
              value={state.editValues.timezone}
              onChange={(ev) => onUpdateSupersede("timezone", ev.target.value)}
              placeholder={t("timezonePlaceholder")}
              className={inputClass}
            />
          </div>

          <div className="flex items-end pb-0.5">
            <button
              type="button"
              onClick={onRevertSupersede}
              className="flex items-center gap-1.5 rounded-md border border-border px-3 py-2 text-sm transition-colors hover:bg-muted"
            >
              <UndoIcon className="size-3.5" />
              {t("revertButton")}
            </button>
          </div>
        </div>
      </div>
    )
  }

  // Default state
  return (
    <div className="flex items-center justify-between gap-4 rounded-md border border-border p-4">
      <div className="flex min-w-0 flex-col gap-1">
        <div className="flex items-center gap-2">
          <span className="shrink-0 rounded-full bg-muted px-2 py-0.5 text-xs font-medium">
            {t(`deadlineTypes.${deadline.type}`)}
          </span>
          <span className="truncate text-sm font-medium">{deadline.description}</span>
          {deadline.is_optional && (
            <span className="shrink-0 rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">
              {t("optionalLabel")}
            </span>
          )}
        </div>
        <p className="text-xs text-muted-foreground">
          {deadline.date}
          {deadline.time && <span> · {deadline.time}</span>}
          {deadline.timezone && <span> {deadline.timezone}</span>}
        </p>
      </div>

      <div className="flex shrink-0 items-center gap-2">
        <button
          type="button"
          onClick={onStartSupersede}
          className="flex items-center gap-1.5 rounded-md border border-border px-3 py-1.5 text-sm transition-colors hover:bg-muted"
        >
          <PencilIcon className="size-3.5" />
          {t("supersedeButton")}
        </button>
        <button
          type="button"
          onClick={onCancelDeadline}
          className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm text-destructive transition-colors hover:bg-destructive/10"
        >
          <XCircleIcon className="size-3.5" />
          {t("cancelDeadlineButton")}
        </button>
      </div>
    </div>
  )
}

// --- Export ---

export { DeadlineCard }
export type { DeadlineCardProps }
