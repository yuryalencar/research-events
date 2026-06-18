import type { JSX } from "react"
import Link from "next/link"
import { CheckCircleIcon } from "lucide-react"
import { useLocale, useTranslations } from "next-intl"

import type { EventListItem } from "@/types/api"

// --- Types ---

interface DeadlineManageSuccessProps {
  event: EventListItem
  summary: { added: number; updated: number; cancelled: number }
}

// --- Component ---

function DeadlineManageSuccess({ event, summary }: DeadlineManageSuccessProps): JSX.Element {
  const t = useTranslations("deadlines.manage.success")
  const locale = useLocale()

  return (
    <div className="mx-auto flex max-w-2xl flex-col items-center gap-8 px-4 py-8 text-center">
      <CheckCircleIcon className="size-16 text-green-500" />

      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold">{t("title")}</h1>
      </div>

      <div className="w-full rounded-lg border border-border bg-card p-6 text-left">
        <h2 className="mb-4 font-semibold text-card-foreground">{event.name}</h2>
        <p className="text-sm text-muted-foreground">
          {t("summary", {
            added: summary.added,
            updated: summary.updated,
            cancelled: summary.cancelled,
          })}
        </p>
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

export { DeadlineManageSuccess }
export type { DeadlineManageSuccessProps }
