"use client"

import type { JSX } from "react"
import { useTranslations, useLocale } from "next-intl"
import { useRouter } from "next/navigation"
import { Eye, EyeOff, Check, X } from "lucide-react"
import Link from "next/link"

import { useUpdatePassword } from "@/hooks/useUpdatePassword"

// --- Types ---

interface UpdatePasswordCardProps {
  dashboardHref: string
}

// --- Sub-components ---

interface ComplexityItemProps {
  met: boolean
  label: string
}

function ComplexityItem({ met, label }: ComplexityItemProps): JSX.Element {
  return (
    <li className={`flex items-center gap-1.5 text-xs ${met ? "text-green-600" : "text-red-500"}`}>
      {met ? <Check size={12} strokeWidth={3} /> : <X size={12} strokeWidth={3} />}
      {label}
    </li>
  )
}

interface PasswordFieldProps {
  id: string
  label: string
  value: string
  onChange: (v: string) => void
  show: boolean
  onToggleShow: () => void
  showLabel: string
  hideLabel: string
  disabled: boolean
  autoComplete: string
}

function PasswordField({
  id,
  label,
  value,
  onChange,
  show,
  onToggleShow,
  showLabel,
  hideLabel,
  disabled,
  autoComplete,
}: PasswordFieldProps): JSX.Element {
  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={id} className="text-sm font-medium text-foreground">
        {label}
      </label>
      <div className="relative">
        <input
          id={id}
          type={show ? "text" : "password"}
          autoComplete={autoComplete}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
          className="w-full rounded-md border border-border bg-background px-3 py-2 pr-10 text-sm focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
        />
        <button
          type="button"
          onClick={onToggleShow}
          disabled={disabled}
          className="absolute inset-y-0 right-0 flex items-center px-3 text-muted-foreground hover:text-foreground disabled:opacity-50"
          aria-label={show ? hideLabel : showLabel}
        >
          {show ? <EyeOff size={16} /> : <Eye size={16} />}
        </button>
      </div>
    </div>
  )
}

// --- Component ---

function UpdatePasswordCard({ dashboardHref }: UpdatePasswordCardProps): JSX.Element {
  const t = useTranslations("manage.updatePassword")
  // Root-level translator so handleApiError can resolve "errors.CODE" paths correctly.
  // Passing useTranslations("errors") would prepend the namespace twice → "errors.errors.CODE".
  const tRoot = useTranslations()
  const locale = useLocale()
  const router = useRouter()

  function handleAuthError(): void {
    localStorage.removeItem("manage_user")
    router.replace(`/${locale}/manage`)
  }

  const {
    currentPassword, setCurrentPassword,
    newPassword, setNewPassword,
    confirmPassword, setConfirmPassword,
    showCurrent, toggleShowCurrent,
    showNew, toggleShowNew,
    showConfirm, toggleShowConfirm,
    complexity,
    confirmMismatch,
    canSubmit,
    phase,
    submit,
  } = useUpdatePassword(tRoot, handleAuthError)

  const isSubmitting = phase === "submitting"

  if (phase === "success") {
    return (
      <main className="flex min-h-screen items-center justify-center bg-background p-4">
        <div className="w-full max-w-sm rounded-xl border border-border bg-card p-8 shadow-sm text-center">
          <div className="mb-4 flex justify-center">
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-green-100">
              <Check size={24} className="text-green-600" strokeWidth={2.5} />
            </div>
          </div>
          <h1 className="mb-2 text-xl font-semibold text-foreground">{t("success.title")}</h1>
          <p className="mb-6 text-sm text-muted-foreground">{t("success.description")}</p>
          <Link
            href={dashboardHref}
            className="inline-block rounded-md bg-primary px-5 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
          >
            {t("success.backToDashboard")}
          </Link>
        </div>
      </main>
    )
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-background p-4">
      <div className="w-full max-w-sm rounded-xl border border-border bg-card p-8 shadow-sm">
        <h1 className="mb-6 text-center text-2xl font-semibold tracking-tight text-foreground">
          {t("title")}
        </h1>

        <form
          onSubmit={(e) => { e.preventDefault(); void submit() }}
          noValidate
          className="flex flex-col gap-4"
        >
          {/* Current password */}
          <PasswordField
            id="current-password"
            label={t("currentPasswordLabel")}
            value={currentPassword}
            onChange={setCurrentPassword}
            show={showCurrent}
            onToggleShow={toggleShowCurrent}
            showLabel={t("showPassword")}
            hideLabel={t("hidePassword")}
            disabled={isSubmitting}
            autoComplete="current-password"
          />

          {/* New password + complexity checklist */}
          <div className="flex flex-col gap-1">
            <PasswordField
              id="new-password"
              label={t("newPasswordLabel")}
              value={newPassword}
              onChange={setNewPassword}
              show={showNew}
              onToggleShow={toggleShowNew}
              showLabel={t("showPassword")}
              hideLabel={t("hidePassword")}
              disabled={isSubmitting}
              autoComplete="new-password"
            />
            <ul className="mt-1 flex flex-col gap-0.5 pl-1">
              <ComplexityItem met={complexity.minLength} label={t("complexity.minLength")} />
              <ComplexityItem met={complexity.hasUppercase} label={t("complexity.hasUppercase")} />
              <ComplexityItem met={complexity.hasLowercase} label={t("complexity.hasLowercase")} />
              <ComplexityItem met={complexity.hasSpecial} label={t("complexity.hasSpecial")} />
            </ul>
          </div>

          {/* Confirm new password */}
          <div className="flex flex-col gap-1">
            <PasswordField
              id="confirm-password"
              label={t("confirmPasswordLabel")}
              value={confirmPassword}
              onChange={setConfirmPassword}
              show={showConfirm}
              onToggleShow={toggleShowConfirm}
              showLabel={t("showPassword")}
              hideLabel={t("hidePassword")}
              disabled={isSubmitting}
              autoComplete="new-password"
            />
            {confirmMismatch && (
              <p className="text-xs text-red-500">{t("confirmMismatch")}</p>
            )}
          </div>

          <button
            type="submit"
            disabled={!canSubmit || isSubmitting}
            className="mt-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-50"
          >
            {isSubmitting ? t("savingButton") : t("saveButton")}
          </button>
        </form>

        <div className="mt-4 text-center">
          <Link
            href={dashboardHref}
            className="text-sm text-muted-foreground underline-offset-4 hover:underline"
          >
            {t("backToDashboard")}
          </Link>
        </div>
      </div>
    </main>
  )
}

// --- Export ---

export { UpdatePasswordCard }
export type { UpdatePasswordCardProps }
