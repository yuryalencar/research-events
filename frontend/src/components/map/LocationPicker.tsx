"use client"

import "leaflet/dist/leaflet.css"
import { useEffect, useRef } from "react"
import type { JSX } from "react"
import type { Map, Marker } from "leaflet"

import { useMediaQuery } from "@/hooks/useMediaQuery"
import { COUNTRY_COORDINATES } from "@/lib/countryCoordinates"

// --- Types ---

interface LocationPickerProps {
  latitude: number | null
  longitude: number | null
  country: string | null
  onChange: (lat: number, lng: number) => void
}

// --- Constants ---

const DESKTOP_QUERY = "(min-width: 768px)"

// --- Component ---

// LocationPicker renders a Leaflet map where the user drops a draggable pin
// to set the event's lat/lng. Must be loaded via dynamic() with ssr:false —
// Leaflet requires browser APIs (window, document) that do not exist on the
// server. The onChange callback is stored in a ref so the Leaflet event
// handlers always call the latest version without needing to re-attach.
function LocationPicker({ latitude, longitude, country, onChange }: LocationPickerProps): JSX.Element {
  const isDesktop = useMediaQuery(DESKTOP_QUERY)
  const containerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<Map | null>(null)
  const markerRef = useRef<Marker | null>(null)

  // Keep callback ref current so Leaflet handlers never capture a stale closure.
  const onChangeRef = useRef(onChange)
  useEffect(() => {
    onChangeRef.current = onChange
  }, [onChange])

  // Initialise the map once on mount. The returned cleanup removes it so
  // React StrictMode's double-mount does not create two overlapping maps.
  //
  // The `cancelled` flag solves a subtle async timing bug: `import("leaflet")`
  // is a promise, so the cleanup function often runs BEFORE it resolves (e.g.
  // in Strict Mode). At that point `map` is still undefined, so `map?.remove()`
  // is a no-op and the container stays "live" from Leaflet's perspective. When
  // the second mount then triggers another import and both promises resolve,
  // both try to call `L.map(container)` → "Map container is already initialized".
  // Setting `cancelled = true` in cleanup and checking it inside the promise
  // prevents the first (stale) callback from touching the DOM at all.
  useEffect(() => {
    if (!containerRef.current || mapRef.current) return

    let cancelled = false
    let map: Map

    import("leaflet").then((L) => {
      if (cancelled || !containerRef.current) return

      // Fix Leaflet's default icon URLs which break under Next.js module
      // bundling because webpack renames the image assets.
      const icons = L.Icon.Default.prototype as unknown as Record<string, unknown>
      delete icons._getIconUrl
      L.Icon.Default.mergeOptions({
        iconRetinaUrl: "https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png",
        iconUrl: "https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png",
        shadowUrl: "https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png",
      })

      map = L.map(containerRef.current).setView([20, 0], 2)
      mapRef.current = map

      L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
        attribution:
          '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
      }).addTo(map)

      map.on("click", (e) => {
        const { lat, lng } = e.latlng

        if (markerRef.current) {
          markerRef.current.setLatLng([lat, lng])
        } else {
          const m = L.marker([lat, lng], { draggable: true }).addTo(map)
          m.on("dragend", () => {
            const pos = m.getLatLng()
            onChangeRef.current(pos.lat, pos.lng)
          })
          markerRef.current = m
        }

        onChangeRef.current(lat, lng)
      })
    })

    return (): void => {
      cancelled = true
      map?.remove()
      mapRef.current = null
      markerRef.current = null
    }
  }, [])

  // Pan to the selected country's approximate centre whenever it changes.
  useEffect(() => {
    if (!mapRef.current || !country) return
    const coords = COUNTRY_COORDINATES[country]
    if (coords) {
      mapRef.current.setView(coords, 5)
    }
  }, [country])

  // If lat/lng are set externally (e.g. on Back navigation), place the marker.
  useEffect(() => {
    if (!mapRef.current || latitude === null || longitude === null) return

    import("leaflet").then((L) => {
      if (!mapRef.current) return
      if (markerRef.current) {
        markerRef.current.setLatLng([latitude, longitude])
      } else {
        const m = L.marker([latitude, longitude], { draggable: true }).addTo(mapRef.current)
        m.on("dragend", () => {
          const pos = m.getLatLng()
          onChangeRef.current(pos.lat, pos.lng)
        })
        markerRef.current = m
      }
    })
  }, [latitude, longitude])

  const height = isDesktop ? 300 : 220

  return (
    <div
      ref={containerRef}
      style={{ height }}
      className="w-full rounded-md border border-border"
      aria-label="Map for dropping event location pin"
    />
  )
}

// --- Export ---

export { LocationPicker }
export type { LocationPickerProps }
