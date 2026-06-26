import { toast } from "sonner"

import { ApiError } from "./client"

// --- Constants ---

// VALIDATION_ERROR's raw backend message is field-specific English text
// (e.g. "slug must match ^[A-Za-z0-9_-]+$") and must never reach the user —
// it always maps to a generic "check your input" message instead.
// Per-field inline validation is a separate future feature.
const VALIDATION_ERROR_CODE = "VALIDATION_ERROR"
const UNKNOWN_CODE = "UNKNOWN"

// --- Public API ---

// errorMessageKey maps a backend error `code` to an i18n key under the
// `errors` namespace. Unknown/unmapped codes fall back to `errors.UNKNOWN`,
// so a new backend error code never crashes the frontend — it just shows a
// generic message until a translation is added.
function errorMessageKey(code: string): string {
  return `errors.${KNOWN_CODES.has(code) ? code : UNKNOWN_CODE}`
}

// handleApiError shows a toast for any error thrown by apiRequest/
// apiPrivateRequest. ApiErrors are mapped to their specific translated
// message (with VALIDATION_ERROR always generic); anything else (an
// unexpected throw) shows the translated "unknown error" message.
function handleApiError(error: unknown, t: (key: string) => string): void {
  if (error instanceof ApiError) {
    toast.error(t(errorMessageKey(error.code)))
    return
  }

  toast.error(t(errorMessageKey(UNKNOWN_CODE)))
}

// --- Internals ---

// KNOWN_CODES lists every backend error code with a translated
// `errors.<CODE>` key in messages/{en,pt,es,de}.json (see Error cases table
// in specs/frontend/api-client-error-handling.md).
const KNOWN_CODES = new Set([
  VALIDATION_ERROR_CODE,
  "UNAUTHORIZED",
  "TOKEN_MISSING",
  "TOKEN_EXPIRED",
  "TOKEN_INVALID",
  "INVALID_CREDENTIALS",
  "INVALID_CURRENT_PASSWORD",
  "ACCOUNT_LOCKED",
  "REFRESH_TOKEN_MISSING",
  "REFRESH_TOKEN_INVALID",
  "REFRESH_TOKEN_REUSE",
  "FORBIDDEN",
  "CANNOT_REVIEW_OWN_EVENT",
  "CANNOT_UNLOCK_SELF",
  "EVENT_NOT_FOUND",
  "DEADLINE_NOT_FOUND",
  "USER_NOT_FOUND",
  "EVENT_NOT_APPROVED",
  "EVENT_ALREADY_SUBMITTED",
  "DEADLINE_ALREADY_INACTIVE",
  "SLUG_ALREADY_EXISTS",
  "USER_NOT_LOCKED",
  "RATE_LIMIT_EXCEEDED",
  "INTERNAL_ERROR",
  "NETWORK_ERROR",
  "EMAIL_ALREADY_EXISTS",
  "CANNOT_CHANGE_OWN_PASSWORD",
  "PASSWORD_TOO_WEAK",
  "ROLE_UNCHANGED",
  "CANNOT_CHANGE_OWN_ROLE",
])

// --- Export ---

export { errorMessageKey, handleApiError }
