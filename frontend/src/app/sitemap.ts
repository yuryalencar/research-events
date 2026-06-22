import type { MetadataRoute } from "next"

const APP_URL = process.env.NEXT_PUBLIC_APP_URL ?? "https://research-events.vercel.app"
const locales = ["en", "pt", "es", "de"] as const

export default function sitemap(): MetadataRoute.Sitemap {
  return locales.map(locale => ({
    url: `${APP_URL}/${locale}`,
    lastModified: new Date(),
    changeFrequency: "daily" as const,
    priority: 1,
  }))
}
