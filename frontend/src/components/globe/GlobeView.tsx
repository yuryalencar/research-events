"use client"

import { useEffect, useRef } from "react"
import type { JSX } from "react"
import Globe, { type GlobeInstance } from "globe.gl"

import { getPinColor, getClusterPinRadius, PIN_COLOR_CLUSTER } from "@/lib/events"
import { altitudeToZoom } from "@/lib/globe"
import { isCluster } from "@/hooks/useGlobeClusters"
import type { GlobePoint } from "@/hooks/useGlobeClusters"
import type { EventListItem } from "@/types/api"

// --- Types ---

interface GlobeViewProps {
  globePoints: GlobePoint[]
  selectedEvent: EventListItem | null
  onPointClick: (event: EventListItem) => void
  onClusterClick: (clusterId: number) => void
  onZoomChange: (zoom: number) => void
  focusPoint?: { lat: number; lng: number } | null
}

// --- Constants ---

const PIN_ALTITUDE = 0.015
const PIN_RADIUS = 0.8
const SELECTED_PIN_RADIUS = 1.2
const ROTATION_DURATION_MS = 600
// Debounce delay for zoom events — avoids re-clustering on every animation
// frame during a drag gesture while still feeling responsive when zoom settles.
const ZOOM_DEBOUNCE_MS = 150

// --- Component ---

// GlobeView renders the 3D globe with individual and cluster pins via the
// vanilla globe.gl library. This component must only ever be loaded with
// `dynamic(() => import(...), { ssr: false })` (see page.tsx) — globe.gl
// requires WebGL, which doesn't exist during server rendering.
function GlobeView({
  globePoints,
  selectedEvent,
  onPointClick,
  onClusterClick,
  onZoomChange,
  focusPoint,
}: GlobeViewProps): JSX.Element {
  const containerRef = useRef<HTMLDivElement>(null)
  const globeRef = useRef<GlobeInstance | null>(null)

  // Keep callbacks current without re-registering the controls listener on
  // every render. The controls listener is attached once on mount; using a ref
  // ensures its closure always calls the latest prop value.
  const onZoomChangeRef = useRef(onZoomChange)
  onZoomChangeRef.current = onZoomChange

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

    // Emit the initial zoom so page.tsx primes the cluster index before the
    // user interacts. globe.gl defaults to altitude 2.5 on creation.
    onZoomChangeRef.current(altitudeToZoom(globe.pointOfView().altitude))

    // After the debounce settles, report the derived zoom to page.tsx so
    // useGlobeClusters can recompute which pins are clusters vs individuals.
    let zoomTimer: ReturnType<typeof setTimeout> | null = null
    const handleCameraChange = (): void => {
      if (zoomTimer) clearTimeout(zoomTimer)
      zoomTimer = setTimeout(() => {
        const { altitude } = globe.pointOfView()
        onZoomChangeRef.current(altitudeToZoom(altitude))
      }, ZOOM_DEBOUNCE_MS)
    }
    globe.controls().addEventListener("change", handleCameraChange)

    return () => {
      window.removeEventListener("resize", handleResize)
      globe.controls().removeEventListener("change", handleCameraChange)
      if (zoomTimer) clearTimeout(zoomTimer)
      globe._destructor()
      globeRef.current = null
    }
  }, [])

  // Update pins whenever globePoints or the selected event change. Kept
  // separate from globe creation so zoom changes don't recreate the globe.
  useEffect(() => {
    const globe = globeRef.current
    if (!globe) return

    const selectedId = selectedEvent?.id ?? null

    globe
      .pointsData(globePoints)
      // latitude/longitude exist on both union members — no narrowing needed.
      .pointLat((p) => (p as GlobePoint).latitude)
      .pointLng((p) => (p as GlobePoint).longitude)
      .pointColor((p) => {
        const point = p as GlobePoint
        if (isCluster(point)) return PIN_COLOR_CLUSTER
        return getPinColor(point.id, point.end_date, selectedId, new Date())
      })
      .pointLabel((p) => {
        const point = p as GlobePoint
        if (isCluster(point)) return `${point.count} events`
        return `${point.name} (${point.year})`
      })
      .pointRadius((p) => {
        const point = p as GlobePoint
        if (isCluster(point)) return getClusterPinRadius(point.count)
        return point.id === selectedId ? SELECTED_PIN_RADIUS : PIN_RADIUS
      })
      .onPointClick((p) => {
        const point = p as GlobePoint
        if (isCluster(point)) {
          onClusterClick(point.clusterId)
          return
        }
        onPointClick(point)
      })

    // Rotate to the selected event while preserving the user's current zoom
    // level (altitude). Only re-orient lat/lng — never snap the camera.
    if (selectedEvent !== null) {
      const { altitude } = globe.pointOfView()
      globe.pointOfView(
        { lat: selectedEvent.latitude, lng: selectedEvent.longitude, altitude },
        ROTATION_DURATION_MS,
      )
    }
  }, [globePoints, selectedEvent, onPointClick, onClusterClick])

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
