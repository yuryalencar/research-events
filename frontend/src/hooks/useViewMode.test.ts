import { describe, it, expect, vi, afterEach } from "vitest"
import { renderHook, act } from "@testing-library/react"

import { useViewMode } from "./useViewMode"

// Reuses the same matchMedia mock pattern established in useMediaQuery.test.ts —
// jsdom does not implement matchMedia, so we stub window.matchMedia globally.
function createMatchMediaMock(initialMatches: boolean): {
  matchMedia: (query: string) => MediaQueryList
  triggerChange: (newMatches: boolean) => void
} {
  let currentMatches = initialMatches
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

describe("useViewMode", () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  // --- Initial state ---
  // Spec: "Globe view (default)" — the page always opens showing the globe.

  it("starts in globe mode", () => {
    const { matchMedia } = createMatchMediaMock(true)
    vi.stubGlobal("matchMedia", matchMedia)

    const { result } = renderHook(() => useViewMode(vi.fn()))

    expect(result.current.viewMode).toBe("globe")
  })

  // --- Manual toggle ---
  // Spec: "Clicking the button switches between globe and table"

  it("toggleView switches from globe to table", () => {
    const { matchMedia } = createMatchMediaMock(true)
    vi.stubGlobal("matchMedia", matchMedia)

    const { result } = renderHook(() => useViewMode(vi.fn()))

    act(() => result.current.toggleView())

    expect(result.current.viewMode).toBe("table")
  })

  it("toggleView switches from table back to globe", () => {
    const { matchMedia } = createMatchMediaMock(true)
    vi.stubGlobal("matchMedia", matchMedia)

    const { result } = renderHook(() => useViewMode(vi.fn()))

    act(() => result.current.toggleView())
    act(() => result.current.toggleView())

    expect(result.current.viewMode).toBe("globe")
  })

  // --- Forced switch on resize ---
  // Spec: "When the viewport shrinks below md (768px) while in table mode →
  //        auto-switch to globe + show toast"

  it("forced switch to globe when resize goes below md while in table mode", () => {
    const { matchMedia, triggerChange } = createMatchMediaMock(true)
    vi.stubGlobal("matchMedia", matchMedia)

    const { result } = renderHook(() => useViewMode(vi.fn()))

    // Switch to table first
    act(() => result.current.toggleView())
    expect(result.current.viewMode).toBe("table")

    // Simulate viewport shrinking below md
    act(() => triggerChange(false))

    expect(result.current.viewMode).toBe("globe")
  })

  it("calls onForcedGlobe callback when forced back to globe on resize", () => {
    const { matchMedia, triggerChange } = createMatchMediaMock(true)
    vi.stubGlobal("matchMedia", matchMedia)

    const onForcedGlobe = vi.fn()
    const { result } = renderHook(() => useViewMode(onForcedGlobe))

    act(() => result.current.toggleView())
    act(() => triggerChange(false))

    expect(onForcedGlobe).toHaveBeenCalledTimes(1)
  })

  // --- No spurious forced switch ---
  // Spec: "no forced switch when already in globe mode and resize goes below md"

  it("does not switch or call callback when already in globe mode and resize goes below md", () => {
    const { matchMedia, triggerChange } = createMatchMediaMock(true)
    vi.stubGlobal("matchMedia", matchMedia)

    const onForcedGlobe = vi.fn()
    const { result } = renderHook(() => useViewMode(onForcedGlobe))

    // Still in globe mode (default) — resize should be a no-op
    act(() => triggerChange(false))

    expect(result.current.viewMode).toBe("globe")
    expect(onForcedGlobe).not.toHaveBeenCalled()
  })

  // Spec: "no forced switch when resizing above md"

  it("does not call callback when resizing from below md to above md", () => {
    const { matchMedia, triggerChange } = createMatchMediaMock(false)
    vi.stubGlobal("matchMedia", matchMedia)

    const onForcedGlobe = vi.fn()
    renderHook(() => useViewMode(onForcedGlobe))

    // Resize upward — expanding the viewport should never trigger a forced switch
    act(() => triggerChange(true))

    expect(onForcedGlobe).not.toHaveBeenCalled()
  })
})
