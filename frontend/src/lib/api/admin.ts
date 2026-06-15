import { apiPrivateRequest } from "./client"
import type { ReviewEventInput, ReviewEventResult, UnlockUserResult } from "@/types/api"

// --- Public API ---

// reviewEvent approves or rejects a pending event, optionally editing its
// fields in the same request. Admin/moderator only.
async function reviewEvent(id: number, input: ReviewEventInput): Promise<ReviewEventResult> {
  return apiPrivateRequest<ReviewEventResult>(`/api/v1/admin/events/${id}/review`, {
    method: "PATCH",
    body: JSON.stringify(input),
  })
}

// unlockUser clears a locked account's failed-login state. Admin/moderator
// only; a user cannot unlock themselves (enforced by the backend).
async function unlockUser(id: number): Promise<UnlockUserResult> {
  return apiPrivateRequest<UnlockUserResult>(`/api/v1/admin/users/${id}/unlock`, {
    method: "PATCH",
  })
}

// --- Export ---

export { reviewEvent, unlockUser }
