import type { Metadata } from "next"
import type { JSX } from "react"
import { getTranslations } from "next-intl/server"

import { SubmitWizard } from "@/components/events/submit/SubmitWizard"

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("meta")
  return {
    title: t("submitTitle"),
    description: t("submitDescription"),
    robots: { index: false, follow: false },
  }
}

export default function SubmitEventPage(): JSX.Element {
  return <SubmitWizard />
}
