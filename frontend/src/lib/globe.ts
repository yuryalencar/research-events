// --- Constants ---

// Chosen so that altitude 10 (fully zoomed out) maps to zoom 0 and
// altitude 0.0001 (maximum zoom) maps to zoom 20 via the formula below.
const ZOOM_SCALE = 1.2
const ZOOM_OFFSET = 4

// --- Pure helpers ---

// altitudeToZoom converts a globe.gl camera altitude (the camera distance above
// the surface as a multiple of the globe radius) to an integer supercluster
// zoom level in [0, 20]. Higher altitude = more zoomed out = lower zoom number.
function altitudeToZoom(altitude: number): number {
  const raw = Math.round(-Math.log2(altitude) * ZOOM_SCALE + ZOOM_OFFSET)
  return Math.max(0, Math.min(20, raw))
}

// --- Export ---

export { altitudeToZoom }
