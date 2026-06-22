"use client"

import type { JSX, FormEvent } from "react"
import { useState, useEffect } from "react"
import Image from "next/image"
import { useRouter } from "next/navigation"
import { useLocale, useTranslations } from "next-intl"
import { Eye, EyeOff } from "lucide-react"

import { login } from "@/lib/api/auth"
import { ApiError } from "@/lib/api/client"
import { decodeJwt } from "@/lib/utils/decodeJwt"
import { ADMIN_EMAIL } from "@/lib/constants"

// --- Types ---

type LoginPhase = "ready" | "submitting"

// --- Component ---

export default function ManageLoginPage(): JSX.Element {
  const t = useTranslations("manage")
  const tErrors = useTranslations("errors")
  const locale = useLocale()
  const router = useRouter()

  const [phase, setPhase] = useState<LoginPhase>("ready")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [showPassword, setShowPassword] = useState(false)
  const [fieldErrors, setFieldErrors] = useState<{ email?: string; password?: string }>({})
  const [apiError, setApiError] = useState<string | null>(null)

  // If the user already has a stored session, redirect them to their dashboard
  // immediately. This is a sync localStorage read — no API call needed.
  useEffect(() => {
    try {
      const stored = localStorage.getItem("manage_user")
      if (!stored) return
      const parsed = JSON.parse(stored) as { role?: string }
      if (parsed.role === "admin" || parsed.role === "moderator") {
        router.replace(`/${locale}/manage/${parsed.role}`)
      }
    } catch {
      localStorage.removeItem("manage_user")
    }
  }, [locale, router])

  function validate(): boolean {
    const errors: { email?: string; password?: string } = {}
    if (!email.trim()) errors.email = t("login.emailRequired")
    if (!password.trim()) errors.password = t("login.passwordRequired")
    setFieldErrors(errors)
    return Object.keys(errors).length === 0
  }

  async function handleSubmit(e: FormEvent<HTMLFormElement>): Promise<void> {
    e.preventDefault()
    if (!validate()) return

    setPhase("submitting")
    setApiError(null)

    try {
      const result = await login({ email: email.trim(), password })
      const claims = decodeJwt(result.token)

      if (claims.role !== "admin" && claims.role !== "moderator") {
        setApiError(tErrors("FORBIDDEN"))
        setPhase("ready")
        return
      }

      localStorage.setItem(
        "manage_user",
        JSON.stringify({ id: parseInt(claims.sub, 10), name: claims.name, role: claims.role, email: claims.email }),
      )
      router.push(`/${locale}/manage/${claims.role}`)
    } catch (err: unknown) {
      const code = err instanceof ApiError ? err.code : "UNKNOWN"
      setApiError(tErrors(code as Parameters<typeof tErrors>[0]))
      setPhase("ready")
    }
  }

  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-4 bg-background p-4">
      {/* Login card */}
      <div className="w-full max-w-sm rounded-xl border border-border bg-card p-8 shadow-sm">
        <div className="mb-6 flex justify-center">
          <Image src="/logo-with-opensource.png" alt="ReSEARCH Events" width={325} height={130} style={{ width: 305, height: "auto" }} />
        </div>

        <form onSubmit={handleSubmit} noValidate className="flex flex-col gap-4">
          {/* Email field */}
          <div className="flex flex-col gap-1">
            <label htmlFor="email" className="text-sm font-medium text-foreground">
              {t("login.emailLabel")}
            </label>
            <input
              id="email"
              type="email"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={phase === "submitting"}
              className={
                fieldErrors.email
                  ? "rounded-md border border-red-400 bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-red-400 disabled:opacity-50"
                  : "rounded-md border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
              }
            />
            {fieldErrors.email && (
              <p className="text-xs text-red-500">{fieldErrors.email}</p>
            )}
          </div>

          {/* Password field */}
          <div className="flex flex-col gap-1">
            <label htmlFor="password" className="text-sm font-medium text-foreground">
              {t("login.passwordLabel")}
            </label>
            <div className="relative">
              <input
                id="password"
                type={showPassword ? "text" : "password"}
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={phase === "submitting"}
                className={
                  fieldErrors.password
                    ? "w-full rounded-md border border-red-400 bg-background px-3 py-2 pr-10 text-sm focus:outline-none focus:ring-2 focus:ring-red-400 disabled:opacity-50"
                    : "w-full rounded-md border border-border bg-background px-3 py-2 pr-10 text-sm focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
                }
              />
              <button
                type="button"
                onClick={() => setShowPassword((v) => !v)}
                disabled={phase === "submitting"}
                className="absolute inset-y-0 right-0 flex items-center px-3 text-muted-foreground hover:text-foreground disabled:opacity-50"
                aria-label={showPassword ? "Hide password" : "Show password"}
              >
                {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
            </div>
            {fieldErrors.password && (
              <p className="text-xs text-red-500">{fieldErrors.password}</p>
            )}
          </div>

          {/* API error */}
          {apiError && (
            <p className="text-sm text-red-500">{apiError}</p>
          )}

          <button
            type="submit"
            disabled={phase === "submitting"}
            className="mt-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-50"
          >
            {phase === "submitting" ? t("dashboard.loading") : t("login.submitButton")}
          </button>
        </form>
      </div>

      {/* Become a moderator card */}
      <div className="w-full max-w-sm rounded-xl border border-border bg-card p-6 shadow-sm">
        <h2 className="mb-1 text-base font-semibold text-foreground">
          {t("moderatorBanner.title")}
        </h2>
        <p className="mb-3 text-sm text-muted-foreground">
          {t("moderatorBanner.description")}
        </p>
        <a
          href={`mailto:${ADMIN_EMAIL}`}
          className="text-sm font-medium text-primary underline-offset-4 hover:underline"
        >
          {t("moderatorBanner.emailLabel")}
        </a>
      </div>
    </main>
  )
}
