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

const APP_URL = process.env.NEXT_PUBLIC_APP_URL ?? "https://research-events.vercel.app"

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: "app" })

  return {
    metadataBase: new URL(APP_URL),
    title: {
      default: t("title"),
      template: `%s | ${t("title")}`,
    },
    description: t("description"),
    keywords: ["research conferences", "software engineering", "computer science", "academic events", "submission deadlines", "call for papers"],
    authors: [{ name: "Yury Lima" }],
    creator: "Yury Lima",
    openGraph: {
      type: "website",
      siteName: t("title"),
      title: t("title"),
      description: t("description"),
      url: `${APP_URL}/${locale}`,
      locale,
      images: [{ url: "/logo-with-opensource.png", alt: t("title") }],
    },
    twitter: {
      card: "summary_large_image",
      title: t("title"),
      description: t("description"),
      images: ["/logo-with-opensource.png"],
    },
    alternates: {
      canonical: `${APP_URL}/${locale}`,
      languages: {
        en: `${APP_URL}/en`,
        pt: `${APP_URL}/pt`,
        es: `${APP_URL}/es`,
        de: `${APP_URL}/de`,
      },
    },
  }
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
