"use client"

import type { JSX } from "react"
import { useRouter } from "next/navigation"
import { useTranslations } from "next-intl"

import { isOwnEvent } from "@/hooks/useReviewEvents"
import type { EventListItem } from "@/types/api"

// --- Types ---

interface EventReviewCardProps {
  event: EventListItem
  role: "admin" | "moderator"
  userId: number
  locale: string
}

// --- Component ---

function EventReviewCard({ event, role, userId, locale }: EventReviewCardProps): JSX.Element {
  const t = useTranslations("manage.reviewDashboard")
  const router = useRouter()

  // Moderators cannot review events they submitted themselves.
  // Admins always see the normal card variant.
  const blocked = role === "moderator" && isOwnEvent(event, userId)

  const reviewHref = `/${locale}/manage/${role}/events/${event.slug}/review`

  function handleReview(): void {
    sessionStorage.setItem("manage_review_event", JSON.stringify(event))
    router.push(reviewHref)
  }

  if (blocked) {
    return (
      <div className="flex flex-col gap-2 rounded-lg border border-border bg-card p-4 opacity-50 sm:flex-row sm:items-center sm:justify-between">
        <span className="text-sm text-muted-foreground">{event.name}</span>
        <span className="text-sm text-muted-foreground">{t("cannotReview")}</span>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-2 rounded-lg border border-border bg-card p-4 sm:flex-row sm:items-center sm:justify-between">
      <a
        href={event.website_url}
        target="_blank"
        rel="noopener noreferrer"
        className="text-sm font-medium text-foreground underline-offset-4 hover:underline"
      >
        {event.name}
      </a>
      <button
        onClick={handleReview}
        className="inline-flex items-center justify-center rounded-md bg-primary px-4 py-1.5 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
      >
        {t("reviewButton")}
      </button>
    </div>
  )
}

// --- Export ---

export { EventReviewCard }
export type { EventReviewCardProps }
