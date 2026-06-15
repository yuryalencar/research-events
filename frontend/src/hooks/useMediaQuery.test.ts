import { describe, it, expect, vi, afterEach } from "vitest"
import { renderHook, act } from "@testing-library/react"

import { useMediaQuery } from "./useMediaQuery"

// Minimal MediaQueryList mock — jsdom does not implement matchMedia.
function createMatchMediaMock(matches: boolean): {
  matchMedia: (query: string) => MediaQueryList
  triggerChange: (newMatches: boolean) => void
} {
  let currentMatches = matches
  let changeListener: ((event: MediaQueryListEvent) => void) | null = null

  const mql = {
    get matches() {
      return currentMatches
    },
    media: "",
    onchange: null,
    addEventListener: vi.fn((event: string, listener: (event: MediaQueryListEvent) => void) => {
      if (event === "change") changeListener = listener
    }),
    removeEventListener: vi.fn((event: string, listener: (event: MediaQueryListEvent) => void) => {
      if (event === "change" && changeListener === listener) changeListener = null
    }),
    dispatchEvent: vi.fn(),
    addListener: vi.fn(),
    removeListener: vi.fn(),
  } as unknown as MediaQueryList

  return {
    matchMedia: () => mql,
    triggerChange: (newMatches: boolean) => {
      currentMatches = newMatches
      changeListener?.({ matches: newMatches } as MediaQueryListEvent)
    },
  }
}

describe("useMediaQuery", () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it("returns true when the query matches on mount", () => {
    const { matchMedia } = createMatchMediaMock(true)
    vi.stubGlobal("matchMedia", matchMedia)

    const { result } = renderHook(() => useMediaQuery("(min-width: 768px)"))

    expect(result.current).toBe(true)
  })

  it("returns false when the query does not match on mount", () => {
    const { matchMedia } = createMatchMediaMock(false)
    vi.stubGlobal("matchMedia", matchMedia)

    const { result } = renderHook(() => useMediaQuery("(min-width: 768px)"))

    expect(result.current).toBe(false)
  })

  it("updates when the media query match state changes", () => {
    const { matchMedia, triggerChange } = createMatchMediaMock(false)
    vi.stubGlobal("matchMedia", matchMedia)

    const { result } = renderHook(() => useMediaQuery("(min-width: 768px)"))

    expect(result.current).toBe(false)

    act(() => triggerChange(true))

    expect(result.current).toBe(true)
  })
})
