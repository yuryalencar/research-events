import path from "path"

import { loadEnv } from "vite"
import { defineConfig } from "vitest/config"

export default defineConfig({
  resolve: {
    alias: {
      // Mirrors tsconfig.json's "@/*" -> "./src/*" path mapping, so tests
      // can use the same import aliases as application code.
      "@": path.resolve(__dirname, "./src"),
    },
  },
  test: {
    environment: "jsdom",
    // Load .env (e.g. NEXT_PUBLIC_API_URL) into process.env for tests,
    // mirroring how Next.js inlines these vars at build time.
    env: loadEnv("test", process.cwd(), ""),
  },
})
