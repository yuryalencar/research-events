"use client"

import { useEffect, useRef } from "react"
import type { JSX } from "react"
import Globe, { type GlobeInstance } from "globe.gl"

import { getPinColor } from "@/lib/events"
import type { EventListItem } from "@/types/api"

// --- Types ---

interface GlobeViewProps {
  events: EventListItem[]
  selectedEvent: EventListItem | null
  onPointClick: (event: EventListItem) => void
  focusPoint?: { lat: number; lng: number } | null
}

// --- Constants ---

const PIN_ALTITUDE = 0.015
const PIN_RADIUS = 0.8
const SELECTED_PIN_RADIUS = 1.2
const ROTATION_DURATION_MS = 600

// --- Component ---

// GlobeView renders the 3D globe and one pin per event, via the vanilla
// globe.gl library. This component must only ever be loaded with
// `dynamic(() => import(...), { ssr: false })` (see page.tsx) — globe.gl
// requires WebGL, which doesn't exist during server rendering.
function GlobeView({ events, selectedEvent, onPointClick, focusPoint }: GlobeViewProps): JSX.Element {
  const containerRef = useRef<HTMLDivElement>(null)
  const globeRef = useRef<GlobeInstance | null>(null)

  // Create the globe instance once, on mount.
  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const globe = new Globe(container)
      .globeImageUrl("//unpkg.com/three-globe/example/img/earth-blue-marble.jpg")
      .backgroundColor("rgba(0,0,0,0)")
      .pointAltitude(PIN_ALTITUDE)
      .width(container.clientWidth)
      .height(container.clientHeight)

    globeRef.current = globe

    const handleResize = (): void => {
      globe.width(container.clientWidth).height(container.clientHeight)
    }
    window.addEventListener("resize", handleResize)

    return () => {
      window.removeEventListener("resize", handleResize)
      globe._destructor()
      globeRef.current = null
    }
  }, [])

  // Update pins whenever events or the selected event change. Kept separate
  // from globe creation so resizing/re-creating the globe instance isn't
  // needed just to re-color a pin.
  useEffect(() => {
    const globe = globeRef.current
    if (!globe) return

    const selectedId = selectedEvent?.id ?? null

    globe
      .pointsData(events)
      .pointLat((point) => (point as EventListItem).latitude)
      .pointLng((point) => (point as EventListItem).longitude)
      .pointColor((point) => {
        const event = point as EventListItem
        return getPinColor(event.id, event.end_date, selectedId, new Date())
      })
      .pointLabel((point) => {
        const event = point as EventListItem
        return `${event.name} (${event.year})`
      })
      .pointRadius((point) => ((point as EventListItem).id === selectedId ? SELECTED_PIN_RADIUS : PIN_RADIUS))
      .onPointClick((point) => onPointClick(point as EventListItem))

    // Rotate to the selected event while preserving the user's current zoom
    // level (altitude). Only re-orient lat/lng — never snap the camera.
    if (selectedEvent !== null) {
      const { altitude } = globe.pointOfView()
      globe.pointOfView(
        { lat: selectedEvent.latitude, lng: selectedEvent.longitude, altitude },
        ROTATION_DURATION_MS,
      )
    }
  }, [events, selectedEvent, onPointClick])

  // Rotate to the first result when filters produce a new set of events.
  // Preserves the user's current zoom level — only lat/lng are changed.
  useEffect(() => {
    const globe = globeRef.current
    if (!globe || !focusPoint) return
    const { altitude } = globe.pointOfView()
    globe.pointOfView({ lat: focusPoint.lat, lng: focusPoint.lng, altitude }, ROTATION_DURATION_MS)
  }, [focusPoint])

  return <div ref={containerRef} className="h-full w-full" />
}

// --- Export ---

export { GlobeView }
