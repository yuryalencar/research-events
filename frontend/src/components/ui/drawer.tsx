"use client"

import type { ComponentProps, JSX } from "react"
import { Drawer as DrawerPrimitive } from "vaul"
import { XIcon } from "lucide-react"

import { cn } from "@/lib/utils"

// --- Root + triggers ---
// Drawer wraps vaul, a bottom-sheet primitive built for touch gestures — used
// for the mobile detail view (slides up from the bottom) per the responsive
// layout in specs/frontend/globe-homepage.md.

function Drawer(props: ComponentProps<typeof DrawerPrimitive.Root>): JSX.Element {
  return <DrawerPrimitive.Root data-slot="drawer" {...props} />
}

function DrawerTrigger(props: ComponentProps<typeof DrawerPrimitive.Trigger>): JSX.Element {
  return <DrawerPrimitive.Trigger data-slot="drawer-trigger" {...props} />
}

function DrawerClose(props: ComponentProps<typeof DrawerPrimitive.Close>): JSX.Element {
  return <DrawerPrimitive.Close data-slot="drawer-close" {...props} />
}

function DrawerPortal(props: ComponentProps<typeof DrawerPrimitive.Portal>): JSX.Element {
  return <DrawerPrimitive.Portal data-slot="drawer-portal" {...props} />
}

// --- Overlay + content ---

function DrawerOverlay({ className, ...props }: ComponentProps<typeof DrawerPrimitive.Overlay>): JSX.Element {
  return (
    <DrawerPrimitive.Overlay
      data-slot="drawer-overlay"
      className={cn(
        "data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/50",
        className
      )}
      {...props}
    />
  )
}

function DrawerContent({ className, children, ...props }: ComponentProps<typeof DrawerPrimitive.Content>): JSX.Element {
  return (
    <DrawerPortal>
      <DrawerOverlay />
      <DrawerPrimitive.Content
        data-slot="drawer-content"
        className={cn(
          "group/drawer-content bg-background fixed z-50 flex h-auto flex-col rounded-t-lg border-t",
          "inset-x-0 bottom-0 mt-24 max-h-[80vh]",
          className
        )}
        {...props}
      >
        <div className="bg-muted mx-auto mt-4 h-2 w-[100px] shrink-0 rounded-full" />
        {children}
        <DrawerPrimitive.Close className="ring-offset-background focus:ring-ring absolute top-4 right-4 rounded-xs opacity-70 transition-opacity hover:opacity-100 focus:ring-2 focus:ring-offset-2 focus:outline-none disabled:pointer-events-none">
          <XIcon className="size-4" />
          <span className="sr-only">Close</span>
        </DrawerPrimitive.Close>
      </DrawerPrimitive.Content>
    </DrawerPortal>
  )
}

// --- Header / footer / title / description ---

function DrawerHeader({ className, ...props }: ComponentProps<"div">): JSX.Element {
  return <div data-slot="drawer-header" className={cn("flex flex-col gap-1.5 p-4", className)} {...props} />
}

function DrawerFooter({ className, ...props }: ComponentProps<"div">): JSX.Element {
  return <div data-slot="drawer-footer" className={cn("mt-auto flex flex-col gap-2 p-4", className)} {...props} />
}

function DrawerTitle({ className, ...props }: ComponentProps<typeof DrawerPrimitive.Title>): JSX.Element {
  return (
    <DrawerPrimitive.Title
      data-slot="drawer-title"
      className={cn("text-foreground font-semibold", className)}
      {...props}
    />
  )
}

function DrawerDescription({ className, ...props }: ComponentProps<typeof DrawerPrimitive.Description>): JSX.Element {
  return (
    <DrawerPrimitive.Description
      data-slot="drawer-description"
      className={cn("text-muted-foreground text-sm", className)}
      {...props}
    />
  )
}

// --- Export ---

export {
  Drawer,
  DrawerTrigger,
  DrawerClose,
  DrawerContent,
  DrawerHeader,
  DrawerFooter,
  DrawerTitle,
  DrawerDescription,
}
