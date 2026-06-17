import type { JSX } from "react"
import Link from "next/link"
import { CheckCircleIcon } from "lucide-react"
import { useLocale, useTranslations } from "next-intl"

import { Badge } from "@/components/ui/badge"
import { formatDateRange } from "@/lib/events"
import type { SubmitEventResult } from "@/types/api"

// --- Types ---

interface SubmitSuccessProps {
  event: SubmitEventResult
}

// --- Component ---

function SubmitSuccess({ event }: SubmitSuccessProps): JSX.Element {
  const t = useTranslations("submit.success")
  const locale = useLocale()

  return (
    <div className="flex flex-col items-center gap-8 py-12 text-center">
      <CheckCircleIcon className="size-16 text-green-500" />

      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold">{t("title")}</h1>
        <p className="max-w-md text-muted-foreground">{t("description")}</p>
      </div>

      <div className="w-full max-w-md rounded-lg border border-border bg-card p-6 text-left">
        <h2 className="mb-4 font-semibold text-card-foreground">{event.name}</h2>
        <dl className="flex flex-col gap-3 text-sm">
          <div className="flex justify-between gap-2">
            <dt className="text-muted-foreground">{t("slug")}</dt>
            <dd className="font-mono">{event.slug}</dd>
          </div>
          <div className="flex justify-between gap-2">
            <dt className="text-muted-foreground">{t("dates")}</dt>
            <dd>{formatDateRange(event.start_date, event.end_date, locale)}</dd>
          </div>
          <div className="flex justify-between gap-2">
            <dt className="text-muted-foreground">{t("website")}</dt>
            <dd>
              <a
                href={event.website_url}
                target="_blank"
                rel="noreferrer"
                className="text-primary underline underline-offset-2"
              >
                {event.website_url}
              </a>
            </dd>
          </div>
          <div className="flex justify-between gap-2">
            <dt className="text-muted-foreground">{t("statusLabel")}</dt>
            <dd>
              <Badge variant="secondary">{t("statusValue")}</Badge>
            </dd>
          </div>
          <div className="flex justify-between gap-2">
            <dt className="text-muted-foreground">{t("submittedBy")}</dt>
            <dd>{event.created_by.name}</dd>
          </div>
        </dl>
      </div>

      <Link
        href={`/${locale}`}
        className="rounded-md bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
      >
        {t("goHome")}
      </Link>
    </div>
  )
}

// --- Export ---

export { SubmitSuccess }
export type { SubmitSuccessProps }
