"use client"

import type { JSX } from "react"
import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import { useLocale, useTranslations } from "next-intl"

import { ReviewWizard } from "@/components/manage/review/ReviewWizard"
import type { SessionUser } from "@/hooks/useReviewWizard"
import type { EventListItem } from "@/types/api"

// --- Component ---

export default function AdminReviewPage(): JSX.Element {
  const t = useTranslations("manage.dashboard")
  const locale = useLocale()
  const router = useRouter()

  const [user, setUser] = useState<SessionUser | null>(null)
  const [event, setEvent] = useState<EventListItem | null>(null)

  useEffect(() => {
    try {
      // Auth guard — must be logged in as admin
      const stored = localStorage.getItem("manage_user")
      if (!stored) {
        router.replace(`/${locale}/manage`)
        return
      }
      const parsed = JSON.parse(stored) as SessionUser
      if (parsed.role !== "admin") {
        router.replace(`/${locale}/manage/${parsed.role}`)
        return
      }

      // Event guard — must have arrived via the dashboard Review button
      const raw = sessionStorage.getItem("manage_review_event")
      if (!raw) {
        router.replace(`/${locale}/manage/admin`)
        return
      }

      setUser(parsed)
      setEvent(JSON.parse(raw) as EventListItem)
    } catch {
      router.replace(`/${locale}/manage`)
    }
  }, [locale, router])

  if (!user || !event) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-background">
        <p className="text-sm text-muted-foreground">{t("loading")}</p>
      </main>
    )
  }

  return <ReviewWizard event={event} user={user} />
}
