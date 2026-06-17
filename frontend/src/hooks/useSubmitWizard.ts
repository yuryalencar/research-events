import { useState, useCallback, useRef } from "react"

import { submitEvent } from "@/lib/api/events"
import { ApiError } from "@/lib/api/client"
import type { DeadlineType, EventTier, SubmitEventResult } from "@/types/api"

// --- Types ---

type WizardStep = 1 | 2 | 3

interface DeadlineFormRow {
  id: string // local key only — never sent to the API
  type: DeadlineType | ""
  description: string
  date: string
  isOptional: boolean
}

interface WizardFormData {
  submitterName: string
  submitterEmail: string
  fullName: string
  slug: string
  domain: string
  tier: EventTier | ""
  startDate: string
  endDate: string
  websiteUrl: string
  country: string
  city: string
  latitude: number | null
  longitude: number | null
  deadlines: DeadlineFormRow[]
}

type WizardErrors = Record<string, string>

interface UseSubmitWizardReturn {
  step: WizardStep
  formData: WizardFormData
  errors: WizardErrors
  isSubmitting: boolean
  submittedEvent: SubmitEventResult | null
  bannerError: string | null
  goToStep: (step: WizardStep) => void
  updateField: (field: keyof WizardFormData, value: unknown) => void
  addDeadline: () => void
  removeDeadline: (id: string) => void
  updateDeadline: (id: string, field: keyof Omit<DeadlineFormRow, "id">, value: unknown) => void
  validateStep2: () => boolean
  submit: (skipDeadlines?: boolean) => Promise<void>
}

// --- Internals ---

const initialFormData: WizardFormData = {
  submitterName: "",
  submitterEmail: "",
  fullName: "",
  slug: "",
  domain: "computer_science",
  tier: "",
  startDate: "",
  endDate: "",
  websiteUrl: "",
  country: "",
  city: "",
  latitude: null,
  longitude: null,
  deadlines: [],
}

// computeErrors is a pure function — same input always produces the same
// output with no side effects. Defined at module level so it can be tested
// or reused independently of the hook.
function computeErrors(data: WizardFormData): WizardErrors {
  const errors: WizardErrors = {}

  if (!data.submitterName.trim()) {
    errors.submitterName = "Your name is required"
  }
  if (!data.submitterEmail.trim()) {
    errors.submitterEmail = "Your email is required"
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(data.submitterEmail)) {
    errors.submitterEmail = "Enter a valid email address"
  }
  if (!data.fullName.trim()) {
    errors.fullName = "Full event name is required"
  }
  if (!data.slug.trim()) {
    errors.slug = "Slug is required"
  } else if (!/^[A-Za-z0-9_-]+$/.test(data.slug)) {
    errors.slug = "Slug can only contain letters, numbers, hyphens, and underscores"
  }
  if (!data.domain) {
    errors.domain = "Domain is required"
  }
  if (!data.startDate) {
    errors.startDate = "Start date is required"
  }
  if (!data.endDate) {
    errors.endDate = "End date is required"
  } else if (data.startDate && data.endDate < data.startDate) {
    // ISO date strings (YYYY-MM-DD) sort lexicographically, so < works correctly.
    errors.endDate = "End date must be on or after the start date"
  }
  if (!data.websiteUrl.trim()) {
    errors.websiteUrl = "Website URL is required"
  } else if (!/^https?:\/\/.+/.test(data.websiteUrl)) {
    errors.websiteUrl = "Website URL must start with http:// or https://"
  }
  if (!data.country) {
    errors.country = "Country is required"
  }
  if (!data.city.trim()) {
    errors.city = "City is required"
  }
  if (data.latitude === null || data.longitude === null) {
    errors.location = "Please drop a pin on the map to set the event location"
  }

  return errors
}

// --- Hook ---

// useSubmitWizard manages the 3-step event submission wizard: step navigation,
// form data, validation, and API submission.
//
// errors are stored in a useRef and mutated in place so that validateStep2()
// results are immediately visible to callers without waiting for a React
// re-render. A separate counter state (errorVersion) triggers re-renders so
// that components observe the change on the next paint.
function useSubmitWizard(): UseSubmitWizardReturn {
  const [step, setStep] = useState<WizardStep>(1)
  const [formData, setFormData] = useState<WizardFormData>(initialFormData)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [submittedEvent, setSubmittedEvent] = useState<SubmitEventResult | null>(null)
  const [bannerError, setBannerError] = useState<string | null>(null)

  // Errors live in a ref so that validateStep2() callers see the updated
  // value immediately (same object reference — in-place mutation).
  const errorsRef = useRef<WizardErrors>({})
  const [, setErrorVersion] = useState(0)

  const applyErrors = useCallback((newErrors: WizardErrors): void => {
    const current = errorsRef.current
    Object.keys(current).forEach(k => {
      delete (current as Record<string, string>)[k]
    })
    Object.assign(current, newErrors)
    setErrorVersion(v => v + 1)
  }, [])

  const goToStep = useCallback((s: WizardStep): void => {
    setStep(s)
  }, [])

  const updateField = useCallback((field: keyof WizardFormData, value: unknown): void => {
    setFormData(prev => ({ ...prev, [field]: value }))
  }, [])

  const addDeadline = useCallback((): void => {
    const row: DeadlineFormRow = {
      id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
      type: "",
      description: "",
      date: "",
      isOptional: false,
    }
    setFormData(prev => ({ ...prev, deadlines: [...prev.deadlines, row] }))
  }, [])

  const removeDeadline = useCallback((id: string): void => {
    setFormData(prev => ({
      ...prev,
      deadlines: prev.deadlines.filter(d => d.id !== id),
    }))
  }, [])

  const updateDeadline = useCallback(
    (id: string, field: keyof Omit<DeadlineFormRow, "id">, value: unknown): void => {
      setFormData(prev => ({
        ...prev,
        deadlines: prev.deadlines.map(d => (d.id === id ? { ...d, [field]: value } : d)),
      }))
    },
    [],
  )

  const validateStep2 = useCallback((): boolean => {
    const newErrors = computeErrors(formData)
    applyErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }, [formData, applyErrors])

  const submit = useCallback(
    async (skipDeadlines = false): Promise<void> => {
      setIsSubmitting(true)
      setBannerError(null)

      const deadlines = skipDeadlines
        ? []
        : formData.deadlines.map(d => ({
            type: d.type as DeadlineType,
            description: d.description,
            date: d.date,
            is_optional: d.isOptional,
          }))

      // Omit tier from payload when empty — backend defaults it to "unranked".
      const payload = {
        name: formData.fullName,
        slug: formData.slug,
        country: formData.country,
        city: formData.city,
        latitude: formData.latitude as number,
        longitude: formData.longitude as number,
        start_date: formData.startDate,
        end_date: formData.endDate,
        website_url: formData.websiteUrl,
        domain: formData.domain,
        ...(formData.tier ? { tier: formData.tier as EventTier } : {}),
        submitter: {
          name: formData.submitterName,
          email: formData.submitterEmail,
        },
        deadlines,
      }

      try {
        const result = await submitEvent(payload)
        setSubmittedEvent(result)
      } catch (err) {
        if (err instanceof ApiError) {
          if (err.status === 409) {
            // Spec: "409 → Return to Step 2; slug field shows error"
            applyErrors({ slug: "This slug is already taken by a pending or approved event" })
            setStep(2)
          } else if (err.status === 400) {
            // Spec: "400 → Return to Step 2; per-field errors shown"
            setBannerError(err.message)
            setStep(2)
          } else {
            // Spec: "429 / network error → banner on Step 3"
            setBannerError(err.message)
          }
        } else {
          setBannerError("An unexpected error occurred. Please try again.")
        }
      } finally {
        setIsSubmitting(false)
      }
    },
    [formData, applyErrors],
  )

  return {
    step,
    formData,
    errors: errorsRef.current,
    isSubmitting,
    submittedEvent,
    bannerError,
    goToStep,
    updateField,
    addDeadline,
    removeDeadline,
    updateDeadline,
    validateStep2,
    submit,
  }
}

// --- Export ---

export { useSubmitWizard }
export type { UseSubmitWizardReturn, WizardStep, WizardFormData, WizardErrors, DeadlineFormRow }
