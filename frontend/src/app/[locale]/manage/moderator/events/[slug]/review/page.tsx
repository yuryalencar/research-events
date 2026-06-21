"use client"

import type { JSX } from "react"
import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import { useLocale, useTranslations } from "next-intl"

import { ReviewWizard } from "@/components/manage/review/ReviewWizard"
import type { EventListItem } from "@/types/api"
import { useSessionGuard } from "@/hooks/useSessionGuard"

// --- Component ---

export default function ModeratorReviewPage(): JSX.Element {
  const t = useTranslations("manage.dashboard")
  const locale = useLocale()
  const router = useRouter()

  const { user } = useSessionGuard("moderator")
  const [event, setEvent] = useState<EventListItem | null>(null)

  // Event guard runs only once the auth guard has resolved a valid user.
  useEffect(() => {
    if (!user) return
    try {
      const raw = sessionStorage.getItem("manage_review_event")
      if (!raw) {
        router.replace(`/${locale}/manage/moderator`)
        return
      }
      setEvent(JSON.parse(raw) as EventListItem)
    } catch {
      router.replace(`/${locale}/manage/moderator`)
    }
  }, [user, locale, router])

  if (!user || !event) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-background">
        <p className="text-sm text-muted-foreground">{t("loading")}</p>
      </main>
    )
  }

  return <ReviewWizard event={event} user={user} />
}
