"use client"

import type { JSX } from "react"
import { useEffect, useMemo, useState } from "react"
import { useLocale, useTranslations } from "next-intl"

import { MIN_FILTER_YEAR, DOMAINS, TIERS } from "@/lib/constants"
import { COUNTRIES } from "@/lib/countries"
import type { UseFiltersReturn } from "@/hooks/useFilters"

// --- Types ---

interface FilterPanelProps {
  filters: UseFiltersReturn
}

// DESKTOP_QUERY determines when the panel starts expanded by default.
const DESKTOP_QUERY = "(min-width: 768px)"

// --- Component ---

function FilterPanel({ filters }: FilterPanelProps): JSX.Element {
  const t = useTranslations("filters")
  // locale from next-intl (e.g. "pt", "de") so month names match the UI language.
  const locale = useLocale()

  // The panel starts expanded on desktop and collapsed on mobile, but only
  // after mount (SSR snapshot always returns false, so we can't use
  // useMediaQuery's return value as useState initializer).
  const [isOpen, setIsOpen] = useState(false)

  useEffect(() => {
    setIsOpen(window.matchMedia(DESKTOP_QUERY).matches)
  }, [])

  // Month names re-computed when the active locale changes. First letter is
  // capitalised because some locales (e.g. pt) return lowercase month names.
  const months = useMemo(
    () =>
      Array.from({ length: 12 }, (_, i) => {
        const raw = new Date(2000, i, 1).toLocaleString(locale, { month: "long" })
        return { value: i + 1, label: raw.charAt(0).toUpperCase() + raw.slice(1) }
      }),
    [locale],
  )

  const {
    draftFilters,
    isDirty,
    setYear,
    setDomain,
    setTier,
    setCountry,
    setFirstDeadlineMonth,
    apply,
    reset,
  } = filters

  // On mobile, close the panel after applying so the globe is fully visible.
  function handleApply(): void {
    apply()
    if (!window.matchMedia(DESKTOP_QUERY).matches) {
      setIsOpen(false)
    }
  }

  return (
    <div className="fixed top-4 left-4 z-40 flex flex-col items-start gap-2">
      {/* Toggle button */}
      <button
        type="button"
        aria-label={t("toggleLabel")}
        aria-expanded={isOpen}
        onClick={() => setIsOpen((prev) => !prev)}
        className="relative flex h-9 w-9 items-center justify-center rounded-full bg-white/90 shadow-md backdrop-blur-sm transition hover:bg-white focus:outline-none focus-visible:ring-2 focus-visible:ring-white"
      >
        {/* Filter icon (funnel) */}
        <svg
          aria-hidden="true"
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="text-gray-700"
        >
          <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" />
        </svg>

        {/* Active-filter dot indicator */}
        {isDirty && (
          <span
            aria-label={t("activeFiltersIndicator")}
            className="absolute top-0 right-0 h-2.5 w-2.5 rounded-full bg-blue-500 ring-2 ring-white"
          />
        )}
      </button>

      {/* Filter panel */}
      {isOpen && (
        <div className="w-64 rounded-xl bg-white/95 p-4 shadow-lg backdrop-blur-sm">
          <p className="mb-3 text-sm font-semibold text-gray-800">{t("title")}</p>

          {/* Year stepper */}
          <div className="mb-3">
            <label className="mb-1 block text-xs font-medium text-gray-600">{t("yearLabel")}</label>
            <div className="flex items-center gap-2">
              <button
                type="button"
                aria-label={t("prevYear")}
                disabled={draftFilters.year <= MIN_FILTER_YEAR}
                onClick={() => setYear(draftFilters.year - 1)}
                className="flex h-7 w-7 items-center justify-center rounded-md bg-gray-100 text-gray-700 transition hover:bg-gray-200 disabled:cursor-not-allowed disabled:opacity-40"
              >
                ‹
              </button>
              <span className="flex-1 text-center text-sm font-medium text-gray-800">{draftFilters.year}</span>
              <button
                type="button"
                aria-label={t("nextYear")}
                onClick={() => setYear(draftFilters.year + 1)}
                className="flex h-7 w-7 items-center justify-center rounded-md bg-gray-100 text-gray-700 transition hover:bg-gray-200"
              >
                ›
              </button>
            </div>
          </div>

          {/* Domain dropdown */}
          <div className="mb-3">
            <label className="mb-1 block text-xs font-medium text-gray-600">{t("domainLabel")}</label>
            <select
              value={draftFilters.domain ?? ""}
              onChange={(e) => setDomain(e.target.value || undefined)}
              className="w-full rounded-md border border-gray-200 bg-white px-2 py-1.5 text-sm text-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-400"
            >
              <option value="">{t("allDomains")}</option>
              {DOMAINS.map((d) => (
                <option key={d} value={d}>
                  {t(`domains.${d}` as Parameters<typeof t>[0])}
                </option>
              ))}
            </select>
          </div>

          {/* Tier chips */}
          <div className="mb-3">
            <label className="mb-1 block text-xs font-medium text-gray-600">{t("tierLabel")}</label>
            <div className="flex flex-wrap gap-1">
              <TierChip
                label={t("allTiers")}
                active={draftFilters.tier === undefined}
                onClick={() => setTier(undefined)}
              />
              {TIERS.map((tier) => (
                <TierChip
                  key={tier}
                  label={t(`tiers.${tier}` as Parameters<typeof t>[0])}
                  active={draftFilters.tier === tier}
                  onClick={() => setTier(draftFilters.tier === tier ? undefined : tier)}
                />
              ))}
            </div>
          </div>

          {/* Country dropdown — browser-native type-to-search handles filtering */}
          <div className="mb-3">
            <label className="mb-1 block text-xs font-medium text-gray-600">{t("countryLabel")}</label>
            <select
              value={draftFilters.country ?? ""}
              onChange={(e) => setCountry(e.target.value || undefined)}
              className="w-full rounded-md border border-gray-200 bg-white px-2 py-1.5 text-sm text-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-400"
            >
              <option value="">{t("allCountries")}</option>
              {COUNTRIES.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>
          </div>

          {/* First deadline month dropdown */}
          <div className="mb-4">
            <label className="mb-1 block text-xs font-medium text-gray-600">{t("deadlineMonthLabel")}</label>
            <select
              value={draftFilters.firstDeadlineMonth ?? ""}
              onChange={(e) => setFirstDeadlineMonth(e.target.value ? Number(e.target.value) : undefined)}
              className="w-full rounded-md border border-gray-200 bg-white px-2 py-1.5 text-sm text-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-400"
            >
              <option value="">{t("allMonths")}</option>
              {months.map(({ value, label }) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </div>

          {/* Apply / Reset buttons */}
          <div className="flex gap-2">
            <button
              type="button"
              onClick={reset}
              className="flex-1 rounded-md border border-gray-200 bg-white py-1.5 text-sm font-medium text-gray-700 transition hover:bg-gray-50"
            >
              {t("reset")}
            </button>
            <button
              type="button"
              onClick={handleApply}
              className="flex-1 rounded-md bg-blue-600 py-1.5 text-sm font-medium text-white transition hover:bg-blue-700"
            >
              {t("apply")}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

// --- Sub-components ---

interface TierChipProps {
  label: string
  active: boolean
  onClick: () => void
}

function TierChip({ label, active, onClick }: TierChipProps): JSX.Element {
  return (
    <button
      type="button"
      onClick={onClick}
      className={[
        "rounded-full px-2.5 py-0.5 text-xs font-medium transition",
        active
          ? "bg-blue-600 text-white"
          : "bg-gray-100 text-gray-600 hover:bg-gray-200",
      ].join(" ")}
    >
      {label}
    </button>
  )
}

// --- Export ---

export { FilterPanel }
export type { FilterPanelProps }
