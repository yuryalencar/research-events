"use client"

import type { JSX, ChangeEvent } from "react"
import dynamic from "next/dynamic"
import { useTranslations } from "next-intl"

import { COUNTRIES } from "@/lib/countries"
import { DOMAINS, TIERS } from "@/lib/constants"
import type { WizardFormData, WizardErrors } from "@/hooks/useSubmitWizard"

// LocationPicker requires browser APIs — must never be server-rendered.
const LocationPicker = dynamic(
  () => import("@/components/map/LocationPicker").then(mod => mod.LocationPicker),
  { ssr: false },
)

// --- Types ---

interface Step2DetailsProps {
  formData: WizardFormData
  errors: WizardErrors
  onBack: () => void
  onContinue: () => void
  onChange: (field: keyof WizardFormData, value: unknown) => void
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
  "rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
const errorInputClass =
  "rounded-md border border-red-400 bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-red-400"

// --- Component ---

function Step2Details({ formData, errors, onBack, onContinue, onChange }: Step2DetailsProps): JSX.Element {
  const t = useTranslations("submit.step2")
  const tSubmit = useTranslations("submit")

  const field = (name: keyof WizardFormData) => (e: ChangeEvent<HTMLInputElement | HTMLSelectElement>): void => {
    onChange(name, e.target.value)
  }

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h2 className="text-xl font-semibold">{t("title")}</h2>
        <p className="mt-1 text-sm text-muted-foreground">{t("description")}</p>
      </div>

      {/* Submitter section */}
      <section className="flex flex-col gap-4">
        <h3 className="text-base font-medium text-foreground">{t("submitterSection")}</h3>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field label={t("submitterName")} error={errors.submitterName} required>
            <input
              type="text"
              value={formData.submitterName}
              onChange={field("submitterName")}
              autoComplete="name"
              className={errors.submitterName ? errorInputClass : inputClass}
            />
          </Field>
          <Field label={t("submitterEmail")} error={errors.submitterEmail} required>
            <input
              type="email"
              value={formData.submitterEmail}
              onChange={field("submitterEmail")}
              autoComplete="email"
              className={errors.submitterEmail ? errorInputClass : inputClass}
            />
          </Field>
        </div>
      </section>

      {/* Event section */}
      <section className="flex flex-col gap-4">
        <h3 className="text-base font-medium text-foreground">{t("eventSection")}</h3>

        <Field label={t("fullName")} error={errors.fullName} hint={t("fullNameHint")} required>
          <input
            type="text"
            value={formData.fullName}
            onChange={field("fullName")}
            className={errors.fullName ? errorInputClass : inputClass}
          />
        </Field>

        <Field label={t("slug")} error={errors.slug} hint={t("slugHint")} required>
          <input
            type="text"
            value={formData.slug}
            onChange={field("slug")}
            className={errors.slug ? errorInputClass : inputClass}
          />
        </Field>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field label={t("domain")} error={errors.domain} required>
            <select
              value={formData.domain}
              onChange={field("domain")}
              className={errors.domain ? errorInputClass : inputClass}
            >
              {DOMAINS.map(d => (
                <option key={d} value={d}>
                  {d === "computer_science" ? "Computer Science" : d}
                </option>
              ))}
            </select>
          </Field>

          <Field label={t("tier")} error={errors.tier}>
            <select
              value={formData.tier}
              onChange={field("tier")}
              className={inputClass}
            >
              <option value="">{t("tierPlaceholder")}</option>
              {TIERS.map(tier => (
                <option key={tier} value={tier}>
                  {tier}
                </option>
              ))}
            </select>
          </Field>
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field label={t("startDate")} error={errors.startDate} required>
            <input
              type="date"
              value={formData.startDate}
              onChange={field("startDate")}
              className={errors.startDate ? errorInputClass : inputClass}
            />
          </Field>
          <Field label={t("endDate")} error={errors.endDate} required>
            <input
              type="date"
              value={formData.endDate}
              onChange={field("endDate")}
              className={errors.endDate ? errorInputClass : inputClass}
            />
          </Field>
        </div>

        <Field label={t("websiteUrl")} error={errors.websiteUrl} required>
          <input
            type="url"
            value={formData.websiteUrl}
            onChange={field("websiteUrl")}
            placeholder="https://"
            className={errors.websiteUrl ? errorInputClass : inputClass}
          />
        </Field>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field label={t("country")} error={errors.country} required>
            <select
              value={formData.country}
              onChange={field("country")}
              className={errors.country ? errorInputClass : inputClass}
            >
              <option value="">{t("countryPlaceholder")}</option>
              {COUNTRIES.map(c => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>
          </Field>
          <Field label={t("city")} error={errors.city} required>
            <input
              type="text"
              value={formData.city}
              onChange={field("city")}
              className={errors.city ? errorInputClass : inputClass}
            />
          </Field>
        </div>

        {/* Leaflet map */}
        <Field label={t("mapLabel")} error={errors.location} required>
          <p className="text-xs text-muted-foreground">{t("mapInstruction")}</p>
          <LocationPicker
            latitude={formData.latitude}
            longitude={formData.longitude}
            country={formData.country || null}
            onChange={(lat, lng) => {
              onChange("latitude", lat)
              onChange("longitude", lng)
            }}
          />
          {formData.latitude !== null && formData.longitude !== null && (
            <p className="text-xs text-muted-foreground">
              {t("latLng", {
                lat: formData.latitude.toFixed(5),
                lng: formData.longitude.toFixed(5),
              })}
            </p>
          )}
        </Field>
      </section>

      {/* Navigation */}
      <div className="flex justify-between border-t border-border pt-4">
        <button
          type="button"
          onClick={onBack}
          className="rounded-md border border-border px-4 py-2.5 text-sm font-medium text-foreground transition-colors hover:bg-muted"
        >
          ← {tSubmit("back")}
        </button>
        <button
          type="button"
          onClick={onContinue}
          className="rounded-md bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
        >
          {t("continueButton")}
        </button>
      </div>
    </div>
  )
}

// --- Export ---

export { Step2Details }
export type { Step2DetailsProps }
