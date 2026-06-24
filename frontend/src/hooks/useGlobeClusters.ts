import { useMemo, useCallback } from "react"
import Supercluster from "supercluster"
import type { PointFeature, ClusterFeature } from "supercluster"

import type { EventListItem } from "@/types/api"

// --- Types ---

// ClusterPoint represents a group of geographically close events collapsed
// into a single pin. The clusterId is supercluster's internal id, needed to
// retrieve the full event list via getClusterEvents.
interface ClusterPoint {
  type: "cluster"
  clusterId: number
  count: number
  latitude: number
  longitude: number
}

// GlobePoint is what GlobeView renders: either a cluster pin or an individual
// event. The discriminated union lets accessor functions branch cleanly on type.
type GlobePoint = EventListItem | ClusterPoint

// isCluster narrows a GlobePoint to ClusterPoint. EventListItem has no `type`
// field, so the presence of `type === "cluster"` is a reliable discriminant.
function isCluster(point: GlobePoint): point is ClusterPoint {
  return "type" in point && (point as ClusterPoint).type === "cluster"
}

// --- Internal supercluster types ---

type PointProps = { event: EventListItem }
type ClusterProps = Record<string, never>
type SuperclusterFeature = ClusterFeature<ClusterProps> | PointFeature<PointProps>

// --- Constants ---

const CLUSTER_RADIUS = 40
const CLUSTER_MAX_ZOOM = 16
// Full-world bounding box for supercluster queries. The latitude is capped at
// ±85° (not ±90°) because supercluster uses the Web Mercator projection, which
// cannot represent the poles.
const WORLD_BBOX: [number, number, number, number] = [-180, -85, 180, 85]

// --- Private helpers ---

// toGeoJSONFeatures converts EventListItem[] to the GeoJSON point format that
// supercluster.load() expects. The full event is stored in properties so it
// can be recovered after clustering without a secondary lookup.
function toGeoJSONFeatures(events: EventListItem[]): Array<PointFeature<PointProps>> {
  return events.map((event) => ({
    type: "Feature",
    geometry: { type: "Point", coordinates: [event.longitude, event.latitude] },
    properties: { event },
  }))
}

// toGlobePoints converts the raw supercluster output into the GlobePoint union
// type that GlobeView consumes. Cluster features become ClusterPoints; point
// features are unwrapped back to their original EventListItem.
function toGlobePoints(features: SuperclusterFeature[]): GlobePoint[] {
  return features.map((feature) => {
    if ("cluster" in feature.properties && feature.properties.cluster) {
      const cluster = feature as ClusterFeature<ClusterProps>
      return {
        type: "cluster" as const,
        clusterId: cluster.properties.cluster_id,
        count: cluster.properties.point_count,
        latitude: cluster.geometry.coordinates[1],
        longitude: cluster.geometry.coordinates[0],
      }
    }
    return (feature as PointFeature<PointProps>).properties.event
  })
}

// --- Hook ---

interface UseGlobeClustersReturn {
  globePoints: GlobePoint[]
  getClusterEvents: (clusterId: number) => EventListItem[]
}

// useGlobeClusters groups events into cluster pins using supercluster, then
// returns a mixed array of ClusterPoints and individual EventListItems ready
// for GlobeView to render. Re-clusters automatically when events or zoom change.
function useGlobeClusters(events: EventListItem[], zoom: number): UseGlobeClustersReturn {
  // Rebuild the supercluster index only when the events array changes (filter
  // applied). Loading is O(n log n) — more expensive than getClusters queries.
  const clusterIndex = useMemo(() => {
    const index = new Supercluster<PointProps, ClusterProps>({
      radius: CLUSTER_RADIUS,
      maxZoom: CLUSTER_MAX_ZOOM,
    })
    index.load(toGeoJSONFeatures(events))
    return index
  }, [events])

  // Re-query the index on every zoom change — getClusters is O(log n) and cheap.
  const globePoints = useMemo(() => {
    return toGlobePoints(clusterIndex.getClusters(WORLD_BBOX, zoom) as SuperclusterFeature[])
  }, [clusterIndex, zoom])

  // getClusterEvents retrieves all original EventListItems inside a cluster.
  // Passing Infinity as the limit returns every leaf (no pagination).
  const getClusterEvents = useCallback(
    (clusterId: number): EventListItem[] => {
      // getLeaves always returns PointFeature<PointProps> — no cast needed.
      return clusterIndex.getLeaves(clusterId, Infinity).map((f) => f.properties.event)
    },
    [clusterIndex],
  )

  return { globePoints, getClusterEvents }
}

// --- Export ---

export { useGlobeClusters, isCluster }
export type { ClusterPoint, GlobePoint, UseGlobeClustersReturn }
