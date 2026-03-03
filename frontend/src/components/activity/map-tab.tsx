"use client";

import { useEffect, useState, useRef } from "react";
import Map, { Source, Layer, Marker, type MapRef } from "react-map-gl/maplibre";
import type { LayerSpecification } from "maplibre-gl";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from "recharts";
import { api } from "@/lib/api";
import "maplibre-gl/dist/maplibre-gl.css";

interface MapTabProps {
  activityId: string;
}

interface RoutePoint {
  lat: number;
  lon: number;
  distance?: number;
}

const routeLayer: LayerSpecification = {
  id: "route",
  type: "line",
  source: "route",
  layout: { "line-join": "round", "line-cap": "round" },
  paint: { "line-color": "#F97316", "line-width": 3, "line-opacity": 0.9 },
};

function niceStep(range: number, steps: number[]): number {
  for (const s of steps) if (range / s <= 8) return s;
  return steps[steps.length - 1];
}

function findClosestPoint(route: RoutePoint[], distance: number): RoutePoint | null {
  if (route.length === 0) return null;
  let best = route[0];
  let bestDiff = Infinity;
  for (const p of route) {
    if (p.distance == null) continue;
    const diff = Math.abs(p.distance - distance);
    if (diff < bestDiff) {
      bestDiff = diff;
      best = p;
    }
  }
  return best;
}

export function MapTab({ activityId }: MapTabProps) {
  const [route, setRoute] = useState<RoutePoint[]>([]);
  const [elevation, setElevation] = useState<{ distance: number; altitude: number }[]>([]);
  const [loading, setLoading] = useState(true);
  const [marker, setMarker] = useState<{ lat: number; lon: number } | null>(null);
  const mapRef = useRef<MapRef>(null);

  useEffect(() => {
    Promise.all([
      api.getActivityRoute(activityId),
      api.getElevationProfile(activityId),
    ])
      .then(([routePoints, elevPoints]) => {
        setRoute(routePoints);
        setElevation(elevPoints);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [activityId]);

  if (loading) {
    return (
      <div className="mt-6 flex h-32 items-center justify-center">
        <p className="text-foreground-muted">Loading map data...</p>
      </div>
    );
  }

  if (route.length === 0) {
    return (
      <div className="mt-6 rounded-lg border border-border bg-background-subtle p-6 text-center">
        <p className="text-foreground-muted">No GPS data available for this activity</p>
      </div>
    );
  }

  const coords: [number, number][] = route.map((p) => [p.lon, p.lat]);
  const lons = coords.map((c) => c[0]);
  const lats = coords.map((c) => c[1]);
  const bounds: [[number, number], [number, number]] = [
    [Math.min(...lons), Math.min(...lats)],
    [Math.max(...lons), Math.max(...lats)],
  ];

  const geojson: GeoJSON.FeatureCollection = {
    type: "FeatureCollection",
    features: [
      {
        type: "Feature",
        geometry: { type: "LineString", coordinates: coords },
        properties: {},
      },
    ],
  };

  // Elevation chart axis helpers
  const hasElevation = elevation.length > 0;
  const minAlt = hasElevation ? Math.min(...elevation.map((d) => d.altitude)) : 0;
  const maxAlt = hasElevation ? Math.max(...elevation.map((d) => d.altitude)) : 0;
  const maxDist = hasElevation ? Math.max(...elevation.map((d) => d.distance)) : 0;
  const altStep = niceStep(maxAlt - minAlt, [10, 20, 50, 100, 200, 500]);
  const altMin = Math.floor(minAlt / altStep) * altStep;
  const altMax = Math.ceil(maxAlt / altStep) * altStep;
  const altTicks: number[] = [];
  for (let t = altMin; t <= altMax; t += altStep) altTicks.push(t);
  const distStep = niceStep(maxDist, [1, 2, 5, 10, 20, 50]);
  const distTicks: number[] = [];
  for (let t = 0; t <= maxDist; t += distStep) distTicks.push(parseFloat(t.toFixed(1)));

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  function handleChartMouseMove(state: any) {
    if (!state?.activePayload?.length) return;
    const distance: number = state.activePayload[0]?.payload?.distance;
    if (distance == null) return;
    const point = findClosestPoint(route, distance);
    if (point) setMarker({ lat: point.lat, lon: point.lon });
  }

  function handleChartMouseLeave() {
    setMarker(null);
  }

  return (
    <div className="mt-6 space-y-4">
      {/* Map */}
      <div className="overflow-hidden rounded-lg border border-border" style={{ height: 460 }}>
        <Map
          ref={mapRef}
          mapStyle="https://tiles.openfreemap.org/styles/positron"
          initialViewState={{ bounds, fitBoundsOptions: { padding: 40 } }}
          style={{ width: "100%", height: "100%" }}
        >
          <Source id="route" type="geojson" data={geojson}>
            <Layer {...routeLayer} />
          </Source>
          {marker && (
            <Marker longitude={marker.lon} latitude={marker.lat} anchor="center">
              <div
                style={{
                  width: 18,
                  height: 18,
                  borderRadius: "50%",
                  backgroundColor: "#3B82F6",
                  border: "3px solid white",
                  boxShadow: "0 0 0 3px #3B82F6, 0 2px 8px rgba(0,0,0,0.4)",
                }}
              />
            </Marker>
          )}
        </Map>
      </div>

      {/* Elevation chart — hover moves the marker on the map */}
      {hasElevation && (
        <div className="rounded-lg border border-border bg-background-subtle p-4">
          <p className="mb-3 text-xs text-foreground-muted">Elevation — hover to track position on map</p>
          <ResponsiveContainer width="100%" height={140}>
            <AreaChart
              data={elevation}
              margin={{ top: 4, right: 8, left: 0, bottom: 0 }}
              onMouseMove={handleChartMouseMove}
              onMouseLeave={handleChartMouseLeave}
            >
              <defs>
                <linearGradient id="elevGradientMap" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#22C55E" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#22C55E" stopOpacity={0.02} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
              <XAxis
                dataKey="distance"
                type="number"
                domain={[0, maxDist]}
                ticks={distTicks}
                tickFormatter={(v) => `${v}km`}
                tick={{ fontSize: 10, fill: "var(--color-foreground-muted)" }}
                axisLine={false}
                tickLine={false}
              />
              <YAxis
                domain={[altMin, altMax]}
                ticks={altTicks}
                tickFormatter={(v) => `${v}m`}
                tick={{ fontSize: 10, fill: "var(--color-foreground-muted)" }}
                axisLine={false}
                tickLine={false}
                width={45}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: "var(--color-background-subtle)",
                  border: "1px solid var(--color-border)",
                  borderRadius: "6px",
                  fontSize: "12px",
                  color: "var(--color-foreground)",
                }}
                formatter={(value: number | undefined) => [`${Math.round(value ?? 0)}m`, "Elevation"]}
                labelFormatter={(v) => `${Number(v).toFixed(1)} km`}
              />
              <Area
                type="monotone"
                dataKey="altitude"
                stroke="#22C55E"
                strokeWidth={1.5}
                fill="url(#elevGradientMap)"
                dot={false}
                isAnimationActive={false}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
}
