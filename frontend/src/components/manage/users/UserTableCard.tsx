"use client"

import type { JSX } from "react"
import { useTranslations } from "next-intl"
import { ChevronDown, ChevronUp, Check } from "lucide-react"

import { useUserCard } from "@/hooks/useUserCard"
import { PasswordField, ComplexityItem } from "@/components/ui/PasswordField"
import type { AdminUserListItem } from "@/types/api"

// --- Types ---

interface UserTableCardProps {
  user: AdminUserListItem
  onRoleChanged: (newRole: string) => void
  onUnlocked: () => void
}

// --- Sub-components ---

interface SectionBannerProps {
  phase: "idle" | "submitting" | "success" | "error"
  successMessage: string
  errorKey: string | null
  t: (key: string) => string
}

function SectionBanner({ phase, successMessage, errorKey, t }: SectionBannerProps): JSX.Element | null {
  if (phase === "success") {
    return (
      <p className="flex items-center gap-1.5 text-xs text-green-600">
        <Check size={12} strokeWidth={3} />
        {successMessage}
      </p>
    )
  }
  if (phase === "error" && errorKey) {
    return <p className="text-xs text-red-500">{t(errorKey)}</p>
  }
  return null
}

interface ConfirmModalProps {
  open: boolean
  title: string
  body: string
  cancelLabel: string
  confirmLabel: string
  onCancel: () => void
  onConfirm: () => void
}

function ConfirmModal({ open, title, body, cancelLabel, confirmLabel, onCancel, onConfirm }: ConfirmModalProps): JSX.Element | null {
  if (!open) return null
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="w-full max-w-sm rounded-xl border border-border bg-card p-6 shadow-lg">
        <h2 className="mb-2 text-base font-semibold text-foreground">{title}</h2>
        <p className="mb-6 text-sm text-muted-foreground">{body}</p>
        <div className="flex justify-end gap-3">
          <button
            type="button"
            onClick={onCancel}
            className="rounded-md border border-border px-4 py-2 text-sm text-foreground transition-colors hover:bg-muted"
          >
            {cancelLabel}
          </button>
          <button
            type="button"
            onClick={onConfirm}
            className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}

// --- Component ---

function UserTableCard({ user, onRoleChanged, onUnlocked }: UserTableCardProps): JSX.Element {
  const t = useTranslations("manage.users")
  const tRoot = useTranslations()

  const {
    isExpanded, toggleExpanded,
    selectedRole, setSelectedRole, canApplyRole,
    rolePhase, roleError, roleConfirmOpen, openRoleConfirm, closeRoleConfirm, submitRoleChange,
    newPassword, setNewPassword, confirmPassword, setConfirmPassword,
    showNew, toggleShowNew, showConfirm, toggleShowConfirm,
    complexity, passwordMismatch, canApplyPassword,
    passwordPhase, passwordError, passwordConfirmOpen, openPasswordConfirm, closePasswordConfirm, submitPasswordReset,
    unlockPhase, unlockError, submitUnlock,
    currentRole, isLocked,
  } = useUserCard(user, onRoleChanged, onUnlocked)

  const isDeleted = user.deleted_at !== null
  const isSubmitting = rolePhase === "submitting" || passwordPhase === "submitting" || unlockPhase === "submitting"

  return (
    <>
      {/* Confirmation modals */}
      <ConfirmModal
        open={roleConfirmOpen}
        title={t("card.roleConfirmTitle")}
        body={t("card.roleConfirmBody", { from: currentRole, to: selectedRole })}
        cancelLabel={t("card.confirmCancel")}
        confirmLabel={t("card.confirmConfirm")}
        onCancel={closeRoleConfirm}
        onConfirm={() => void submitRoleChange()}
      />
      <ConfirmModal
        open={passwordConfirmOpen}
        title={t("card.passwordConfirmTitle")}
        body={t("card.passwordConfirmBody", { name: user.name })}
        cancelLabel={t("card.confirmCancel")}
        confirmLabel={t("card.confirmConfirm")}
        onCancel={closePasswordConfirm}
        onConfirm={() => void submitPasswordReset()}
      />

      <div className="rounded-xl border border-border bg-card shadow-sm">
        {/* Collapsed header */}
        <button
          type="button"
          onClick={toggleExpanded}
          aria-label={isExpanded ? t("card.collapse") : t("card.expand")}
          className="flex w-full items-center gap-3 px-4 py-3 text-left"
        >
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <span className="truncate text-sm font-medium text-foreground">{user.name}</span>
              <span className="text-xs text-muted-foreground">·</span>
              <span className="truncate text-xs text-muted-foreground">{user.email}</span>
            </div>
            <div className="mt-1 flex flex-wrap gap-1.5">
              <span className="inline-block rounded-full border border-border px-2 py-0.5 text-xs text-foreground">
                {currentRole}
              </span>
              {isLocked && (
                <span className="inline-block rounded-full bg-yellow-100 px-2 py-0.5 text-xs text-yellow-800">
                  {t("card.lockedBadge")}
                </span>
              )}
              {isDeleted && (
                <span className="inline-block rounded-full bg-red-100 px-2 py-0.5 text-xs text-red-700">
                  {t("card.deletedBadge")}
                </span>
              )}
            </div>
          </div>
          {isExpanded ? <ChevronUp size={16} className="shrink-0 text-muted-foreground" /> : <ChevronDown size={16} className="shrink-0 text-muted-foreground" />}
        </button>

        {/* Expanded content */}
        {isExpanded && (
          <div className="border-t border-border">
            {isDeleted && (
              <div className="px-4 py-3">
                <p className="text-sm text-muted-foreground">{t("card.deletedBanner")}</p>
              </div>
            )}

            {/* Section 1 — Change Role */}
            <div className="px-4 py-4">
              <p className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                {t("card.roleSection")}
              </p>
              <div className="flex items-center gap-3">
                <select
                  value={selectedRole}
                  onChange={(e) => setSelectedRole(e.target.value)}
                  disabled={isDeleted || isSubmitting}
                  className="rounded-md border border-border bg-background px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                >
                  <option value="admin">admin</option>
                  <option value="moderator">moderator</option>
                  <option value="contributor">contributor</option>
                </select>
                <button
                  type="button"
                  onClick={openRoleConfirm}
                  disabled={!canApplyRole || isDeleted || isSubmitting}
                  className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-50"
                >
                  {t("card.applyRole")}
                </button>
              </div>
              <div className="mt-2">
                <SectionBanner
                  phase={rolePhase}
                  successMessage={t("card.roleSuccess", { role: currentRole })}
                  errorKey={roleError}
                  t={tRoot}
                />
              </div>
            </div>

            {/* Section 2 — Reset Password */}
            <div className="border-t border-border px-4 py-4">
              <p className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                {t("card.passwordSection")}
              </p>
              <div className="flex flex-col gap-3">
                <div className="flex flex-col gap-1">
                  <PasswordField
                    id={`new-pw-${user.id}`}
                    label={t("card.newPasswordLabel")}
                    value={newPassword}
                    onChange={setNewPassword}
                    show={showNew}
                    onToggleShow={toggleShowNew}
                    showLabel="Show"
                    hideLabel="Hide"
                    disabled={isDeleted || isSubmitting}
                    autoComplete="new-password"
                  />
                  <ul className="mt-1 flex flex-col gap-0.5 pl-1">
                    <ComplexityItem met={complexity.minLength} label="8+ characters" />
                    <ComplexityItem met={complexity.hasUppercase} label="Uppercase" />
                    <ComplexityItem met={complexity.hasLowercase} label="Lowercase" />
                    <ComplexityItem met={complexity.hasSpecial} label="Special character" />
                  </ul>
                </div>

                <div className="flex flex-col gap-1">
                  <PasswordField
                    id={`confirm-pw-${user.id}`}
                    label={t("card.confirmPasswordLabel")}
                    value={confirmPassword}
                    onChange={setConfirmPassword}
                    show={showConfirm}
                    onToggleShow={toggleShowConfirm}
                    showLabel="Show"
                    hideLabel="Hide"
                    disabled={isDeleted || isSubmitting}
                    autoComplete="new-password"
                  />
                  {passwordMismatch && (
                    <p className="text-xs text-red-500">{t("card.passwordMismatch")}</p>
                  )}
                </div>

                <div className="flex items-center justify-between">
                  <SectionBanner
                    phase={passwordPhase}
                    successMessage={t("card.passwordSuccess")}
                    errorKey={passwordError}
                    t={tRoot}
                  />
                  <button
                    type="button"
                    onClick={openPasswordConfirm}
                    disabled={!canApplyPassword || isDeleted || isSubmitting}
                    className="ml-auto rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-50"
                  >
                    {t("card.applyPassword")}
                  </button>
                </div>
              </div>
            </div>

            {/* Section 3 — Unlock Account (only when locked) */}
            {isLocked && !isDeleted && (
              <div className="border-t border-border px-4 py-4">
                <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  {t("card.unlockSection")}
                </p>
                <p className="mb-3 text-sm text-muted-foreground">{t("card.unlockDescription")}</p>
                <div className="flex items-center justify-between">
                  <SectionBanner
                    phase={unlockPhase}
                    successMessage={t("card.unlockSuccess")}
                    errorKey={unlockError}
                    t={tRoot}
                  />
                  <button
                    type="button"
                    onClick={() => void submitUnlock()}
                    disabled={unlockPhase === "submitting"}
                    className="ml-auto rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground transition-colors hover:bg-muted disabled:opacity-50"
                  >
                    {t("card.unlockButton")}
                  </button>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </>
  )
}

// --- Export ---

export { UserTableCard }
export type { UserTableCardProps }
