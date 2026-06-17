"use client"

import type { JSX } from "react"
import Link from "next/link"
import { useLocale, useTranslations } from "next-intl"

import { useEventSearch } from "@/hooks/useEventSearch"

// --- Types ---

interface Step1SearchProps {
  onContinue: () => void
}

// --- Component ---

function Step1Search({ onContinue }: Step1SearchProps): JSX.Element {
  const t = useTranslations("submit.step1")
  const locale = useLocale()
  const search = useEventSearch(10)

  const handleApply = (): void => {
    search.apply()
  }

  const handleClear = (): void => {
    search.clear()
  }

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-xl font-semibold">{t("title")}</h2>
        <p className="mt-1 text-sm text-muted-foreground">{t("description")}</p>
      </div>

      {/* Search controls */}
      <div className="flex flex-wrap gap-3">
        <input
          type="text"
          value={search.searchText}
          onChange={e => search.setSearchText(e.target.value)}
          placeholder={t("searchPlaceholder")}
          className="min-w-48 flex-1 rounded-md border border-border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
        />
        <div className="flex items-center gap-2">
          <label htmlFor="search-year" className="text-sm text-muted-foreground">
            {t("yearLabel")}
          </label>
          <input
            id="search-year"
            type="number"
            value={search.year}
            onChange={e => search.setYear(Number(e.target.value))}
            className="w-24 rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>
        <button
          type="button"
          onClick={handleApply}
          className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
        >
          {t("applyButton")}
        </button>
        <button
          type="button"
          onClick={handleClear}
          className="rounded-md border border-border px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-muted"
        >
          {t("clearButton")}
        </button>
      </div>

      {/* Error state */}
      {search.error && !search.isLoading && (
        <div className="rounded-md border border-yellow-400 bg-yellow-50 px-4 py-3 text-sm text-yellow-900 dark:border-yellow-700 dark:bg-yellow-950/30 dark:text-yellow-300">
          {t("errorLoad")}
        </div>
      )}

      {/* Table */}
      {search.isLoading ? (
        <div className="flex justify-center py-8">
          <div
            role="status"
            className="size-6 animate-spin rounded-full border-2 border-muted border-t-primary"
            aria-label="Loading"
          />
        </div>
      ) : search.filteredEvents.length === 0 ? (
        <p className="rounded-md border border-border bg-muted/30 px-4 py-6 text-center text-sm text-muted-foreground">
          {search.total === 0 ? t("emptySystem") : t("emptySearch")}
        </p>
      ) : (
        <div className="overflow-x-auto rounded-md border border-border">
          <table className="w-full text-sm">
            <thead className="border-b border-border bg-muted/50">
              <tr>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">
                  {t("tableNameHeader")}
                </th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">
                  {t("tableSlugHeader")}
                </th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">
                  {t("tableYearHeader")}
                </th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">
                  {t("tableCountryHeader")}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {search.filteredEvents.map(event => (
                <tr key={event.id} className="bg-background hover:bg-muted/30 transition-colors">
                  <td className="px-4 py-3">
                    <a
                      href={event.website_url}
                      target="_blank"
                      rel="noreferrer"
                      className="text-primary underline underline-offset-2"
                    >
                      {event.name}
                    </a>
                  </td>
                  <td className="px-4 py-3 font-mono text-xs">{event.slug}</td>
                  <td className="px-4 py-3">{event.year}</td>
                  <td className="px-4 py-3">{event.country}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Pagination */}
      {search.totalPages > 1 && !search.isLoading && (
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">
            {t("foundCount", { count: search.total })}
          </span>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => search.goToPage(search.page - 1)}
              disabled={search.page <= 1}
              aria-label={t("prevPage")}
              className="rounded-md border border-border px-3 py-1.5 text-sm transition-colors hover:bg-muted disabled:cursor-not-allowed disabled:opacity-40"
            >
              ‹
            </button>
            <span className="text-muted-foreground">
              {t("pageIndicator", { page: search.page, total: search.totalPages })}
            </span>
            <button
              type="button"
              onClick={() => search.goToPage(search.page + 1)}
              disabled={search.page >= search.totalPages}
              aria-label={t("nextPage")}
              className="rounded-md border border-border px-3 py-1.5 text-sm transition-colors hover:bg-muted disabled:cursor-not-allowed disabled:opacity-40"
            >
              ›
            </button>
          </div>
        </div>
      )}

      {/* Navigation */}
      <div className="flex justify-between border-t border-border pt-4">
        <Link
          href={`/${locale}`}
          className="rounded-md border border-border px-4 py-2.5 text-sm font-medium text-foreground transition-colors hover:bg-muted"
        >
          ← {t("backToHome")}
        </Link>
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

export { Step1Search }
export type { Step1SearchProps }
