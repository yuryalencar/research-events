"use client"

import type { JSX } from "react"
import { Eye, EyeOff, Check, X } from "lucide-react"

// --- Types ---

interface ComplexityItemProps {
  met: boolean
  label: string
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

// --- Components ---

function ComplexityItem({ met, label }: ComplexityItemProps): JSX.Element {
  return (
    <li className={`flex items-center gap-1.5 text-xs ${met ? "text-green-600" : "text-red-500"}`}>
      {met ? <Check size={12} strokeWidth={3} /> : <X size={12} strokeWidth={3} />}
      {label}
    </li>
  )
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

// --- Export ---

export { PasswordField, ComplexityItem }
export type { PasswordFieldProps, ComplexityItemProps }
