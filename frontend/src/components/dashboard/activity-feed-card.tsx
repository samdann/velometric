"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { Map, Source, Layer } from "react-map-gl/maplibre";
import "maplibre-gl/dist/maplibre-gl.css";
import type { FeedActivity } from "@/lib/api";
import { MAP_STYLE, routeLayerSpec } from "@/lib/map-config";

function formatRelativeDate(iso: string): string {
  const date = new Date(iso);
  const now = new Date();
  const todayStr = now.toDateString();
  const yesterday = new Date(now);
  yesterday.setDate(now.getDate() - 1);

  const timeStr = date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });

  if (date.toDateString() === todayStr) return `Today at ${timeStr}`;
  if (date.toDateString() === yesterday.toDateString()) return `Yesterday at ${timeStr}`;

  return (
    date.toLocaleDateString([], { day: "numeric", month: "short", year: "numeric" }) +
    ` at ${timeStr}`
  );
}

function formatDuration(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  if (h > 0) return `${h}h ${m.toString().padStart(2, "0")}m`;
  return `${m}m ${s.toString().padStart(2, "0")}s`;
}


function MapPreview({ id, points }: { id: string; points: { lat: number; lon: number }[] }) {
  const [mounted, setMounted] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setMounted(true);
          observer.disconnect();
        }
      },
      { rootMargin: "200px" }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  if (points.length < 2) {
    return (
      <div className="flex h-full items-center justify-center">
        <span className="text-xs text-foreground-muted">No GPS data</span>
      </div>
    );
  }

  const lons = points.map((p) => p.lon);
  const lats = points.map((p) => p.lat);
  const bounds: [[number, number], [number, number]] = [
    [Math.min(...lons), Math.min(...lats)],
    [Math.max(...lons), Math.max(...lats)],
  ];

  const geojson = {
    type: "Feature" as const,
    geometry: {
      type: "LineString" as const,
      coordinates: points.map((p) => [p.lon, p.lat]),
    },
    properties: {},
  };

  return (
    <div ref={containerRef} style={{ width: "100%", height: "100%" }}>
      {mounted && (
        <Map
          id={`map-${id}`}
          initialViewState={{ bounds, fitBoundsOptions: { padding: 20 } }}
          style={{ width: "100%", height: "100%" }}
          mapStyle={MAP_STYLE}
          interactive={false}
          attributionControl={false}
        >
          <Source id="route" type="geojson" data={geojson}>
            <Layer {...routeLayerSpec} />
          </Source>
        </Map>
      )}
    </div>
  );
}

export function ActivityFeedCard({ activity }: { activity: FeedActivity }) {
  const meta = [
    formatRelativeDate(activity.startTime),
    activity.deviceName,
    activity.location,
  ]
    .filter(Boolean)
    .join(" · ");

  return (
    <div className="overflow-hidden border border-border bg-background-subtle transition-colors hover:border-border/80">
      {/* Card body */}
      <div className="p-5">
        {/* Row 1: user name */}
        <p className="text-xs font-medium text-foreground-muted">{activity.userName}</p>

        {/* Row 2: date · device · location */}
        <p className="mt-0.5 text-xs text-foreground-muted">{meta}</p>

        {/* Row 3: activity title */}
        <h3
          className="mt-3 font-semibold text-foreground"
          style={{ fontFamily: "var(--font-instrument-sans)" }}
        >
          <Link
            href={`/activities/${activity.id}`}
            className="hover:text-primary transition-colors"
          >
            {activity.name}
          </Link>
        </h3>

        {/* Row 4: quick stats */}
        <div className="mt-3 flex gap-6">
          <div>
            <p className="text-[10px] uppercase tracking-wider text-foreground-muted">Distance</p>
            <p className="font-mono text-sm font-semibold text-foreground">
              {activity.distanceKm.toFixed(1)}{" "}
              <span className="text-xs font-normal text-foreground-muted">km</span>
            </p>
          </div>
          <div>
            <p className="text-[10px] uppercase tracking-wider text-foreground-muted">Time</p>
            <p className="font-mono text-sm font-semibold text-foreground">
              {formatDuration(activity.durationSeconds)}
            </p>
          </div>
          <div>
            <p className="text-[10px] uppercase tracking-wider text-foreground-muted">Elevation</p>
            <p className="font-mono text-sm font-semibold text-foreground">
              {Math.round(activity.elevationGainM).toLocaleString()}{" "}
              <span className="text-xs font-normal text-foreground-muted">m</span>
            </p>
          </div>
        </div>
      </div>

      {/* Row 5: map */}
      <div className="relative border-t border-border bg-background" style={{ aspectRatio: "2 / 1" }}>
        <MapPreview id={activity.id} points={activity.route} />
      </div>
    </div>
  );
}
