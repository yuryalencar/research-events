import { describe, it, expect, vi } from "vitest"

import { apiRequest } from "./client"
import { getHealth } from "./health"
import type { HealthResult } from "@/types/api"

vi.mock("./client", () => ({
  apiRequest: vi.fn(),
}))

describe("getHealth", () => {
  it("gets /health (no /api/v1 prefix) via apiRequest", async () => {
    const result: HealthResult = {
      status: "healthy",
      version: "1.0.0",
      timestamp: "2026-01-15T10:00:00Z",
      uptime: "3h22m10s",
      checks: { database: { status: "healthy", latency_ms: 4 } },
    }
    vi.mocked(apiRequest).mockResolvedValue(result)

    const data = await getHealth()

    expect(apiRequest).toHaveBeenCalledWith("/health")
    expect(data).toEqual(result)
  })
})
