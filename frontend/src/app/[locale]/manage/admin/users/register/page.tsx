"use client"

import type { JSX } from "react"
import { useTranslations, useLocale } from "next-intl"
import { useRouter } from "next/navigation"

import { useSessionGuard } from "@/hooks/useSessionGuard"
import { RegisterUserForm } from "@/components/manage/users/RegisterUserForm"

// --- Component ---

export default function AdminRegisterUserPage(): JSX.Element {
  const t = useTranslations("manage")
  const locale = useLocale()
  const router = useRouter()
  const { user } = useSessionGuard("admin")

  if (!user) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-background">
        <p className="text-sm text-muted-foreground">{t("dashboard.loading")}</p>
      </main>
    )
  }

  function handleAuthError(): void {
    localStorage.removeItem("manage_user")
    router.replace(`/${locale}/manage`)
  }

  return (
    <main className="min-h-screen bg-background">
      <RegisterUserForm locale={locale} onAuthError={handleAuthError} />
    </main>
  )
}
