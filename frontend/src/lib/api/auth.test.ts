import { describe, it, expect, vi } from "vitest"

import { apiRequest, apiPrivateRequest } from "./client"
import { login, refreshToken, logout } from "./auth"
import type { LoginInput, AuthResult } from "@/types/api"

vi.mock("./client", () => ({
  apiRequest: vi.fn(),
  apiPrivateRequest: vi.fn(),
}))

const authResult: AuthResult = {
  token: "jwt-token",
  role: "admin",
  user: { id: 1, name: "Admin", email: "admin@example.com" },
}

describe("login", () => {
  it("posts credentials to /api/v1/auth/login via apiRequest", async () => {
    const input: LoginInput = { email: "admin@example.com", password: "secret" }
    vi.mocked(apiRequest).mockResolvedValue(authResult)

    const result = await login(input)

    expect(apiRequest).toHaveBeenCalledWith("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(input),
    })
    expect(result).toEqual(authResult)
  })
})

describe("refreshToken", () => {
  it("posts to /api/v1/auth/refresh-token via apiPrivateRequest", async () => {
    vi.mocked(apiPrivateRequest).mockResolvedValue(authResult)

    const result = await refreshToken()

    expect(apiPrivateRequest).toHaveBeenCalledWith("/api/v1/auth/refresh-token", {
      method: "POST",
    })
    expect(result).toEqual(authResult)
  })
})

describe("logout", () => {
  it("posts to /api/v1/auth/logout via apiPrivateRequest", async () => {
    vi.mocked(apiPrivateRequest).mockResolvedValue(null)

    const result = await logout()

    expect(apiPrivateRequest).toHaveBeenCalledWith("/api/v1/auth/logout", {
      method: "POST",
    })
    expect(result).toBeNull()
  })
})
