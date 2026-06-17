"use client"

import type { JSX } from "react"
import { useTranslations } from "next-intl"

import { useSubmitWizard } from "@/hooks/useSubmitWizard"
import { StepIndicator } from "@/components/events/submit/StepIndicator"
import { Step1Search } from "@/components/events/submit/Step1Search"
import { Step2Details } from "@/components/events/submit/Step2Details"
import { Step3Deadlines } from "@/components/events/submit/Step3Deadlines"
import { SubmitSuccess } from "@/components/events/submit/SubmitSuccess"

// --- Component ---

// SubmitWizard is the top-level controller for the 3-step event submission
// flow. It owns no business logic itself — all state lives in useSubmitWizard.
// Browser refresh returns the user to Step 1 (state is not persisted).
function SubmitWizard(): JSX.Element {
  const t = useTranslations("submit")
  const wizard = useSubmitWizard()

  // Success screen replaces the wizard after a 201 response.
  if (wizard.submittedEvent !== null) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-8">
        <SubmitSuccess event={wizard.submittedEvent} />
      </div>
    )
  }

  const goToStep2 = (): void => wizard.goToStep(2)
  const goToStep3 = (): void => {
    if (wizard.validateStep2()) wizard.goToStep(3)
  }
  const goToStep1 = (): void => wizard.goToStep(1)
  const goToStep2Back = (): void => wizard.goToStep(2)

  return (
    <div className="mx-auto max-w-4xl px-4 py-8">
      {/* Header */}
      <div className="mb-8 flex flex-col gap-4">
        <h1 className="text-2xl font-bold">{t("addEventButton")}</h1>
        <StepIndicator currentStep={wizard.step} totalSteps={3} />
      </div>

      {/* Steps */}
      {wizard.step === 1 && (
        <Step1Search onContinue={goToStep2} />
      )}

      {wizard.step === 2 && (
        <Step2Details
          formData={wizard.formData}
          errors={wizard.errors}
          onBack={goToStep1}
          onContinue={goToStep3}
          onChange={wizard.updateField}
        />
      )}

      {wizard.step === 3 && (
        <Step3Deadlines
          deadlines={wizard.formData.deadlines}
          errors={wizard.errors}
          isSubmitting={wizard.isSubmitting}
          bannerError={wizard.bannerError}
          onBack={goToStep2Back}
          onAddDeadline={wizard.addDeadline}
          onRemoveDeadline={wizard.removeDeadline}
          onUpdateDeadline={wizard.updateDeadline}
          onSubmit={() => wizard.submit(false)}
          onSkipAndSubmit={() => wizard.submit(true)}
        />
      )}
    </div>
  )
}

// --- Export ---

export { SubmitWizard }
