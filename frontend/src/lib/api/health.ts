import { apiRequest } from "./client"
import type { HealthResult } from "@/types/api"

// --- Public API ---

// getHealth checks liveness and dependency status. Public endpoint, no
// /api/v1 prefix — not part of the versioned API.
async function getHealth(): Promise<HealthResult> {
  return apiRequest<HealthResult>("/health")
}

// --- Export ---

export { getHealth }
