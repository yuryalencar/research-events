import { describe, it, expect } from "vitest"

import { checkPasswordComplexity } from "./utils"

describe("checkPasswordComplexity", () => {
  it("returns all true for a fully valid password", () => {
    const result = checkPasswordComplexity("Secure@1")
    expect(result).toEqual({
      minLength: true,
      hasUppercase: true,
      hasLowercase: true,
      hasSpecial: true,
    })
  })

  it("returns minLength:false for a short password", () => {
    const result = checkPasswordComplexity("Ab@1")
    expect(result.minLength).toBe(false)
    expect(result.hasUppercase).toBe(true)
    expect(result.hasLowercase).toBe(true)
    expect(result.hasSpecial).toBe(true)
  })

  it("returns hasUppercase:false when no uppercase letter present", () => {
    const result = checkPasswordComplexity("secure@1")
    expect(result.hasUppercase).toBe(false)
    expect(result.minLength).toBe(true)
    expect(result.hasLowercase).toBe(true)
    expect(result.hasSpecial).toBe(true)
  })

  it("returns hasLowercase:false when no lowercase letter present", () => {
    const result = checkPasswordComplexity("SECURE@1")
    expect(result.hasLowercase).toBe(false)
    expect(result.minLength).toBe(true)
    expect(result.hasUppercase).toBe(true)
    expect(result.hasSpecial).toBe(true)
  })

  it("returns hasSpecial:false when no special character present", () => {
    const result = checkPasswordComplexity("Secure12")
    expect(result.hasSpecial).toBe(false)
    expect(result.minLength).toBe(true)
    expect(result.hasUppercase).toBe(true)
    expect(result.hasLowercase).toBe(true)
  })

  it("returns all false for an empty string", () => {
    const result = checkPasswordComplexity("")
    expect(result).toEqual({
      minLength: false,
      hasUppercase: false,
      hasLowercase: false,
      hasSpecial: false,
    })
  })
})
