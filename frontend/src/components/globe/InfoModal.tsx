"use client"

import type { JSX } from "react"
import { useTranslations } from "next-intl"

import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog"
import { Drawer, DrawerContent, DrawerHeader, DrawerTitle, DrawerDescription } from "@/components/ui/drawer"
import { ADMIN_EMAIL, APP_VERSION, GITHUB_PROFILE_URL, GITHUB_REPO_URL } from "@/lib/constants"
import { PIN_COLOR_DEFAULT, PIN_COLOR_PAST, PIN_COLOR_SELECTED, PIN_COLOR_CLUSTER } from "@/lib/events"

// --- Types ---

interface InfoModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  isDesktop: boolean
}

interface PinLegendRowProps {
  color: string
  label: string
  description: string
}

// --- Component ---

// InfoModal shows project information — as a centered Dialog on desktop or a
// bottom Drawer on mobile, matching the responsive pattern used by
// EventDetailView (see specs/frontend/info-modal.md).
function InfoModal({ open, onOpenChange, isDesktop }: InfoModalProps): JSX.Element {
  const t = useTranslations("info")

  if (isDesktop) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-h-[85vh]">
          <DialogHeader>
            <DialogTitle>{t("title")}</DialogTitle>
            <DialogDescription className="sr-only">{t("title")}</DialogDescription>
          </DialogHeader>
          <InfoModalContent />
        </DialogContent>
      </Dialog>
    )
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle>{t("title")}</DrawerTitle>
          <DrawerDescription className="sr-only">{t("title")}</DrawerDescription>
        </DrawerHeader>
        <InfoModalContent />
      </DrawerContent>
    </Drawer>
  )
}

// --- Sub-components ---

// InfoModalContent is shared between the Dialog and Drawer layouts so both
// desktop and mobile render identical information.
function InfoModalContent(): JSX.Element {
  const t = useTranslations("info")

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-5 overflow-y-auto px-4 pb-6">
      <div className="flex items-center justify-between">
        <span className="text-sm text-muted-foreground">{t("version")}</span>
        <span className="font-mono text-sm font-medium text-foreground">{APP_VERSION}</span>
      </div>

      <hr className="border-border" />

      <div className="flex flex-col gap-3">
        <h3 className="text-sm font-semibold text-foreground">{t("legendTitle")}</h3>
        <PinLegendRow color={PIN_COLOR_DEFAULT} label={t("yellowPin.label")} description={t("yellowPin.description")} />
        <PinLegendRow color={PIN_COLOR_PAST} label={t("redPin.label")} description={t("redPin.description")} />
        <PinLegendRow color={PIN_COLOR_SELECTED} label={t("pinkPin.label")} description={t("pinkPin.description")} />
        <PinLegendRow color={PIN_COLOR_CLUSTER} label={t("violetPin.label")} description={t("violetPin.description")} />
      </div>

      <hr className="border-border" />

      <div className="flex flex-col gap-2">
        <h3 className="text-sm font-semibold text-foreground">{t("howItWorksTitle")}</h3>
        <p className="text-sm leading-relaxed text-muted-foreground">{t("howItWorksDescription")}</p>
      </div>

      <hr className="border-border" />

      <div className="flex flex-col gap-2">
        <h3 className="text-sm font-semibold text-foreground">{t("moderatorTitle")}</h3>
        <p className="text-sm text-muted-foreground">{t("moderatorDescription")}</p>
        <a href={`mailto:${ADMIN_EMAIL}`} className="text-sm font-medium text-primary underline underline-offset-2">
          {t("moderatorEmailLabel")}
        </a>
      </div>

      <hr className="border-border" />

      <div className="flex flex-col gap-2">
        <h3 className="text-sm font-semibold text-foreground">{t("authorTitle")}</h3>
        <p className="text-sm text-muted-foreground">{t("authorDescription")}</p>
        <a
          href={GITHUB_PROFILE_URL}
          target="_blank"
          rel="noopener noreferrer"
          className="text-sm font-medium text-primary underline underline-offset-2"
        >
          {t("authorLinkLabel")}
        </a>
      </div>

      <hr className="border-border" />

      <div className="flex flex-col gap-2">
        <h3 className="text-sm font-semibold text-foreground">{t("contributeTitle")}</h3>
        <p className="text-sm text-muted-foreground">{t("contributeDescription")}</p>
        <a
          href={GITHUB_REPO_URL}
          target="_blank"
          rel="noopener noreferrer"
          className="text-sm font-medium text-primary underline underline-offset-2"
        >
          {t("contributeLinkLabel")}
        </a>
      </div>
    </div>
  )
}

function PinLegendRow({ color, label, description }: PinLegendRowProps): JSX.Element {
  return (
    <div className="flex items-start gap-3">
      <div aria-hidden="true" className="mt-0.5 size-3 shrink-0 rounded-full" style={{ backgroundColor: color }} />
      <div className="flex flex-col gap-0.5">
        <span className="text-sm font-medium text-foreground">{label}</span>
        <span className="text-sm text-muted-foreground">{description}</span>
      </div>
    </div>
  )
}

// --- Export ---

export { InfoModal }
export type { InfoModalProps }
