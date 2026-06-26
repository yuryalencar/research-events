import { useEffect, useState } from "react"

// --- Types ---

interface UseLoadingMessageOptions {
  messages: string[]
  initialDelay?: number
  interval?: number
}

interface UseLoadingMessageReturn {
  message: string
}

// --- Hook ---

// useLoadingMessage advances through a list of messages while isLoading is true.
// The first message shows immediately; after initialDelay ms the hook switches to
// the second message and then advances by one every interval ms until the last
// message is reached. All timers are cleared when isLoading becomes false, and
// the index resets to 0 so the next load starts fresh from message 1.
//
// Callers should stabilise the messages array reference (e.g. with useMemo) to
// prevent the effect from restarting on every render.
function useLoadingMessage(
  isLoading: boolean,
  { messages, initialDelay = 3000, interval = 2500 }: UseLoadingMessageOptions,
): UseLoadingMessageReturn {
  const [index, setIndex] = useState(0)

  useEffect(() => {
    if (!isLoading) {
      setIndex(0)
      return
    }

    let intervalId: ReturnType<typeof setInterval> | null = null

    const timeoutId = setTimeout(() => {
      setIndex(1)

      intervalId = setInterval(() => {
        setIndex((prev) => Math.min(prev + 1, messages.length - 1))
      }, interval)
    }, initialDelay)

    return () => {
      clearTimeout(timeoutId)
      if (intervalId !== null) clearInterval(intervalId)
    }
  }, [isLoading, initialDelay, interval, messages.length])

  return { message: messages[index] ?? messages[0] }
}

// --- Export ---

export { useLoadingMessage }
export type { UseLoadingMessageOptions, UseLoadingMessageReturn }
