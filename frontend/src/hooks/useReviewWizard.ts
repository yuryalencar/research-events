import { useState, useCallback } from "react"

import { reviewEvent } from "@/lib/api/admin"
import type {
  EventListItem,
  EventStatus,
  EventTier,
  ReviewEventInput,
  ReviewEventEditInput,
} from "@/types/api"

// --- Types ---

type ReviewStep = 1 | 2 | 3

interface ReviewFormData {
  name: string
  country: string
  city: string
  startDate: string
  endDate: string
  websiteUrl: string
  domain: string
  tier: EventTier | ""
  latitude: number
  longitude: number
}

interface ReviewSuccessState {
  event: EventListItem
  action: "approve" | "reject"
  reason: string | undefined
}

interface SessionUser {
  id: number
  name: string
  role: "admin" | "moderator"
  email: string
}

interface UseReviewWizardReturn {
  step: ReviewStep
  formData: ReviewFormData
  errors: Record<string, string>
  status: EventStatus
  isSubmitting: boolean
  bannerError: string | null
  successState: ReviewSuccessState | null
  goToStep: (s: ReviewStep) => void
  updateField: (field: keyof ReviewFormData, value: string | number) => void
  validateStep1: () => boolean
  approve: (note: string) => Promise<void>
  reject: (reason: string) => Promise<void>
  clearSuccess: () => void
}

// --- Pure helpers ---

// FP: pure function
// buildEventPatch computes only the fields that differ between the original event
// and the reviewer's edited form data. Sending only changed fields keeps the PATCH
// body minimal and prevents overwriting fields the reviewer didn't intend to touch.
// Pure: depends only on its arguments; no I/O or side effects.
function buildEventPatch(
  original: EventListItem,
  formData: ReviewFormData,
): ReviewEventEditInput | undefined {
  const patch: ReviewEventEditInput = {}

  if (formData.name !== original.name) patch.name = formData.name
  if (formData.country !== original.country) patch.country = formData.country
  if (formData.city !== original.city) patch.city = formData.city
  if (formData.latitude !== original.latitude) patch.latitude = formData.latitude
  if (formData.longitude !== original.longitude) patch.longitude = formData.longitude
  if (formData.startDate !== original.start_date) patch.start_date = formData.startDate
  if (formData.endDate !== original.end_date) patch.end_date = formData.endDate
  if (formData.websiteUrl !== original.website_url) patch.website_url = formData.websiteUrl
  if (formData.domain !== original.domain) patch.domain = formData.domain

  // tier "" means "unset / no selection" — omit from patch rather than clearing it
  if (formData.tier !== "") {
    const originalTier = original.tier ?? ""
    if (formData.tier !== originalTier) patch.tier = formData.tier as EventTier
  }

  return Object.keys(patch).length === 0 ? undefined : patch
}

// FP: pure function
// validateFields checks all required Step 1 form fields and returns an error map.
// An empty map means the form is valid. Pure: same input → same output, no I/O.
function validateFields(formData: ReviewFormData): Record<string, string> {
  const errors: Record<string, string> = {}
  if (!formData.name.trim()) errors.name = "required"
  if (!formData.country.trim()) errors.country = "required"
  if (!formData.city.trim()) errors.city = "required"
  if (!formData.startDate.trim()) errors.startDate = "required"
  if (!formData.endDate.trim()) errors.endDate = "required"
  if (!formData.websiteUrl.trim()) errors.websiteUrl = "required"
  return errors
}

// FP: pure function
// initFormData converts an EventListItem into the flat ReviewFormData shape used
// by the form. Pure: derives new state from its input without mutating anything.
function initFormData(event: EventListItem): ReviewFormData {
  return {
    name: event.name,
    country: event.country,
    city: event.city,
    startDate: event.start_date,
    endDate: event.end_date,
    websiteUrl: event.website_url,
    domain: event.domain,
    tier: event.tier ?? "",
    latitude: event.latitude,
    longitude: event.longitude,
  }
}

// --- Hook ---

// useReviewWizard manages the 3-step review wizard: form state, step navigation,
// approve/reject API calls, and the success screen state after a decision.
// Navigation after a decision is handled by the component via successState.
function useReviewWizard(event: EventListItem, _user: SessionUser): UseReviewWizardReturn {
  // 1. State
  const [step, setStep] = useState<ReviewStep>(1)
  const [status, setStatus] = useState<EventStatus>(event.status)
  const [formData, setFormData] = useState<ReviewFormData>(() => initFormData(event))
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [bannerError, setBannerError] = useState<string | null>(null)
  const [successState, setSuccessState] = useState<ReviewSuccessState | null>(null)

  // 2. Handlers

  const goToStep = useCallback((s: ReviewStep): void => {
    setStep(s)
  }, [])

  const updateField = useCallback(
    (field: keyof ReviewFormData, value: string | number): void => {
      setFormData((prev) => ({ ...prev, [field]: value }))
      // Clear the validation error for the field the reviewer just edited so the
      // error message disappears as soon as they start correcting the value.
      setErrors((prev) => {
        if (!(field in prev)) return prev
        const next = { ...prev }
        delete next[field]
        return next
      })
    },
    [],
  )

  const validateStep1 = useCallback((): boolean => {
    const errs = validateFields(formData)
    setErrors(errs)
    return Object.keys(errs).length === 0
  }, [formData])

  const approve = useCallback(
    async (note: string): Promise<void> => {
      setIsSubmitting(true)
      setBannerError(null)
      try {
        const patch = buildEventPatch(event, formData)
        const trimmedNote = note.trim()
        const input: ReviewEventInput = {
          action: "approve",
          ...(trimmedNote ? { reason: trimmedNote } : {}),
          ...(patch ? { event: patch } : {}),
        }
        const updated = await reviewEvent(event.id, input)
        setStatus("approved")
        setSuccessState({
          event: updated,
          action: "approve",
          reason: trimmedNote || undefined,
        })
      } catch (err) {
        setBannerError(err instanceof Error ? err.message : "Unknown error")
      } finally {
        setIsSubmitting(false)
      }
    },
    [event, formData],
  )

  const reject = useCallback(
    async (reason: string): Promise<void> => {
      setIsSubmitting(true)
      setBannerError(null)
      try {
        const patch = buildEventPatch(event, formData)
        const input: ReviewEventInput = {
          action: "reject",
          reason: reason.trim(),
          ...(patch ? { event: patch } : {}),
        }
        const updated = await reviewEvent(event.id, input)
        setSuccessState({
          event: updated,
          action: "reject",
          reason: reason.trim(),
        })
      } catch (err) {
        setBannerError(err instanceof Error ? err.message : "Unknown error")
      } finally {
        setIsSubmitting(false)
      }
    },
    [event, formData],
  )

  const clearSuccess = useCallback((): void => {
    setSuccessState(null)
  }, [])

  // 3. Return
  return {
    step,
    formData,
    errors,
    status,
    isSubmitting,
    bannerError,
    successState,
    goToStep,
    updateField,
    validateStep1,
    approve,
    reject,
    clearSuccess,
  }
}

// --- Export ---

export { useReviewWizard, buildEventPatch }
export type {
  ReviewStep,
  ReviewFormData,
  ReviewSuccessState,
  SessionUser,
  UseReviewWizardReturn,
}
