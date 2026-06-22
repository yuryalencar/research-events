import type { NextConfig } from "next"
import createNextIntlPlugin from "next-intl/plugin"

const withNextIntl = createNextIntlPlugin()

const nextConfig: NextConfig = {
  async rewrites() {
    const backendUrl = process.env.BACKEND_URL
    if (!backendUrl) return []
    return [
      { source: "/api/:path*", destination: `${backendUrl}/api/:path*` },
      { source: "/health",     destination: `${backendUrl}/health` },
    ]
  },
}

export default withNextIntl(nextConfig)
