import { loadEnv } from "vite"
import { defineConfig } from "vitest/config"

export default defineConfig({
  test: {
    environment: "jsdom",
    // Load .env (e.g. NEXT_PUBLIC_API_URL) into process.env for tests,
    // mirroring how Next.js inlines these vars at build time.
    env: loadEnv("test", process.cwd(), ""),
  },
})
