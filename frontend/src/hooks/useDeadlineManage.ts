import { useState, useCallback, useMemo, useRef } from "react"

import { cancelDeadline, supersedeDeadline, addDeadlines } from "@/lib/api/events"
import { ApiError } from "@/lib/api/client"
import type { DeadlineType, DeadlineResponse, EventListItem, SupersedeDeadlineInput } from "@/types/api"

// --- Types ---

type DeadlineMode = "default" | "superseding" | "pendingCancel"

interface EditValues {
  date: string
  time: string
  timezone: string
}

interface DeadlineState {
  mode: DeadlineMode
  editValues: EditValues
}

interface NewDeadlineRow {
  localId: string
  type: DeadlineType | ""
  description: string
  date: string
  time: string
  timezone: string
  isOptional: boolean
}

type DeadlineErrors = Record<string, string>

interface UseDeadlineManageReturn {
  event: EventListItem
  deadlineStates: Record<number, DeadlineState>
  newDeadlines: NewDeadlineRow[]
  contributor: { name: string; email: string }
  errors: DeadlineErrors
  apiError: string | null
  pageState: "editing" | "submitting" | "success"
  hasChanges: boolean
  successSummary: { added: number; updated: number; cancelled: number } | null
  startSupersede: (id: number) => void
  revertSupersede: (id: number) => void
  cancelDeadlineLocal: (id: number) => void
  revertCancel: (id: number) => void
  updateSupersede: (id: number, field: keyof EditValues, value: string) => void
  addNewDeadline: () => void
  removeNewDeadline: (localId: string) => void
  updateNewDeadline: (localId: string, field: keyof Omit<NewDeadlineRow, "localId">, value: unknown) => void
  updateContributor: (field: "name" | "email", value: string) => void
  validate: () => boolean
  submitChanges: () => Promise<void>
}

// --- Pure helpers ---

function toEditValues(d: DeadlineResponse): EditValues {
  return {
    date: d.date,
    time: d.time ?? "",
    timezone: d.timezone ?? "",
  }
}

// supersedeHasChanged compares current edit values against the original
// deadline, treating null API values as "". A supersede that results in
// identical date/time/timezone is not counted as a change.
function supersedeHasChanged(state: DeadlineState, original: DeadlineResponse): boolean {
  return (
    state.editValues.date !== original.date ||
    state.editValues.time !== (original.time ?? "") ||
    state.editValues.timezone !== (original.timezone ?? "")
  )
}

function buildInitialStates(deadlines: DeadlineResponse[]): Record<number, DeadlineState> {
  const result: Record<number, DeadlineState> = {}
  for (const d of deadlines.filter(d => d.is_active)) {
    result[d.id] = { mode: "default", editValues: toEditValues(d) }
  }
  return result
}

// TIME_RE matches HH:MM for hours 00–23 and minutes 00–59.
const TIME_RE = /^([01]\d|2[0-3]):[0-5]\d$/

// computeErrors validates all in-flight changes and returns a flat error map.
// Keys: "contributor.name", "contributor.email",
//       "supersede.{id}.date", "supersede.{id}.time",
//       "new.{localId}.type", "new.{localId}.description",
//       "new.{localId}.date", "new.{localId}.time"
function computeErrors(
  contributor: { name: string; email: string },
  deadlineStates: Record<number, DeadlineState>,
  newDeadlines: NewDeadlineRow[],
  endDate: string,
): DeadlineErrors {
  const errors: DeadlineErrors = {}

  // Contributor
  if (!contributor.name.trim()) {
    errors["contributor.name"] = "Name is required"
  }
  if (!contributor.email.trim()) {
    errors["contributor.email"] = "Email is required"
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(contributor.email)) {
    errors["contributor.email"] = "Enter a valid email address"
  }

  // Superseding deadlines — only cards currently in superseding mode
  for (const [idStr, state] of Object.entries(deadlineStates)) {
    if (state.mode !== "superseding") continue
    const prefix = `supersede.${idStr}`
    if (!state.editValues.date) {
      errors[`${prefix}.date`] = "Date is required"
    } else if (state.editValues.date > endDate) {
      errors[`${prefix}.date`] = "Date must not be after the event end date"
    }
    if (state.editValues.time && !TIME_RE.test(state.editValues.time)) {
      errors[`${prefix}.time`] = "Time must be in HH:MM format (e.g. 14:30)"
    }
  }

  // New deadline cards
  for (const d of newDeadlines) {
    const prefix = `new.${d.localId}`
    if (!d.type) {
      errors[`${prefix}.type`] = "Type is required"
    }
    if (!d.description.trim()) {
      errors[`${prefix}.description`] = "Description is required"
    }
    if (!d.date) {
      errors[`${prefix}.date`] = "Date is required"
    } else if (d.date > endDate) {
      errors[`${prefix}.date`] = "Date must not be after the event end date"
    }
    if (d.time && !TIME_RE.test(d.time)) {
      errors[`${prefix}.time`] = "Time must be in HH:MM format (e.g. 14:30)"
    }
  }

  return errors
}

// --- Hook ---

function useDeadlineManage(event: EventListItem): UseDeadlineManageReturn {
  const originalById = useMemo(() => {
    const map: Record<number, DeadlineResponse> = {}
    for (const d of event.deadlines.filter(d => d.is_active)) {
      map[d.id] = d
    }
    return map
  }, [event.deadlines])

  const [deadlineStates, setDeadlineStates] = useState<Record<number, DeadlineState>>(
    () => buildInitialStates(event.deadlines),
  )
  const [newDeadlines, setNewDeadlines] = useState<NewDeadlineRow[]>([])
  const [contributor, setContributor] = useState({ name: "", email: "" })
  const [pageState, setPageState] = useState<"editing" | "submitting" | "success">("editing")
  const [apiError, setApiError] = useState<string | null>(null)
  const [successSummary, setSuccessSummary] = useState<{
    added: number
    updated: number
    cancelled: number
  } | null>(null)

  // Errors live in a ref so validate() callers can read them synchronously
  // without waiting for a re-render. setErrorVersion triggers the re-render
  // so components update after errors change.
  const errorsRef = useRef<DeadlineErrors>({})
  const [, setErrorVersion] = useState(0)

  const applyErrors = useCallback((next: DeadlineErrors): void => {
    const current = errorsRef.current
    Object.keys(current).forEach(k => delete (current as DeadlineErrors)[k])
    Object.assign(current, next)
    setErrorVersion(v => v + 1)
  }, [])

  const hasChanges = useMemo(() => {
    for (const [idStr, state] of Object.entries(deadlineStates)) {
      const id = Number(idStr)
      if (state.mode === "pendingCancel") return true
      if (state.mode === "superseding") {
        const original = originalById[id]
        if (original && supersedeHasChanged(state, original)) return true
      }
    }
    return newDeadlines.length > 0
  }, [deadlineStates, newDeadlines, originalById])

  // startSupersede copies the original deadline's values into editValues and
  // switches the card to superseding mode, regardless of its previous state.
  // This also clears a pending-cancel if the user clicks the pencil while the
  // card shows the strikethrough state (spec: pencil on pending-cancel reverts
  // the cancel and enters superseding).
  const startSupersede = useCallback(
    (id: number) => {
      const original = originalById[id]
      if (!original) return
      setDeadlineStates(prev => ({
        ...prev,
        [id]: { mode: "superseding", editValues: toEditValues(original) },
      }))
    },
    [originalById],
  )

  const revertSupersede = useCallback(
    (id: number) => {
      const original = originalById[id]
      if (!original) return
      setDeadlineStates(prev => ({
        ...prev,
        [id]: { mode: "default", editValues: toEditValues(original) },
      }))
    },
    [originalById],
  )

  // cancelDeadlineLocal transitions a card to pendingCancel, exiting any
  // superseding state in the process (spec: X on supersede-editing card →
  // exit edit + enter pending-cancel).
  const cancelDeadlineLocal = useCallback((id: number) => {
    setDeadlineStates(prev => ({
      ...prev,
      [id]: { ...prev[id], mode: "pendingCancel" },
    }))
  }, [])

  const revertCancel = useCallback(
    (id: number) => {
      const original = originalById[id]
      if (!original) return
      setDeadlineStates(prev => ({
        ...prev,
        [id]: { mode: "default", editValues: toEditValues(original) },
      }))
    },
    [originalById],
  )

  const updateSupersede = useCallback((id: number, field: keyof EditValues, value: string) => {
    setDeadlineStates(prev => ({
      ...prev,
      [id]: { ...prev[id], editValues: { ...prev[id].editValues, [field]: value } },
    }))
  }, [])

  const addNewDeadline = useCallback(() => {
    const row: NewDeadlineRow = {
      localId: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
      type: "",
      description: "",
      date: "",
      time: "",
      timezone: "",
      isOptional: false,
    }
    setNewDeadlines(prev => [...prev, row])
  }, [])

  const removeNewDeadline = useCallback((localId: string) => {
    setNewDeadlines(prev => prev.filter(d => d.localId !== localId))
  }, [])

  const updateNewDeadline = useCallback(
    (localId: string, field: keyof Omit<NewDeadlineRow, "localId">, value: unknown) => {
      setNewDeadlines(prev => prev.map(d => (d.localId === localId ? { ...d, [field]: value } : d)))
    },
    [],
  )

  const updateContributor = useCallback((field: "name" | "email", value: string) => {
    setContributor(prev => ({ ...prev, [field]: value }))
  }, [])

  const validate = useCallback((): boolean => {
    const next = computeErrors(contributor, deadlineStates, newDeadlines, event.end_date)
    applyErrors(next)
    return Object.keys(next).length === 0
  }, [contributor, deadlineStates, newDeadlines, event.end_date, applyErrors])

  const submitChanges = useCallback(async (): Promise<void> => {
    if (!validate()) return

    setPageState("submitting")
    setApiError(null)

    const submitter = { name: contributor.name, email: contributor.email }
    const promises: Promise<unknown>[] = []
    let cancelled = 0
    let updated = 0

    for (const [idStr, state] of Object.entries(deadlineStates)) {
      const id = Number(idStr)

      if (state.mode === "pendingCancel") {
        promises.push(cancelDeadline(event.id, id, { submitter }))
        cancelled++
        continue
      }

      if (state.mode === "superseding") {
        const original = originalById[id]
        if (!original || !supersedeHasChanged(state, original)) continue
        const input: SupersedeDeadlineInput = {
          submitter,
          date: state.editValues.date,
          ...(state.editValues.time ? { time: state.editValues.time } : {}),
          ...(state.editValues.timezone ? { timezone: state.editValues.timezone } : {}),
        }
        promises.push(supersedeDeadline(event.id, id, input))
        updated++
      }
    }

    const added = newDeadlines.length
    if (newDeadlines.length > 0) {
      const deadlineInputs = newDeadlines.map(d => ({
        type: d.type as DeadlineType,
        description: d.description,
        date: d.date,
        is_optional: d.isOptional,
        ...(d.time ? { time: d.time } : {}),
        ...(d.timezone ? { timezone: d.timezone } : {}),
      }))
      promises.push(addDeadlines(event.id, { submitter, deadlines: deadlineInputs }))
    }

    try {
      await Promise.all(promises)
      setPageState("success")
      setSuccessSummary({ added, updated, cancelled })
    } catch (err) {
      setApiError(err instanceof ApiError ? err.code : "NETWORK_ERROR")
      setPageState("editing")
    }
  }, [validate, contributor, deadlineStates, newDeadlines, event, originalById])

  return {
    event,
    deadlineStates,
    newDeadlines,
    contributor,
    errors: errorsRef.current,
    apiError,
    pageState,
    hasChanges,
    successSummary,
    startSupersede,
    revertSupersede,
    cancelDeadlineLocal,
    revertCancel,
    updateSupersede,
    addNewDeadline,
    removeNewDeadline,
    updateNewDeadline,
    updateContributor,
    validate,
    submitChanges,
  }
}

// --- Export ---

export { useDeadlineManage }
export type { UseDeadlineManageReturn, DeadlineState, DeadlineMode, EditValues, NewDeadlineRow }
