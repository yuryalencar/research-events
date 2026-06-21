import { apiPrivateRequest } from "./client"

// --- Types ---

interface UpdatePasswordInput {
  current_password: string
  new_password: string
  new_password_confirmation: string
}

interface SessionUser {
  id: number
  name: string
  email: string
  role: "admin" | "moderator"
}

// --- Public API ---

async function updatePassword(input: UpdatePasswordInput): Promise<void> {
  await apiPrivateRequest<void>("/api/v1/users/me/password", {
    method: "PATCH",
    body: JSON.stringify(input),
  })
}

// validateSession calls GET /api/v1/users/me to confirm the access token (or
// refresh token) is still valid. Used by useSessionGuard on every protected page
// load — if this throws an auth error the user is redirected to login.
async function validateSession(): Promise<SessionUser> {
  return apiPrivateRequest<SessionUser>("/api/v1/users/me")
}

// --- Export ---

export { updatePassword, validateSession }
export type { UpdatePasswordInput, SessionUser }
