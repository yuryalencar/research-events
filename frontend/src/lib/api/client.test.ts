import { describe, it, expect, beforeEach, vi, afterEach } from "vitest"

import { apiRequest, apiRequestWithMeta, apiPrivateRequest, ApiError } from "./client"

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  })
}

describe("apiRequest", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn())
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it("returns data on a successful response", async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({ code: "OK", data: { id: 1, name: "Conf" } }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    )

    const result = await apiRequest<{ id: number; name: string }>("/api/v1/events/1")

    expect(result).toEqual({ id: 1, name: "Conf" })
  })

  it("throws ApiError with code and message on an error envelope", async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({ code: "EVENT_NOT_FOUND", error: { message: "event not found" } }), {
        status: 404,
        headers: { "Content-Type": "application/json" },
      }),
    )

    await expect(apiRequest("/api/v1/events/999")).rejects.toMatchObject({
      code: "EVENT_NOT_FOUND",
      status: 404,
      message: "event not found",
    })
    await expect(apiRequest("/api/v1/events/999")).rejects.toBeInstanceOf(ApiError)
  })

  it("throws an INTERNAL_ERROR ApiError when the response body is not valid JSON", async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response("<html>502 Bad Gateway</html>", {
        status: 502,
        headers: { "Content-Type": "text/html" },
      }),
    )

    await expect(apiRequest("/api/v1/events")).rejects.toMatchObject({
      code: "INTERNAL_ERROR",
      status: 502,
    })
  })

  it("throws a NETWORK_ERROR ApiError when fetch itself rejects", async () => {
    vi.mocked(fetch).mockRejectedValue(new TypeError("Failed to fetch"))

    await expect(apiRequest("/api/v1/events")).rejects.toMatchObject({
      code: "NETWORK_ERROR",
      status: 0,
    })
  })

  it("never sends credentials", async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({ code: "OK", data: null }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    )

    await apiRequest("/api/v1/events")

    const [, init] = vi.mocked(fetch).mock.calls[0]
    expect(init?.credentials).toBeUndefined()
  })

  it("prepends NEXT_PUBLIC_API_URL to the given path", async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({ code: "OK", data: null }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    )

    await apiRequest("/api/v1/events")

    const [url] = vi.mocked(fetch).mock.calls[0]
    expect(url).toBe("http://localhost:8080/api/v1/events")
  })
})

describe("apiRequestWithMeta", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn())
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it("returns data and meta for list endpoints", async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(
        JSON.stringify({ code: "OK", data: [{ id: 1 }], meta: { page: 1, total: 1 } }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
    )

    const result = await apiRequestWithMeta<{ id: number }[]>("/api/v1/events")

    expect(result).toEqual({ data: [{ id: 1 }], meta: { page: 1, total: 1 } })
  })
})

describe("apiPrivateRequest", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn())
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it("sends credentials: include", async () => {
    vi.mocked(fetch).mockResolvedValue(jsonResponse(200, { code: "OK", data: { ok: true } }))

    await apiPrivateRequest("/api/v1/admin/events?status=pending")

    const [, init] = vi.mocked(fetch).mock.calls[0]
    expect(init?.credentials).toBe("include")
  })

  it("on 401 TOKEN_EXPIRED, refreshes once and retries the original request, returning the retry's result", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        jsonResponse(401, { code: "TOKEN_EXPIRED", error: { message: "token expired" } }),
      )
      .mockResolvedValueOnce(
        jsonResponse(200, { code: "REFRESH_SUCCESS", data: { token: "new", role: "admin", user: { id: 1, name: "A", email: "a@b.com" } } }),
      )
      .mockResolvedValueOnce(jsonResponse(200, { code: "OK", data: { id: 1, status: "approved" } }))

    const result = await apiPrivateRequest<{ id: number; status: string }>("/api/v1/admin/events/1/review")

    expect(result).toEqual({ id: 1, status: "approved" })
    expect(fetch).toHaveBeenCalledTimes(3)

    const [refreshUrl, refreshInit] = vi.mocked(fetch).mock.calls[1]
    expect(refreshUrl).toBe("http://localhost:8080/api/v1/auth/refresh-token")
    expect(refreshInit?.credentials).toBe("include")
  })

  it("on 401 TOKEN_EXPIRED, if the refresh call itself fails, throws the original TOKEN_EXPIRED error", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        jsonResponse(401, { code: "TOKEN_EXPIRED", error: { message: "token expired" } }),
      )
      .mockResolvedValueOnce(
        jsonResponse(401, { code: "REFRESH_TOKEN_REUSE", error: { message: "refresh token has already been used" } }),
      )

    await expect(apiPrivateRequest("/api/v1/admin/events/1/review")).rejects.toMatchObject({
      code: "TOKEN_EXPIRED",
      status: 401,
    })
    expect(fetch).toHaveBeenCalledTimes(2)
  })

  it("on 401 TOKEN_EXPIRED, if refresh succeeds but the retried request also fails, throws the retry's error", async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(
        jsonResponse(401, { code: "TOKEN_EXPIRED", error: { message: "token expired" } }),
      )
      .mockResolvedValueOnce(
        jsonResponse(200, { code: "REFRESH_SUCCESS", data: { token: "new", role: "admin", user: { id: 1, name: "A", email: "a@b.com" } } }),
      )
      .mockResolvedValueOnce(
        jsonResponse(403, { code: "FORBIDDEN", error: { message: "insufficient permissions" } }),
      )

    await expect(apiPrivateRequest("/api/v1/admin/events/1/review")).rejects.toMatchObject({
      code: "FORBIDDEN",
      status: 403,
    })
    expect(fetch).toHaveBeenCalledTimes(3)
  })

  it.each(["TOKEN_MISSING", "TOKEN_INVALID"])(
    "on 401 %s, does not attempt a refresh",
    async (code) => {
      vi.mocked(fetch).mockResolvedValue(jsonResponse(401, { code, error: { message: "auth error" } }))

      await expect(apiPrivateRequest("/api/v1/admin/events?status=pending")).rejects.toMatchObject({
        code,
        status: 401,
      })
      expect(fetch).toHaveBeenCalledTimes(1)
    },
  )
})
