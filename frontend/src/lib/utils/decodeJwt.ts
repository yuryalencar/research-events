// --- Types ---

interface JwtClaims {
  sub: string
  name: string
  email: string
  role: "admin" | "moderator"
  exp: number
  iat: number
  jti: string
}

// --- Utility ---

// decodeJwt extracts the payload claims from a JWT without verifying the
// signature — trust is fully delegated to the backend. Used client-side to
// read name/role/email from the token returned by the login and refresh
// endpoints.
//
// Throws if the token is structurally invalid (wrong segment count, non-base64
// payload, or non-JSON payload) so callers can treat any throw as an invalid
// session and redirect to /manage.
function decodeJwt(token: string): JwtClaims {
  const segments = token.split(".")
  if (segments.length !== 3) {
    throw new Error("invalid JWT: expected 3 segments")
  }

  // base64url → base64: replace URL-safe chars, restore padding
  const base64 = segments[1].replace(/-/g, "+").replace(/_/g, "/")
  const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), "=")

  const json = atob(padded)
  return JSON.parse(json) as JwtClaims
}

// --- Export ---

export { decodeJwt }
export type { JwtClaims }
