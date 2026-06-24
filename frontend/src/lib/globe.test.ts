import { describe, it, expect } from "vitest"

import { altitudeToZoom } from "./globe"

describe("altitudeToZoom", () => {
  it("returns an integer", () => {
    expect(Number.isInteger(altitudeToZoom(1.0))).toBe(true)
  })

  it("returns a value within the valid supercluster zoom range [0, 20]", () => {
    const zoom = altitudeToZoom(1.0)
    expect(zoom).toBeGreaterThanOrEqual(0)
    expect(zoom).toBeLessThanOrEqual(20)
  })

  it("returns a lower zoom for a higher altitude (more zoomed out)", () => {
    expect(altitudeToZoom(2.0)).toBeLessThan(altitudeToZoom(0.5))
  })

  it("returns a higher zoom for a lower altitude (more zoomed in)", () => {
    expect(altitudeToZoom(0.1)).toBeGreaterThan(altitudeToZoom(1.0))
  })

  it("clamps to 0 for very high altitudes (fully zoomed out)", () => {
    expect(altitudeToZoom(10)).toBe(0)
  })

  it("clamps to 20 for very low altitudes (maximum zoom)", () => {
    expect(altitudeToZoom(0.0001)).toBe(20)
  })
})
