"use client"

import type { JSX } from "react"
import { useEffect, useRef, useState } from "react"
import { useLocale } from "next-intl"

import { usePathname, useRouter } from "@/i18n/navigation"

// --- Types ---

interface LocaleOption {
  code: string
  flag: string
  label: string
}

// --- Constants ---

const LOCALES: LocaleOption[] = [
  { code: "en", flag: "🇺🇸", label: "English" },
  { code: "pt", flag: "🇧🇷", label: "Português" },
  { code: "es", flag: "🇪🇸", label: "Español" },
  { code: "de", flag: "🇩🇪", label: "Deutsch" },
]

// --- Component ---

function LanguageSelector(): JSX.Element {
  const locale = useLocale()
  const router = useRouter()
  const pathname = usePathname()
  const [isOpen, setIsOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const current = LOCALES.find((l) => l.code === locale) ?? LOCALES[0]

  useEffect(() => {
    function handleClickOutside(e: MouseEvent): void {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener("mousedown", handleClickOutside)
    return () => document.removeEventListener("mousedown", handleClickOutside)
  }, [])

  function handleSelect(code: string): void {
    setIsOpen(false)
    router.replace(pathname, { locale: code })
  }

  return (
    <div ref={containerRef} className="fixed bottom-4 right-4 z-40 flex flex-col items-end gap-1">
      {/* Dropdown — renders above the trigger button */}
      {isOpen && (
        <div className="mb-1 w-40 rounded-xl bg-white/95 py-1 shadow-lg backdrop-blur-sm">
          {LOCALES.map(({ code, flag, label }) => {
            const isActive = code === locale
            return (
              <button
                key={code}
                type="button"
                onClick={() => handleSelect(code)}
                className={[
                  "flex w-full items-center gap-2.5 px-3 py-2 text-sm transition hover:bg-gray-50",
                  isActive ? "font-semibold text-blue-600" : "text-gray-700",
                ].join(" ")}
              >
                <span aria-hidden="true" className="text-base leading-none">
                  {flag}
                </span>
                <span className="flex-1 text-left">{label}</span>
                {isActive && (
                  <svg aria-hidden="true" width="12" height="12" viewBox="0 0 12 12" fill="currentColor">
                    <path d="M10.28 2.28L4.75 7.81 1.72 4.78a.75.75 0 0 0-1.06 1.06l3.5 3.5a.75.75 0 0 0 1.06 0l6-6a.75.75 0 0 0-1.06-1.06z" />
                  </svg>
                )}
              </button>
            )
          })}
        </div>
      )}

      {/* Trigger button — shows the active locale's flag */}
      <button
        type="button"
        aria-label="Select language"
        aria-expanded={isOpen}
        onClick={() => setIsOpen((prev) => !prev)}
        className="flex h-9 w-9 items-center justify-center rounded-full bg-white/90 text-lg shadow-md backdrop-blur-sm transition hover:bg-white focus:outline-none focus-visible:ring-2 focus-visible:ring-white"
      >
        <span aria-hidden="true">{current.flag}</span>
      </button>
    </div>
  )
}

// --- Export ---

export { LanguageSelector }
export type { LocaleOption }
