import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { renderHook, act } from "@testing-library/react"

import { useLoadingMessage } from "./useLoadingMessage"

// --- Fixtures ---

const MESSAGES = ["Message 1", "Message 2", "Message 3", "Message 4"]
const INITIAL_DELAY = 2500
const INTERVAL = 1500

// --- Tests ---

describe("useLoadingMessage", () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  // Spec: "show message 1 immediately on load"
  it("returns the first message immediately when isLoading is true", () => {
    const { result } = renderHook(() =>
      useLoadingMessage(true, { messages: MESSAGES, initialDelay: INITIAL_DELAY, interval: INTERVAL }),
    )

    expect(result.current.message).toBe("Message 1")
  })

  // Spec: "stays on first message before initialDelay elapses"
  it("stays on the first message before initialDelay elapses", () => {
    const { result } = renderHook(() =>
      useLoadingMessage(true, { messages: MESSAGES, initialDelay: INITIAL_DELAY, interval: INTERVAL }),
    )

    act(() => {
      vi.advanceTimersByTime(INITIAL_DELAY - 1)
    })

    expect(result.current.message).toBe("Message 1")
  })

  // Spec: "switches to message 2 after 2500ms of continuous loading"
  it("switches to the second message after initialDelay ms", () => {
    const { result } = renderHook(() =>
      useLoadingMessage(true, { messages: MESSAGES, initialDelay: INITIAL_DELAY, interval: INTERVAL }),
    )

    act(() => {
      vi.advanceTimersByTime(INITIAL_DELAY)
    })

    expect(result.current.message).toBe("Message 2")
  })

  // Spec: "advances +1 every 1000ms until last message"
  it("advances to the third message after initialDelay + interval ms", () => {
    const { result } = renderHook(() =>
      useLoadingMessage(true, { messages: MESSAGES, initialDelay: INITIAL_DELAY, interval: INTERVAL }),
    )

    act(() => {
      vi.advanceTimersByTime(INITIAL_DELAY + INTERVAL)
    })

    expect(result.current.message).toBe("Message 3")
  })

  // Spec: "advances +1 every 1000ms until last message"
  it("advances to the fourth message after initialDelay + 2×interval ms", () => {
    const { result } = renderHook(() =>
      useLoadingMessage(true, { messages: MESSAGES, initialDelay: INITIAL_DELAY, interval: INTERVAL }),
    )

    act(() => {
      vi.advanceTimersByTime(INITIAL_DELAY + INTERVAL * 2)
    })

    expect(result.current.message).toBe("Message 4")
  })

  // Spec: "stays on last message — does not loop"
  it("stays on the last message and does not loop past it", () => {
    const { result } = renderHook(() =>
      useLoadingMessage(true, { messages: MESSAGES, initialDelay: INITIAL_DELAY, interval: INTERVAL }),
    )

    act(() => {
      vi.advanceTimersByTime(INITIAL_DELAY + INTERVAL * 20)
    })

    expect(result.current.message).toBe("Message 4")
  })

  // Spec: "timers reset when isLoading flips true again (new fetch)"
  it("resets to the first message when isLoading flips false then true again", () => {
    const { result, rerender } = renderHook(
      ({ isLoading }: { isLoading: boolean }) =>
        useLoadingMessage(isLoading, { messages: MESSAGES, initialDelay: INITIAL_DELAY, interval: INTERVAL }),
      { initialProps: { isLoading: true } },
    )

    act(() => {
      vi.advanceTimersByTime(INITIAL_DELAY)
    })
    expect(result.current.message).toBe("Message 2")

    act(() => rerender({ isLoading: false }))
    act(() => rerender({ isLoading: true }))

    expect(result.current.message).toBe("Message 1")
  })

  // Spec: "isLoading becomes false before initialDelay → no message switch ever happens"
  it("does not switch messages when isLoading becomes false before initialDelay elapses", () => {
    const { result, rerender } = renderHook(
      ({ isLoading }: { isLoading: boolean }) =>
        useLoadingMessage(isLoading, { messages: MESSAGES, initialDelay: INITIAL_DELAY, interval: INTERVAL }),
      { initialProps: { isLoading: true } },
    )

    act(() => {
      vi.advanceTimersByTime(INITIAL_DELAY - 1)
    })

    act(() => rerender({ isLoading: false }))

    act(() => {
      vi.advanceTimersByTime(INITIAL_DELAY * 2)
    })

    expect(result.current.message).toBe("Message 1")
  })
})
