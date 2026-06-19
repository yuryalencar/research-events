import { describe, it, expect } from "vitest"

import { decodeJwt } from "./decodeJwt"

// --- Helpers ---

// Builds a minimal JWT-shaped string (header.payload.signature) from a plain
// object. The signature segment is a placeholder — decodeJwt never verifies it.
function makeJwt(payload: object): string {
  const encode = (obj: object): string =>
    btoa(JSON.stringify(obj))
      .replace(/\+/g, "-")
      .replace(/\//g, "_")
      .replace(/=/g, "")

  const header = encode({ alg: "HS256", typ: "JWT" })
  const body = encode(payload)
  return `${header}.${body}.fakesignature`
}

// --- Tests ---

describe("decodeJwt", () => {
  describe("happy path", () => {
    it("decodes valid admin JWT and returns all claims", () => {
      const token = makeJwt({
        sub: "1",
        name: "Admin User",
        email: "admin@example.com",
        role: "admin",
        exp: 9999999999,
        iat: 1000000000,
        jti: "test-jti-1",
      })

      const claims = decodeJwt(token)

      expect(claims.sub).toBe("1")
      expect(claims.name).toBe("Admin User")
      expect(claims.email).toBe("admin@example.com")
      expect(claims.role).toBe("admin")
      expect(claims.exp).toBe(9999999999)
      expect(claims.iat).toBe(1000000000)
      expect(claims.jti).toBe("test-jti-1")
    })

    it("decodes valid moderator JWT and returns correct role", () => {
      const token = makeJwt({
        sub: "2",
        name: "Mod User",
        email: "mod@example.com",
        role: "moderator",
        exp: 9999999999,
        iat: 1000000000,
        jti: "test-jti-2",
      })

      const claims = decodeJwt(token)

      expect(claims.role).toBe("moderator")
      expect(claims.name).toBe("Mod User")
    })
  })

  describe("malformed input", () => {
    it("throws when token has fewer than 3 segments", () => {
      // A real JWT has exactly 3 dot-separated segments; anything else is invalid.
      expect(() => decodeJwt("only.two")).toThrow()
      expect(() => decodeJwt("onlyone")).toThrow()
    })

    it("throws when payload segment is not valid base64url JSON", () => {
      // Middle segment decodes to the plain string "not-json" — JSON.parse will throw.
      const notJsonPayload = btoa("not-json").replace(/=/g, "")
      const token = `fakheader.${notJsonPayload}.fakesig`

      expect(() => decodeJwt(token)).toThrow()
    })

    it("throws when token is an empty string", () => {
      expect(() => decodeJwt("")).toThrow()
    })
  })
})
