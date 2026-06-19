import { apiRequest, apiPrivateRequest } from "./client"
import type { LoginInput, LoginResult, RefreshTokenResult } from "@/types/api"

// --- Public API ---

// login authenticates an admin/moderator. Public request — the backend sets
// the access_token/refresh_token cookies on the response.
// credentials: "include" is required so the browser accepts and stores the
// HTTP-only cookies from a cross-origin response (dev: :3000 → :8080).
async function login(input: LoginInput): Promise<LoginResult> {
  return apiRequest<LoginResult>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify(input),
    credentials: "include",
  })
}

// refreshToken rotates the access/refresh token pair using the HTTP-only
// refresh_token cookie. Exported for explicit use; apiPrivateRequest also
// calls the same endpoint internally on TOKEN_EXPIRED.
async function refreshToken(): Promise<RefreshTokenResult> {
  return apiPrivateRequest<RefreshTokenResult>("/api/v1/auth/refresh-token", {
    method: "POST",
  })
}

// logout clears the server-side tokens and expires both cookies.
async function logout(): Promise<null> {
  return apiPrivateRequest<null>("/api/v1/auth/logout", {
    method: "POST",
  })
}

// --- Export ---

export { login, refreshToken, logout }
