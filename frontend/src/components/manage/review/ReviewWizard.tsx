"use client"

import type { JSX } from "react"
import { useLocale, useTranslations } from "next-intl"
import { useRouter } from "next/navigation"

import { useReviewWizard } from "@/hooks/useReviewWizard"
import { ManageHeader } from "@/components/manage/ManageHeader"
import { StepIndicator } from "@/components/events/submit/StepIndicator"
import { ReviewStep1Details } from "./ReviewStep1Details"
import { ReviewStep2Decision } from "./ReviewStep2Decision"
import { ReviewStep3Deadlines } from "./ReviewStep3Deadlines"
import { ReviewSuccess } from "./ReviewSuccess"
import type { SessionUser, ReviewStep } from "@/hooks/useReviewWizard"
import type { EventListItem } from "@/types/api"

// WizardStep is 1|2|3 — identical shape to ReviewStep, compatible by structure.
type WizardStep = 1 | 2 | 3

// --- Types ---

interface ReviewWizardProps {
  event: EventListItem
  user: SessionUser
}

// --- Component ---

function ReviewWizard({ event, user }: ReviewWizardProps): JSX.Element {
  const t = useTranslations("manage.review")
  const locale = useLocale()
  const router = useRouter()

  const wizard = useReviewWizard(event, user)

  function handleSignOut(): void {
    localStorage.removeItem("manage_user")
    router.replace(`/${locale}/manage`)
  }

  function handleBackToDashboard(): void {
    router.push(`/${locale}/manage/${user.role}`)
  }

  function handleManageDeadlines(): void {
    wizard.clearSuccess()
    wizard.goToStep(3)
  }

  const stepLabel = t("stepIndicator", { current: wizard.step, total: 3 })

  // The event used for Step 3 is the updated one from the PATCH response when
  // available (e.g. after an approve), otherwise the original passed in as prop.
  const step3Event = wizard.successState?.event ?? event

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <ManageHeader userName={user.name} userRole={user.role} onSignOut={handleSignOut} />

      <main className="pb-8 pt-9 sm:pt-11">
        <div className="mx-auto flex w-full max-w-4xl flex-col gap-8 px-4 sm:px-6">

          {wizard.successState ? (
            <ReviewSuccess
              state={wizard.successState}
              onManageDeadlines={handleManageDeadlines}
              onBackToDashboard={handleBackToDashboard}
            />
          ) : (
            <>
              <StepIndicator
                currentStep={wizard.step as WizardStep}
                totalSteps={3}
                label={stepLabel}
              />

              {wizard.step === 1 && (
                <ReviewStep1Details
                  formData={wizard.formData}
                  errors={wizard.errors}
                  status={wizard.status}
                  onBack={handleBackToDashboard}
                  onNext={() => {
                    if (wizard.validateStep1()) wizard.goToStep(2)
                  }}
                  onChange={wizard.updateField}
                />
              )}

              {wizard.step === 2 && (
                <ReviewStep2Decision
                  status={wizard.status}
                  isSubmitting={wizard.isSubmitting}
                  bannerError={wizard.bannerError}
                  onBack={() => wizard.goToStep(1)}
                  onApprove={wizard.approve}
                  onReject={wizard.reject}
                  onReviewDeadlines={() => wizard.goToStep(3)}
                />
              )}

              {wizard.step === 3 && (
                <ReviewStep3Deadlines
                  event={step3Event}
                  contributor={{ name: user.name, email: user.email }}
                  onBack={() => wizard.goToStep(2)}
                  onDone={handleBackToDashboard}
                />
              )}
            </>
          )}

        </div>
      </main>
    </div>
  )
}

// --- Export ---

export { ReviewWizard }
export type { ReviewWizardProps }
