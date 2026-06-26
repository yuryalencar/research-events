"use client"

import type { JSX } from "react"
import Link from "next/link"
import { useTranslations } from "next-intl"
import { Check } from "lucide-react"

import { useRegisterUser } from "@/hooks/useRegisterUser"
import { PasswordField, ComplexityItem } from "@/components/ui/PasswordField"

// --- Types ---

interface RegisterUserFormProps {
  locale: string
  onAuthError: () => void
}

// --- Sub-components ---

interface FieldProps {
  id: string
  label: string
  children: React.ReactNode
}

function Field({ id, label, children }: FieldProps): JSX.Element {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="text-sm font-medium text-foreground">
        {label}
      </label>
      {children}
    </div>
  )
}

interface ConfirmModalProps {
  email: string
  role: string
  onCancel: () => void
  onConfirm: () => void
  submitting: boolean
  t: ReturnType<typeof useTranslations>
}

function ConfirmModal({ email, role, onCancel, onConfirm, submitting, t }: ConfirmModalProps): JSX.Element {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="w-full max-w-sm rounded-xl border border-border bg-card p-6 shadow-lg">
        <h2 className="mb-2 text-base font-semibold text-foreground">
          {t("register.confirmTitle")}
        </h2>
        <p className="mb-6 text-sm text-muted-foreground">
          {t("register.confirmBody", { role, email })}
        </p>
        <div className="flex justify-end gap-3">
          <button
            type="button"
            onClick={onCancel}
            disabled={submitting}
            className="rounded-md border border-border px-4 py-2 text-sm text-foreground transition-colors hover:bg-muted disabled:opacity-50"
          >
            {t("register.confirmCancel")}
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={submitting}
            className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-50"
          >
            {submitting ? "…" : t("register.confirmCreate")}
          </button>
        </div>
      </div>
    </div>
  )
}

interface SuccessScreenProps {
  name: string
  email: string
  role: string
  locale: string
  onReset: () => void
  t: ReturnType<typeof useTranslations>
}

function SuccessScreen({ name, email, role, locale, onReset, t }: SuccessScreenProps): JSX.Element {
  return (
    <div className="flex flex-col items-center gap-6 py-8 text-center">
      <div className="flex h-14 w-14 items-center justify-center rounded-full bg-green-100">
        <Check size={28} className="text-green-600" strokeWidth={2.5} />
      </div>
      <div className="flex flex-col gap-1">
        <h2 className="text-xl font-semibold text-foreground">{t("register.success.title")}</h2>
        <p className="text-sm text-muted-foreground">
          {t("register.success.subtitle", { name, role })}
        </p>
      </div>

      <div className="w-full max-w-xs rounded-xl border border-border bg-muted/30 px-5 py-4 text-left">
        <dl className="flex flex-col gap-3">
          <div>
            <dt className="text-xs text-muted-foreground">{t("register.success.nameLabel")}</dt>
            <dd className="mt-0.5 text-sm font-medium text-foreground">{name}</dd>
          </div>
          <div>
            <dt className="text-xs text-muted-foreground">{t("register.success.emailLabel")}</dt>
            <dd className="mt-0.5 text-sm font-medium text-foreground">{email}</dd>
          </div>
          <div>
            <dt className="text-xs text-muted-foreground">{t("register.success.roleLabel")}</dt>
            <dd className="mt-0.5 text-sm font-medium text-foreground">{role}</dd>
          </div>
        </dl>
      </div>

      <div className="flex flex-col items-center gap-3">
        <button
          type="button"
          onClick={onReset}
          className="rounded-md bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
        >
          {t("register.success.registerAnother")}
        </button>
        <Link
          href={`/${locale}/manage/admin/users`}
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          {t("register.success.backToUsers")}
        </Link>
      </div>
    </div>
  )
}

// --- Component ---

const ROLE_OPTIONS = ["admin", "moderator"] as const

function RegisterUserForm({ locale, onAuthError }: RegisterUserFormProps): JSX.Element {
  const t = useTranslations("manage.users")

  const {
    name, setName,
    email, setEmail,
    role, setRole,
    password, setPassword,
    confirmPassword, setConfirmPassword,
    showPassword, toggleShowPassword,
    showConfirm, toggleShowConfirm,
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
  } = useRegisterUser(onAuthError)

  const isSubmitting = phase === "submitting"
  const usersHref = `/${locale}/manage/admin/users`

  if (phase === "success" && registeredUser) {
    return (
      <div className="mx-auto w-full max-w-lg px-4 py-6">
        <SuccessScreen
          name={registeredUser.user.name}
          email={registeredUser.user.email}
          role={registeredUser.user.role}
          locale={locale}
          onReset={reset}
          t={t}
        />
      </div>
    )
  }

  return (
    <>
      {/* Confirmation modal */}
      {phase === "confirm" && (
        <ConfirmModal
          email={email}
          role={role}
          onCancel={closeConfirm}
          onConfirm={() => void submit()}
          submitting={isSubmitting}
          t={t}
        />
      )}

      <div className="mx-auto w-full max-w-lg px-4 py-6">
        {/* Header */}
        <div className="mb-6 flex flex-col gap-1">
          <Link href={usersHref} className="text-xs text-muted-foreground hover:text-foreground">
            {t("register.backToUsers")}
          </Link>
          <h1 className="text-xl font-semibold text-foreground">{t("register.pageTitle")}</h1>
        </div>

        {/* Form */}
        <form
          onSubmit={(e) => {
            e.preventDefault()
            openConfirm()
          }}
          className="flex flex-col gap-5"
        >
          {/* Name */}
          <Field id="register-name" label={t("register.nameLabel")}>
            <input
              id="register-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              autoComplete="name"
              disabled={isSubmitting}
              className="rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
            />
          </Field>

          {/* Email */}
          <Field id="register-email" label={t("register.emailLabel")}>
            <input
              id="register-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoComplete="email"
              disabled={isSubmitting}
              className="rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
            />
          </Field>

          {/* Role */}
          <Field id="register-role" label={t("register.roleLabel")}>
            <select
              id="register-role"
              value={role}
              onChange={(e) => setRole(e.target.value)}
              disabled={isSubmitting}
              className="rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
            >
              <option value="" disabled>
                {t("register.rolePlaceholder")}
              </option>
              {ROLE_OPTIONS.map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </select>
          </Field>

          {/* Password */}
          <div className="flex flex-col gap-2">
            <PasswordField
              id="register-password"
              label={t("register.passwordLabel")}
              value={password}
              onChange={setPassword}
              show={showPassword}
              onToggleShow={toggleShowPassword}
              showLabel="Show"
              hideLabel="Hide"
              disabled={isSubmitting}
              autoComplete="new-password"
            />
            <ul className="flex flex-col gap-0.5 pl-1">
              <ComplexityItem met={complexity.minLength} label="8+ characters" />
              <ComplexityItem met={complexity.hasUppercase} label="Uppercase" />
              <ComplexityItem met={complexity.hasLowercase} label="Lowercase" />
              <ComplexityItem met={complexity.hasSpecial} label="Special character" />
            </ul>
          </div>

          {/* Confirm password */}
          <div className="flex flex-col gap-1">
            <PasswordField
              id="register-confirm"
              label={t("register.confirmLabel")}
              value={confirmPassword}
              onChange={setConfirmPassword}
              show={showConfirm}
              onToggleShow={toggleShowConfirm}
              showLabel="Show"
              hideLabel="Hide"
              disabled={isSubmitting}
              autoComplete="new-password"
            />
            {passwordMismatch && (
              <p className="text-xs text-red-500">{t("register.passwordMismatch")}</p>
            )}
          </div>

          {/* Error banner */}
          {phase === "error" && error && (
            <p className="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
              {t("register.errorBanner")}
            </p>
          )}

          {/* Submit */}
          <button
            type="submit"
            disabled={!canSubmit || isSubmitting}
            className="rounded-md bg-primary py-2.5 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-50"
          >
            {t("register.submitButton")}
          </button>
        </form>
      </div>
    </>
  )
}

// --- Export ---

export { RegisterUserForm }
export type { RegisterUserFormProps }
