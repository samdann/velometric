// ─── Shared chart styling constants ──────────────────────────────────────────
// All recharts components and custom charts import from here so styles can be
// changed in one place.

export const CHART_GRID_STROKE = "rgba(255,255,255,0.06)";

export const CHART_TICK_STYLE = {
  fill: "#6b7280",
  fontSize: 11,
  fontFamily: "DM Mono, monospace",
} as const;

export const CHART_TOOLTIP_CONTENT_STYLE = {
  backgroundColor: "var(--color-background-subtle)",
  border: "1px solid var(--color-border)",
  borderRadius: "6px",
  fontSize: 12,
  color: "var(--color-foreground)",
} as const;

export const CHART_COLORS = {
  power:     "#F97316",
  speed:     "#3B82F6",
  heartRate: "#EF4444",
  cadence:   "#A855F7",
  elevation: "#22C55E",
} as const;

// ─── Zone colours ─────────────────────────────────────────────────────────────

export const ZONE_ORANGE_SHADES = [
  "#FED7AA", // Z1 — lightest
  "#FDBA74", // Z2
  "#FB923C", // Z3
  "#F97316", // Z4
  "#EA580C", // Z5
  "#C2410C", // Z6
  "#9A3412", // Z7 — darkest
] as const;

export function zoneColor(zoneNumber: number, total: number): string {
  const idx = Math.round(
    ((zoneNumber - 1) / Math.max(total - 1, 1)) * (ZONE_ORANGE_SHADES.length - 1)
  );
  return ZONE_ORANGE_SHADES[Math.min(idx, ZONE_ORANGE_SHADES.length - 1)];
}

// ─── Power curve durations ────────────────────────────────────────────────────

/** Standard durations shown in the power curve table and statistics chart. */
export const POWER_CURVE_DURATIONS = [5, 15, 30, 60, 300, 600, 1200, 1800, 2700, 3600] as const;

export const DURATION_LABELS: Record<number, string> = {
  5:    "5s",
  15:   "15s",
  30:   "30s",
  60:   "1m",
  300:  "5m",
  600:  "10m",
  1200: "20m",
  1800: "30m",
  2700: "45m",
  3600: "1h",
};

export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) {
    const m = Math.floor(seconds / 60);
    const s = seconds % 60;
    return s > 0 ? `${m}m ${s}s` : `${m}m`;
  }
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  return m > 0 ? `${h}h ${m}m` : `${h}h`;
}

// ─── Zone bar chart helpers ───────────────────────────────────────────────────

export function formatZoneTime(seconds: number): string {
  if (seconds < 60) return `${Math.round(seconds)}s`;
  const m = Math.floor(seconds / 60);
  const s = Math.round(seconds % 60);
  if (m < 60) return `${m}m ${s.toString().padStart(2, "0")}s`;
  const h = Math.floor(m / 60);
  const rem = m % 60;
  return `${h}h ${rem.toString().padStart(2, "0")}m`;
}
