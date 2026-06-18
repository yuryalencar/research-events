"use client"

import type { JSX } from "react"
import { useTranslations } from "next-intl"
import { XIcon } from "lucide-react"

import type { NewDeadlineRow } from "@/hooks/useDeadlineManage"
import type { DeadlineType } from "@/types/api"

// --- Types ---

interface AddDeadlineCardProps {
  deadline: NewDeadlineRow
  errors: Record<string, string>
  onUpdate: (field: keyof Omit<NewDeadlineRow, "localId">, value: unknown) => void
  onRemove: () => void
}

// --- Constants ---

const DEADLINE_TYPES: DeadlineType[] = ["abstract", "paper", "notification", "camera_ready", "other"]

const inputClass =
  "rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"

const errorClass = "mt-1 text-xs text-red-500"

// --- Component ---

function AddDeadlineCard({ deadline, errors, onUpdate, onRemove }: AddDeadlineCardProps): JSX.Element {
  const t = useTranslations("deadlines.manage")

  const e = (field: string): string | undefined =>
    errors[`new.${deadline.localId}.${field}`]

  return (
    <div className="relative flex flex-col gap-3 rounded-md border border-border p-4 pr-12">
      {/* Remove — anchored to top-right corner */}
      <button
        type="button"
        onClick={onRemove}
        aria-label={t("removeLabel")}
        className="absolute right-3 top-3 flex size-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
      >
        <XIcon className="size-4" />
      </button>

      {/* Row 1: Type + Description */}
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <div className="flex flex-col gap-1">
          <label className="text-xs font-medium text-muted-foreground">{t("typeLabel")}</label>
          <select
            value={deadline.type}
            onChange={(ev) => onUpdate("type", ev.target.value)}
            className={inputClass}
          >
            <option value="">{t("typePlaceholder")}</option>
            {DEADLINE_TYPES.map((dt) => (
              <option key={dt} value={dt}>
                {t(`deadlineTypes.${dt}`)}
              </option>
            ))}
          </select>
          {e("type") && <p className={errorClass}>{e("type")}</p>}
        </div>

        <div className="flex flex-col gap-1">
          <label className="text-xs font-medium text-muted-foreground">{t("descriptionLabel")}</label>
          <input
            type="text"
            value={deadline.description}
            onChange={(ev) => onUpdate("description", ev.target.value)}
            placeholder={t("descriptionPlaceholder")}
            className={inputClass}
          />
          {e("description") && <p className={errorClass}>{e("description")}</p>}
        </div>
      </div>

      {/* Row 2: Date + Time + Timezone + Optional */}
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-[auto_auto_auto_1fr]">
        <div className="flex flex-col gap-1">
          <label className="text-xs font-medium text-muted-foreground">{t("dateLabel")}</label>
          <input
            type="date"
            value={deadline.date}
            onChange={(ev) => onUpdate("date", ev.target.value)}
            className={inputClass}
          />
          {e("date") && <p className={errorClass}>{e("date")}</p>}
        </div>

        <div className="flex flex-col gap-1">
          <label className="text-xs font-medium text-muted-foreground">{t("timeLabel")}</label>
          <input
            type="text"
            value={deadline.time}
            onChange={(ev) => onUpdate("time", ev.target.value)}
            placeholder={t("timePlaceholder")}
            className={inputClass}
          />
          {e("time") && <p className={errorClass}>{e("time")}</p>}
        </div>

        <div className="flex flex-col gap-1">
          <label className="text-xs font-medium text-muted-foreground">{t("timezoneLabel")}</label>
          <input
            type="text"
            value={deadline.timezone}
            onChange={(ev) => onUpdate("timezone", ev.target.value)}
            placeholder={t("timezonePlaceholder")}
            className={inputClass}
          />
        </div>

        <div className="flex items-end pb-0.5">
          <label className="flex items-center gap-2 text-sm text-muted-foreground">
            <input
              type="checkbox"
              checked={!deadline.isOptional}
              onChange={(ev) => onUpdate("isOptional", !ev.target.checked)}
              className="size-4 rounded border-border"
            />
            {t("requiredLabel")}
          </label>
        </div>
      </div>
    </div>
  )
}

// --- Export ---

export { AddDeadlineCard }
export type { AddDeadlineCardProps }
