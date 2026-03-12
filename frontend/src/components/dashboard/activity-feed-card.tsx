"use client";

import type { FeedActivity } from "@/lib/api";

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

function RoutePreview({ points }: { points: { lat: number; lon: number }[] }) {
  if (points.length < 2) {
    return (
      <div className="flex h-full items-center justify-center">
        <span className="text-xs text-foreground-muted">No GPS data</span>
      </div>
    );
  }

  const W = 500;
  const H = 250;
  const PAD = 16;

  const lons = points.map((p) => p.lon);
  const lats = points.map((p) => p.lat);
  const minLon = Math.min(...lons);
  const maxLon = Math.max(...lons);
  const minLat = Math.min(...lats);
  const maxLat = Math.max(...lats);
  const dLon = maxLon - minLon || 1e-6;
  const dLat = maxLat - minLat || 1e-6;

  const toX = (lon: number) => PAD + ((lon - minLon) / dLon) * (W - 2 * PAD);
  const toY = (lat: number) => H - PAD - ((lat - minLat) / dLat) * (H - 2 * PAD);

  const d = points
    .map((p, i) => `${i === 0 ? "M" : "L"}${toX(p.lon).toFixed(1)},${toY(p.lat).toFixed(1)}`)
    .join(" ");

  return (
    <svg
      viewBox={`0 0 ${W} ${H}`}
      preserveAspectRatio="xMidYMid meet"
      style={{ width: "100%", height: "100%", display: "block" }}
    >
      <path d={d} fill="none" stroke="#F97316" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export function ActivityFeedCard({ activity }: { activity: FeedActivity }) {
  const meta = [
    formatRelativeDate(activity.startTime),
    activity.deviceName,
    activity.location,
  ].filter(Boolean).join(" · ");

  return (
    <div className="overflow-hidden border border-border bg-background-subtle transition-colors hover:border-border/80">
      {/* Card body */}
      <div className="p-5">
        {/* Row 1: user name */}
        <p className="text-xs font-medium text-foreground-muted">{activity.userName}</p>

        {/* Row 2: date · device · location */}
        <p className="mt-0.5 text-xs text-foreground-muted">{meta}</p>

        {/* Row 3: activity title */}
        <h3 className="mt-3 font-semibold text-foreground" style={{ fontFamily: "var(--font-instrument-sans)" }}>
          {activity.name}
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
        <RoutePreview points={activity.route} />
      </div>
    </div>
  );
}
