import type { ApiMeta } from "@/types/api"

// --- Types ---

// ApiSuccessEnvelope is the shape of every successful backend response:
// { "code": "SOME_CODE", "data": T }, with an optional "meta" for list endpoints.
interface ApiSuccessEnvelope<T> {
  code: string
  data: T
  meta?: ApiMeta
}

// ApiErrorEnvelope is the shape of every error response:
// { "code": "SOME_ERROR_CODE", "error": { "message": "..." } }
interface ApiErrorEnvelope {
  code: string
  error: {
    message: string
  }
}

// --- ApiError ---

// ApiError is thrown by apiRequest/apiPrivateRequest for any non-2xx response,
// network failure, or unparsable body. `code` drives the user-facing message
// lookup in lib/api/errors.ts; `message` is the raw backend message (English),
// useful for debugging but not shown to users directly.
class ApiError extends Error {
  readonly code: string
  readonly status: number

  constructor(code: string, status: number, message: string) {
    super(message)
    this.name = "ApiError"
    this.code = code
    this.status = status
  }
}

// --- Internals ---

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? ""

// sendRequest performs the fetch, parses the envelope, and throws ApiError
// for any failure mode (network, non-JSON body, error envelope). Both
// apiRequest and apiPrivateRequest build on this.
async function sendRequest<T>(path: string, init?: RequestInit): Promise<ApiSuccessEnvelope<T>> {
  let response: Response

  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...init,
      headers: {
        ...(init?.body ? { "Content-Type": "application/json" } : {}),
        ...init?.headers,
      },
    })
  } catch {
    throw new ApiError("NETWORK_ERROR", 0, "could not connect to the server")
  }

  let body: unknown

  try {
    body = await response.json()
  } catch {
    throw new ApiError("NETWORK_ERROR", 0, "received a non-JSON response from the server")
  }

  if (!response.ok) {
    const errorBody = body as ApiErrorEnvelope
    throw new ApiError(errorBody.code, response.status, errorBody.error.message)
  }

  return body as ApiSuccessEnvelope<T>
}

// refreshAccessToken calls POST /api/v1/auth/refresh-token with the
// HTTP-only refresh_token cookie. Resolves to true if the backend issued new
// tokens, false if the refresh itself failed (any error code/status).
async function refreshAccessToken(): Promise<boolean> {
  try {
    await sendRequest("/api/v1/auth/refresh-token", {
      method: "POST",
      credentials: "include",
    })
    return true
  } catch {
    return false
  }
}

// --- Public API ---

// apiRequest performs a public (unauthenticated) request and returns the
// envelope's "data" field. Never sends cookies.
async function apiRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const envelope = await sendRequest<T>(path, init)
  return envelope.data
}

// apiRequestWithMeta is for list endpoints that also return "meta"
// (pagination). Used by listEvents.
async function apiRequestWithMeta<T>(path: string, init?: RequestInit): Promise<{ data: T; meta: ApiMeta }> {
  const envelope = await sendRequest<T>(path, init)
  if (!envelope.meta) {
    throw new ApiError("NETWORK_ERROR", 0, "expected a list response with pagination meta")
  }
  return { data: envelope.data, meta: envelope.meta }
}

// apiPrivateRequest performs an authenticated request, sending the
// HTTP-only access_token/refresh_token cookies via credentials: "include".
//
// If the response is 401 TOKEN_EXPIRED, it calls refreshAccessToken() once:
//   - refresh succeeds  -> retries the original request once and returns/throws
//     whatever that retry produces (even if it also fails).
//   - refresh fails      -> throws the original 401 TOKEN_EXPIRED ApiError.
//
// Any other status (including TOKEN_MISSING/TOKEN_INVALID) is thrown
// immediately — no refresh is attempted.
async function apiPrivateRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const privateInit: RequestInit = { ...init, credentials: "include" }

  try {
    const envelope = await sendRequest<T>(path, privateInit)
    return envelope.data
  } catch (err) {
    if (!(err instanceof ApiError) || err.code !== "TOKEN_EXPIRED") {
      throw err
    }

    const refreshed = await refreshAccessToken()
    if (!refreshed) {
      throw err
    }

    const retryEnvelope = await sendRequest<T>(path, privateInit)
    return retryEnvelope.data
  }
}

// --- Export ---

export { apiRequest, apiRequestWithMeta, apiPrivateRequest, ApiError }
export type { ApiSuccessEnvelope, ApiErrorEnvelope }
