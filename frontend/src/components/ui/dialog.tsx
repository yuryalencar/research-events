"use client"

import type { ComponentProps, JSX } from "react"
import * as DialogPrimitive from "@radix-ui/react-dialog"
import { XIcon } from "lucide-react"

import { cn } from "@/lib/utils"

// --- Root + triggers ---
// Dialog wraps Radix's Dialog primitive as a centered overlay modal —
// distinct from Sheet, which slides in from a screen edge.

function Dialog(props: ComponentProps<typeof DialogPrimitive.Root>): JSX.Element {
  return <DialogPrimitive.Root data-slot="dialog" {...props} />
}

function DialogTrigger(props: ComponentProps<typeof DialogPrimitive.Trigger>): JSX.Element {
  return <DialogPrimitive.Trigger data-slot="dialog-trigger" {...props} />
}

function DialogPortal(props: ComponentProps<typeof DialogPrimitive.Portal>): JSX.Element {
  return <DialogPrimitive.Portal data-slot="dialog-portal" {...props} />
}

function DialogClose(props: ComponentProps<typeof DialogPrimitive.Close>): JSX.Element {
  return <DialogPrimitive.Close data-slot="dialog-close" {...props} />
}

// --- Overlay + content ---

function DialogOverlay({ className, ...props }: ComponentProps<typeof DialogPrimitive.Overlay>): JSX.Element {
  return (
    <DialogPrimitive.Overlay
      data-slot="dialog-overlay"
      className={cn(
        "data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/50",
        className
      )}
      {...props}
    />
  )
}

function DialogContent({ className, children, ...props }: ComponentProps<typeof DialogPrimitive.Content>): JSX.Element {
  return (
    <DialogPortal>
      <DialogOverlay />
      <DialogPrimitive.Content
        data-slot="dialog-content"
        className={cn(
          "bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 fixed top-1/2 left-1/2 z-50 flex w-full max-w-lg -translate-x-1/2 -translate-y-1/2 flex-col gap-4 rounded-lg border shadow-lg duration-200",
          className
        )}
        {...props}
      >
        {children}
        <DialogPrimitive.Close className="ring-offset-background focus:ring-ring absolute top-4 right-4 rounded-xs opacity-70 transition-opacity hover:opacity-100 focus:ring-2 focus:ring-offset-2 focus:outline-none disabled:pointer-events-none">
          <XIcon className="size-4" />
          <span className="sr-only">Close</span>
        </DialogPrimitive.Close>
      </DialogPrimitive.Content>
    </DialogPortal>
  )
}

// --- Header / title / description ---

function DialogHeader({ className, ...props }: ComponentProps<"div">): JSX.Element {
  return <div data-slot="dialog-header" className={cn("flex flex-col gap-1.5 p-4", className)} {...props} />
}

function DialogTitle({ className, ...props }: ComponentProps<typeof DialogPrimitive.Title>): JSX.Element {
  return (
    <DialogPrimitive.Title
      data-slot="dialog-title"
      className={cn("text-foreground font-semibold", className)}
      {...props}
    />
  )
}

function DialogDescription({ className, ...props }: ComponentProps<typeof DialogPrimitive.Description>): JSX.Element {
  return (
    <DialogPrimitive.Description
      data-slot="dialog-description"
      className={cn("text-muted-foreground text-sm", className)}
      {...props}
    />
  )
}

// --- Export ---

export {
  Dialog,
  DialogTrigger,
  DialogPortal,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
}
