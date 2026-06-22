"use client"

import type { JSX } from "react"
import { useState, useEffect, useRef } from "react"
import Image from "next/image"
import { useRouter } from "next/navigation"
import { useTranslations, useLocale } from "next-intl"

// --- Types ---

interface ManageHeaderProps {
  userName: string
  userRole: string
  onSignOut: () => void
}

// --- Component ---

function ManageHeader({ userName, userRole, onSignOut }: ManageHeaderProps): JSX.Element {
  const t = useTranslations("manage")
  const locale = useLocale()
  const router = useRouter()
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)
  const avatarInitial = userName.charAt(0).toUpperCase()

  // Close the avatar dropdown when the user clicks outside of it.
  useEffect(() => {
    if (!menuOpen) return

    function handleOutsideClick(e: MouseEvent): void {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false)
      }
    }

    document.addEventListener("mousedown", handleOutsideClick)
    return () => document.removeEventListener("mousedown", handleOutsideClick)
  }, [menuOpen])

  function handleSignOut(): void {
    setMenuOpen(false)
    onSignOut()
  }

  function handleUpdatePassword(): void {
    setMenuOpen(false)
    router.push(`/${locale}/manage/${userRole}/password`)
  }

  return (
    <header className="sticky top-0 z-10 border-b border-border bg-card shadow-sm">
      <div className="mx-auto flex w-full max-w-4xl items-center justify-between px-4 py-3 sm:px-6">
      {/* Globe logo + welcome heading */}
      <div className="flex min-w-0 flex-1 items-center gap-3">
        <button
          type="button"
          onClick={() => router.push(`/${locale}/manage`)}
          className="shrink-0 transition-opacity hover:opacity-80"
          aria-label="Go to manage home"
        >
          <Image src="/logo-globe.png" alt="ReSEARCH Events" width={56} height={56} />
        </button>

        <div className="flex min-w-0 flex-1 flex-col">
          <span className="truncate text-sm font-semibold text-foreground sm:text-base">
            {t("dashboard.welcomeHeading", { name: userName })}
          </span>
          <span className="truncate text-xs capitalize text-muted-foreground">
            {userRole}
          </span>
        </div>
      </div>

      {/* Avatar + dropdown */}
      <div className="relative" ref={menuRef}>
        <button
          onClick={() => setMenuOpen((v) => !v)}
          className="flex h-9 w-9 cursor-pointer items-center justify-center rounded-full bg-primary text-sm font-semibold text-primary-foreground transition-opacity hover:opacity-90"
          aria-label={t("dashboard.logoutButton")}
          aria-expanded={menuOpen}
          aria-haspopup="menu"
        >
          {avatarInitial}
        </button>

        {menuOpen && (
          <div
            role="menu"
            className="absolute right-0 mt-2 w-48 rounded-md border border-border bg-card py-1 shadow-lg"
          >
            <button
              role="menuitem"
              onClick={handleUpdatePassword}
              className="w-full px-4 py-2 text-left text-sm text-foreground transition-colors hover:bg-muted"
            >
              {t("dashboard.updatePasswordMenuItem")}
            </button>
            <button
              role="menuitem"
              onClick={handleSignOut}
              className="w-full px-4 py-2 text-left text-sm text-foreground transition-colors hover:bg-muted"
            >
              {t("dashboard.logoutButton")}
            </button>
          </div>
        )}
      </div>
      </div>
    </header>
  )
}

// --- Export ---

export { ManageHeader }
export type { ManageHeaderProps }
