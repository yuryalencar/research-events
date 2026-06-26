import { useState, useMemo, useCallback } from "react"

import { registerAdminUser } from "@/lib/api/admin"
import { checkPasswordComplexity } from "@/lib/utils"
import { errorMessageKey } from "@/lib/api/errors"
import { ApiError } from "@/lib/api/client"
import type { RegisterAdminUserResult } from "@/types/api"
import type { PasswordComplexity } from "@/lib/utils"

// --- Types ---

type RegisterPhase = "idle" | "confirm" | "submitting" | "success" | "error"

interface UseRegisterUserReturn {
  name: string
  setName: (v: string) => void
  email: string
  setEmail: (v: string) => void
  role: string
  setRole: (v: string) => void
  password: string
  setPassword: (v: string) => void
  confirmPassword: string
  setConfirmPassword: (v: string) => void
  showPassword: boolean
  toggleShowPassword: () => void
  showConfirm: boolean
  toggleShowConfirm: () => void
  complexity: PasswordComplexity
  passwordMismatch: boolean
  canSubmit: boolean
  phase: RegisterPhase
  error: string | null
  registeredUser: RegisterAdminUserResult | null
  openConfirm: () => void
  closeConfirm: () => void
  submit: () => Promise<void>
  reset: () => void
}

// Codes that indicate a dead session — warrant a redirect to login.
const AUTH_ERROR_CODES = new Set(["TOKEN_MISSING", "TOKEN_EXPIRED", "TOKEN_INVALID"])

// --- Hook ---

function useRegisterUser(onAuthError: () => void): UseRegisterUserReturn {
  // 1. State — form fields
  const [name, setName] = useState("")
  const [email, setEmail] = useState("")
  const [role, setRole] = useState("")
  const [password, setPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)

  // 2. State — submit lifecycle
  const [phase, setPhase] = useState<RegisterPhase>("idle")
  const [error, setError] = useState<string | null>(null)
  const [registeredUser, setRegisteredUser] = useState<RegisterAdminUserResult | null>(null)

  // 3. Derived values
  const complexity = useMemo(() => checkPasswordComplexity(password), [password])

  const allComplexityMet =
    complexity.minLength &&
    complexity.hasUppercase &&
    complexity.hasLowercase &&
    complexity.hasSpecial

  // Mismatch only shown once confirm field has content.
  const passwordMismatch = confirmPassword.length > 0 && confirmPassword !== password

  const canSubmit =
    name.length > 0 &&
    email.length > 0 &&
    role.length > 0 &&
    allComplexityMet &&
    confirmPassword === password &&
    password.length > 0

  // 4. Handlers
  const toggleShowPassword = useCallback(() => setShowPassword((v) => !v), [])
  const toggleShowConfirm = useCallback(() => setShowConfirm((v) => !v), [])

  const openConfirm = useCallback(() => setPhase("confirm"), [])
  const closeConfirm = useCallback(() => setPhase("idle"), [])

  const submit = useCallback(async (): Promise<void> => {
    if (!canSubmit) return
    setPhase("submitting")
    setError(null)
    try {
      const result = await registerAdminUser({ name, email, password, role })
      setRegisteredUser(result)
      setPhase("success")
    } catch (err) {
      const key = errorMessageKey(
        err instanceof ApiError ? err.code : "UNKNOWN",
      )
      setError(key)
      setPhase("error")
      if (err instanceof ApiError && AUTH_ERROR_CODES.has(err.code)) {
        onAuthError()
      }
    }
  }, [canSubmit, name, email, password, role, onAuthError])

  const reset = useCallback(() => {
    setName("")
    setEmail("")
    setRole("")
    setPassword("")
    setConfirmPassword("")
    setShowPassword(false)
    setShowConfirm(false)
    setPhase("idle")
    setError(null)
    setRegisteredUser(null)
  }, [])

  // 5. Return
  return {
    name,
    setName,
    email,
    setEmail,
    role,
    setRole,
    password,
    setPassword,
    confirmPassword,
    setConfirmPassword,
    showPassword,
    toggleShowPassword,
    showConfirm,
    toggleShowConfirm,
    complexity,
    passwordMismatch,
    canSubmit,
    phase,
    error,
    registeredUser,
    openConfirm,
    closeConfirm,
    submit,
    reset,
  }
}

// --- Export ---

export { useRegisterUser }
export type { UseRegisterUserReturn, RegisterPhase }
