import type { Metadata } from "next"
import type { ReactNode } from "react"
import { NextIntlClientProvider } from "next-intl"
import { getMessages, getTranslations } from "next-intl/server"

import { LanguageSelector } from "@/components/ui/LanguageSelector"
import { Toaster } from "@/components/ui/sonner"

import "../globals.css"

interface Props {
  children: ReactNode
  params: Promise<{ locale: string }>
}

// generateMetadata is required here — without any metadata export, Next.js
// falls back to a streaming metadata Suspense boundary that mismatches
// between server and client render (a known Next.js 16 dev-mode hydration
// bug), even on pages that render correctly otherwise.
export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("app")
  return { title: t("title") }
}

export default async function LocaleLayout({ children, params }: Props): Promise<React.JSX.Element> {
  const { locale } = await params
  const messages = await getMessages()

  return (
    <html lang={locale}>
      <body>
        <NextIntlClientProvider messages={messages}>
          {children}
          <Toaster />
          <LanguageSelector />
        </NextIntlClientProvider>
      </body>
    </html>
  )
}
