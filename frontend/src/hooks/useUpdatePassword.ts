import { useState, useMemo, useCallback } from "react"

import { updatePassword } from "@/lib/api/users"
import { handleApiError } from "@/lib/api/errors"
import { ApiError } from "@/lib/api/client"

// Codes that mean the session is gone (not just a bad request) — warrant a redirect to login.
const AUTH_ERROR_CODES = new Set(["TOKEN_MISSING", "TOKEN_EXPIRED", "TOKEN_INVALID"])

// --- Types ---

interface PasswordComplexity {
  minLength: boolean
  hasUppercase: boolean
  hasLowercase: boolean
  hasSpecial: boolean
}

type UpdatePasswordPhase = "idle" | "submitting" | "success"

interface UseUpdatePasswordReturn {
  currentPassword: string
  setCurrentPassword: (v: string) => void
  newPassword: string
  setNewPassword: (v: string) => void
  confirmPassword: string
  setConfirmPassword: (v: string) => void
  showCurrent: boolean
  toggleShowCurrent: () => void
  showNew: boolean
  toggleShowNew: () => void
  showConfirm: boolean
  toggleShowConfirm: () => void
  complexity: PasswordComplexity
  confirmMismatch: boolean
  canSubmit: boolean
  phase: UpdatePasswordPhase
  submit: () => Promise<void>
}

// --- Hook ---

function useUpdatePassword(t: (key: string) => string, onAuthError?: () => void): UseUpdatePasswordReturn {
  // 1. State
  const [currentPassword, setCurrentPassword] = useState("")
  const [newPassword, setNewPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [showCurrent, setShowCurrent] = useState(false)
  const [showNew, setShowNew] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)
  const [phase, setPhase] = useState<UpdatePasswordPhase>("idle")

  // 2. Derived values
  const complexity = useMemo<PasswordComplexity>(() => ({
    minLength: newPassword.length >= 8,
    hasUppercase: /[A-Z]/.test(newPassword),
    hasLowercase: /[a-z]/.test(newPassword),
    hasSpecial: /[^A-Za-z0-9]/.test(newPassword),
  }), [newPassword])

  // Mismatch is only shown once the user has started typing in the confirm field.
  const confirmMismatch = confirmPassword.length > 0 && confirmPassword !== newPassword

  const allComplexityMet =
    complexity.minLength &&
    complexity.hasUppercase &&
    complexity.hasLowercase &&
    complexity.hasSpecial

  const canSubmit =
    currentPassword.length > 0 &&
    allComplexityMet &&
    confirmPassword === newPassword

  // 3. Handlers
  const toggleShowCurrent = useCallback(() => setShowCurrent((v) => !v), [])
  const toggleShowNew = useCallback(() => setShowNew((v) => !v), [])
  const toggleShowConfirm = useCallback(() => setShowConfirm((v) => !v), [])

  const submit = useCallback(async (): Promise<void> => {
    if (!canSubmit) return

    setPhase("submitting")

    try {
      await updatePassword({
        current_password: currentPassword,
        new_password: newPassword,
        new_password_confirmation: confirmPassword,
      })
      setPhase("success")
    } catch (err) {
      handleApiError(err, t)
      setPhase("idle")
      if (err instanceof ApiError && AUTH_ERROR_CODES.has(err.code)) {
        onAuthError?.()
      }
    }
  }, [canSubmit, currentPassword, newPassword, confirmPassword, t, onAuthError])

  // 4. Return
  return {
    currentPassword,
    setCurrentPassword,
    newPassword,
    setNewPassword,
    confirmPassword,
    setConfirmPassword,
    showCurrent,
    toggleShowCurrent,
    showNew,
    toggleShowNew,
    showConfirm,
    toggleShowConfirm,
    complexity,
    confirmMismatch,
    canSubmit,
    phase,
    submit,
  }
}

// --- Export ---

export { useUpdatePassword }
export type { UseUpdatePasswordReturn, PasswordComplexity, UpdatePasswordPhase }
