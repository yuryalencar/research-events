import type { JSX } from "react"
import { useTranslations } from "next-intl"

import type { WizardStep } from "@/hooks/useSubmitWizard"

// --- Types ---

interface StepIndicatorProps {
  currentStep: WizardStep
  totalSteps: number
}

// --- Component ---

function StepIndicator({ currentStep, totalSteps }: StepIndicatorProps): JSX.Element {
  const t = useTranslations("submit")

  return (
    <div className="flex items-center gap-3" aria-label={t("stepIndicator", { current: currentStep, total: totalSteps })}>
      {Array.from({ length: totalSteps }, (_, i) => {
        const step = (i + 1) as WizardStep
        const isComplete = step < currentStep
        const isCurrent = step === currentStep

        return (
          <div key={step} className="flex items-center gap-3">
            <div
              aria-current={isCurrent ? "step" : undefined}
              className={[
                "flex size-8 items-center justify-center rounded-full border-2 text-sm font-semibold transition-colors",
                isCurrent
                  ? "border-primary bg-primary text-primary-foreground"
                  : isComplete
                    ? "border-primary bg-primary/20 text-primary"
                    : "border-muted-foreground/30 text-muted-foreground",
              ].join(" ")}
            >
              {step}
            </div>
            {i < totalSteps - 1 && (
              <div
                className={[
                  "h-0.5 w-8",
                  isComplete ? "bg-primary" : "bg-muted-foreground/20",
                ].join(" ")}
              />
            )}
          </div>
        )
      })}
      <span className="ml-1 text-sm text-muted-foreground">
        {t("stepIndicator", { current: currentStep, total: totalSteps })}
      </span>
    </div>
  )
}

// --- Export ---

export { StepIndicator }
export type { StepIndicatorProps }
