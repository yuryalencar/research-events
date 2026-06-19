"use client"

import type { JSX, ChangeEvent } from "react"
import dynamic from "next/dynamic"
import { useTranslations } from "next-intl"

import { COUNTRIES } from "@/lib/countries"
import { DOMAINS, TIERS } from "@/lib/constants"
import type { ReviewFormData } from "@/hooks/useReviewWizard"
import type { EventStatus } from "@/types/api"

// LocationPicker requires browser APIs — must never be server-rendered.
const LocationPicker = dynamic(
  () => import("@/components/map/LocationPicker").then((mod) => mod.LocationPicker),
  { ssr: false },
)

// --- Types ---

interface ReviewStep1DetailsProps {
  formData: ReviewFormData
  errors: Record<string, string>
  status: EventStatus
  onBack: () => void
  onNext: () => void
  onChange: (field: keyof ReviewFormData, value: string | number) => void
}

// --- Sub-components ---

interface FieldProps {
  label: string
  error?: string
  hint?: string
  required?: boolean
  children: React.ReactNode
}

function Field({ label, error, hint, required, children }: FieldProps): JSX.Element {
  return (
    <div className="flex flex-col gap-1">
      <label className="text-sm font-medium text-foreground">
        {label}
        {required && <span className="ml-0.5 text-red-500">*</span>}
      </label>
      {children}
      {hint && !error && <p className="text-xs text-muted-foreground">{hint}</p>}
      {error && <p className="text-xs text-red-500">{error}</p>}
    </div>
  )
}

const inputClass =
  "rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
const errorInputClass =
  "rounded-md border border-red-400 bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-red-400"

// --- Status badge ---

const STATUS_STYLES: Record<EventStatus, string> = {
  pending: "bg-yellow-100 text-yellow-800 border-yellow-300",
  approved: "bg-green-100 text-green-800 border-green-300",
  rejected: "bg-red-100 text-red-800 border-red-300",
}

function StatusBadge({ status, label }: { status: EventStatus; label: string }): JSX.Element {
  return (
    <span
      className={`inline-flex items-center rounded-full border px-3 py-0.5 text-xs font-semibold capitalize ${STATUS_STYLES[status]}`}
    >
      {label}
    </span>
  )
}

// --- Component ---

function ReviewStep1Details({
  formData,
  errors,
  status,
  onBack,
  onNext,
  onChange,
}: ReviewStep1DetailsProps): JSX.Element {
  const t = useTranslations("manage.review")

  const field =
    (name: keyof ReviewFormData) =>
    (e: ChangeEvent<HTMLInputElement | HTMLSelectElement>): void => {
      onChange(name, e.target.value)
    }

  return (
    <div className="flex flex-col gap-8">
      {/* Status badge */}
      <div className="flex items-center gap-2">
        <span className="text-sm text-muted-foreground">{t("currentStatusLabel")}:</span>
        <StatusBadge status={status} label={t(`statusBadge.${status}`)} />
      </div>

      {/* Event fields */}
      <section className="flex flex-col gap-4">
        <Field label={t("fields.nameLabel")} error={errors.name} required>
          <input
            type="text"
            value={formData.name}
            onChange={field("name")}
            className={errors.name ? errorInputClass : inputClass}
          />
        </Field>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field label={t("fields.domainLabel")} required>
            <select
              value={formData.domain}
              onChange={field("domain")}
              className={inputClass}
            >
              {DOMAINS.map((d) => (
                <option key={d} value={d}>
                  {d === "computer_science" ? "Computer Science" : d}
                </option>
              ))}
            </select>
          </Field>

          <Field label={t("fields.tierLabel")}>
            <select
              value={formData.tier}
              onChange={field("tier")}
              className={inputClass}
            >
              <option value="">—</option>
              {TIERS.map((tier) => (
                <option key={tier} value={tier}>
                  {tier}
                </option>
              ))}
            </select>
          </Field>
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field label={t("fields.countryLabel")} error={errors.country} required>
            <select
              value={formData.country}
              onChange={field("country")}
              className={errors.country ? errorInputClass : inputClass}
            >
              <option value="">—</option>
              {COUNTRIES.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>
          </Field>

          <Field label={t("fields.cityLabel")} error={errors.city} required>
            <input
              type="text"
              value={formData.city}
              onChange={field("city")}
              className={errors.city ? errorInputClass : inputClass}
            />
          </Field>
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field label={t("fields.startDateLabel")} error={errors.startDate} required>
            <input
              type="date"
              value={formData.startDate}
              onChange={field("startDate")}
              className={errors.startDate ? errorInputClass : inputClass}
            />
          </Field>

          <Field label={t("fields.endDateLabel")} error={errors.endDate} required>
            <input
              type="date"
              value={formData.endDate}
              onChange={field("endDate")}
              className={errors.endDate ? errorInputClass : inputClass}
            />
          </Field>
        </div>

        <Field label={t("fields.websiteLabel")} error={errors.websiteUrl} required>
          <input
            type="url"
            value={formData.websiteUrl}
            onChange={field("websiteUrl")}
            className={errors.websiteUrl ? errorInputClass : inputClass}
          />
        </Field>
      </section>

      {/* Location */}
      <section className="flex flex-col gap-3">
        <Field label={t("fields.locationLabel")} hint={t("fields.locationHint")}>
          <LocationPicker
            latitude={formData.latitude}
            longitude={formData.longitude}
            country={formData.country}
            onChange={(lat, lng) => {
              onChange("latitude", lat)
              onChange("longitude", lng)
            }}
          />
        </Field>
      </section>

      {/* Footer */}
      <div className="flex justify-between border-t border-border pt-4">
        <button
          type="button"
          onClick={onBack}
          className="cursor-pointer rounded-md border border-border px-4 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted"
        >
          {t("backToDashboard")}
        </button>
        <button
          type="button"
          onClick={onNext}
          className="cursor-pointer rounded-md bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
        >
          {t("nextButton")}
        </button>
      </div>
    </div>
  )
}

// --- Export ---

export { ReviewStep1Details, StatusBadge }
export type { ReviewStep1DetailsProps }
