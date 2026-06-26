import { useState, useMemo, useCallback } from "react"

import { changeUserRole, resetUserPassword, unlockUser } from "@/lib/api/admin"
import { checkPasswordComplexity } from "@/lib/utils"
import { errorMessageKey } from "@/lib/api/errors"
import type { AdminUserListItem } from "@/types/api"
import type { PasswordComplexity } from "@/lib/utils"

// extractErrorKey returns the i18n key for an error so the component can
// translate it with its own t() function — keeping the hook decoupled from i18n.
function extractErrorKey(err: unknown): string {
  return errorMessageKey(err instanceof Error ? (err as { code?: string }).code ?? "UNKNOWN" : "UNKNOWN")
}

// --- Types ---

type CardSectionPhase = "idle" | "submitting" | "success" | "error"

interface UseUserCardReturn {
  isExpanded: boolean
  toggleExpanded: () => void
  // role section
  selectedRole: string
  setSelectedRole: (role: string) => void
  canApplyRole: boolean
  rolePhase: CardSectionPhase
  roleError: string | null
  roleConfirmOpen: boolean
  openRoleConfirm: () => void
  closeRoleConfirm: () => void
  submitRoleChange: () => Promise<void>
  // password section
  newPassword: string
  setNewPassword: (v: string) => void
  confirmPassword: string
  setConfirmPassword: (v: string) => void
  showNew: boolean
  toggleShowNew: () => void
  showConfirm: boolean
  toggleShowConfirm: () => void
  complexity: PasswordComplexity
  passwordMismatch: boolean
  canApplyPassword: boolean
  passwordPhase: CardSectionPhase
  passwordError: string | null
  passwordConfirmOpen: boolean
  openPasswordConfirm: () => void
  closePasswordConfirm: () => void
  submitPasswordReset: () => Promise<void>
  // unlock section
  unlockPhase: CardSectionPhase
  unlockError: string | null
  submitUnlock: () => Promise<void>
  // live user state (optimistically updated after successful actions)
  currentRole: string
  isLocked: boolean
}

// --- Hook ---

function useUserCard(
  user: AdminUserListItem,
  onRoleChanged: (newRole: string) => void,
  onUnlocked: () => void,
): UseUserCardReturn {
  // 1. State — expand toggle
  const [isExpanded, setIsExpanded] = useState(false)

  // 2. State — live user values (update optimistically on success)
  const [currentRole, setCurrentRole] = useState(user.role)
  const [isLocked, setIsLocked] = useState(user.locked_at !== null)

  // 3. State — role section
  const [selectedRole, setSelectedRole] = useState(user.role)
  const [rolePhase, setRolePhase] = useState<CardSectionPhase>("idle")
  const [roleError, setRoleError] = useState<string | null>(null)
  const [roleConfirmOpen, setRoleConfirmOpen] = useState(false)

  // 4. State — password section
  const [newPassword, setNewPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [showNew, setShowNew] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)
  const [passwordPhase, setPasswordPhase] = useState<CardSectionPhase>("idle")
  const [passwordError, setPasswordError] = useState<string | null>(null)
  const [passwordConfirmOpen, setPasswordConfirmOpen] = useState(false)

  // 5. State — unlock section
  const [unlockPhase, setUnlockPhase] = useState<CardSectionPhase>("idle")
  const [unlockError, setUnlockError] = useState<string | null>(null)

  // 6. Derived values
  const canApplyRole = selectedRole !== currentRole

  const complexity = useMemo(() => checkPasswordComplexity(newPassword), [newPassword])

  const allComplexityMet =
    complexity.minLength &&
    complexity.hasUppercase &&
    complexity.hasLowercase &&
    complexity.hasSpecial

  // Mismatch is only shown once the confirm field has been filled.
  const passwordMismatch = confirmPassword.length > 0 && confirmPassword !== newPassword

  const canApplyPassword =
    newPassword.length > 0 && allComplexityMet && confirmPassword === newPassword

  // 7. Handlers — expand
  const toggleExpanded = useCallback(() => setIsExpanded((v) => !v), [])

  // 8. Handlers — role section
  const openRoleConfirm = useCallback(() => setRoleConfirmOpen(true), [])
  const closeRoleConfirm = useCallback(() => setRoleConfirmOpen(false), [])
  const toggleShowNew = useCallback(() => setShowNew((v) => !v), [])
  const toggleShowConfirm = useCallback(() => setShowConfirm((v) => !v), [])

  const submitRoleChange = useCallback(async (): Promise<void> => {
    if (!canApplyRole) return
    setRolePhase("submitting")
    setRoleConfirmOpen(false)
    try {
      await changeUserRole(user.id, selectedRole)
      setCurrentRole(selectedRole)
      setRolePhase("success")
      setRoleError(null)
      onRoleChanged(selectedRole)
    } catch (err) {
      setRolePhase("error")
      setRoleError(extractErrorKey(err))
      setSelectedRole(currentRole)
    }
  }, [canApplyRole, user.id, selectedRole, currentRole, onRoleChanged])

  // 9. Handlers — password section
  const openPasswordConfirm = useCallback(() => setPasswordConfirmOpen(true), [])
  const closePasswordConfirm = useCallback(() => setPasswordConfirmOpen(false), [])

  const submitPasswordReset = useCallback(async (): Promise<void> => {
    if (!canApplyPassword) return
    setPasswordPhase("submitting")
    setPasswordConfirmOpen(false)
    try {
      await resetUserPassword(user.id, {
        new_password: newPassword,
        new_password_confirmation: confirmPassword,
      })
      setPasswordPhase("success")
      setPasswordError(null)
      setNewPassword("")
      setConfirmPassword("")
    } catch (err) {
      setPasswordPhase("error")
      setPasswordError(extractErrorKey(err))
    }
  }, [canApplyPassword, user.id, newPassword, confirmPassword])

  // 10. Handlers — unlock section
  const submitUnlock = useCallback(async (): Promise<void> => {
    setUnlockPhase("submitting")
    try {
      await unlockUser(user.id)
      setIsLocked(false)
      setUnlockPhase("success")
      setUnlockError(null)
      onUnlocked()
    } catch (err) {
      setUnlockPhase("error")
      setUnlockError(extractErrorKey(err))
    }
  }, [user.id, onUnlocked])

  // 11. Return
  return {
    isExpanded,
    toggleExpanded,
    selectedRole,
    setSelectedRole,
    canApplyRole,
    rolePhase,
    roleError,
    roleConfirmOpen,
    openRoleConfirm,
    closeRoleConfirm,
    submitRoleChange,
    newPassword,
    setNewPassword,
    confirmPassword,
    setConfirmPassword,
    showNew,
    toggleShowNew,
    showConfirm,
    toggleShowConfirm,
    complexity,
    passwordMismatch,
    canApplyPassword,
    passwordPhase,
    passwordError,
    passwordConfirmOpen,
    openPasswordConfirm,
    closePasswordConfirm,
    submitPasswordReset,
    unlockPhase,
    unlockError,
    submitUnlock,
    currentRole,
    isLocked,
  }
}

// --- Export ---

export { useUserCard }
export type { UseUserCardReturn, CardSectionPhase }
