import type { MetadataRoute } from "next"

const APP_URL = process.env.NEXT_PUBLIC_APP_URL ?? "https://research-events.vercel.app"

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: "*",
      allow: "/",
      disallow: ["/en/manage", "/pt/manage", "/es/manage", "/de/manage"],
    },
    sitemap: `${APP_URL}/sitemap.xml`,
  }
}
