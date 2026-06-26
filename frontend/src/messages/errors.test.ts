import { describe, it, expect } from "vitest"

import en from "./en.json"
import pt from "./pt.json"
import es from "./es.json"
import de from "./de.json"

// Every backend error code from specs/frontend/api-client-error-handling.md's
// "Error cases" table, plus the UNKNOWN fallback.
const EXPECTED_KEYS = [
  "VALIDATION_ERROR",
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
  "UNKNOWN",
  "EMAIL_ALREADY_EXISTS",
  "CANNOT_CHANGE_OWN_PASSWORD",
  "PASSWORD_TOO_WEAK",
  "ROLE_UNCHANGED",
  "CANNOT_CHANGE_OWN_ROLE",
].sort()

describe("errors namespace i18n key parity", () => {
  it("en.json (source of truth) has every expected errors.* key", () => {
    expect(Object.keys(en.errors ?? {}).sort()).toEqual(EXPECTED_KEYS)
  })

  it.each([
    ["pt", pt],
    ["es", es],
    ["de", de],
  ])("%s.json mirrors every errors.* key from en.json", (_locale, messages) => {
    const keys = Object.keys((messages as typeof en).errors ?? {}).sort()
    expect(keys).toEqual(EXPECTED_KEYS)
  })
})
