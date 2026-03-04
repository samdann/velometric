"use client";

import { HRZoneDistributionPoint } from "@/lib/api";

interface HRZonesChartProps {
  distribution: HRZoneDistributionPoint[];
}

function formatTime(seconds: number): string {
  if (seconds < 60) return `${Math.round(seconds)}s`;
  const m = Math.floor(seconds / 60);
  const s = Math.round(seconds % 60);
  if (m < 60) return `${m}m ${s.toString().padStart(2, "0")}s`;
  const h = Math.floor(m / 60);
  const rem = m % 60;
  return `${h}h ${rem.toString().padStart(2, "0")}m`;
}

// Shades of red: Z1 = lightest, Zn = darkest
// Pre-compute for up to 7 zones
const RED_SHADES = [
  "#FECACA", // Z1 — very light red
  "#FCA5A5", // Z2
  "#F87171", // Z3
  "#EF4444", // Z4
  "#DC2626", // Z5
  "#B91C1C", // Z6
  "#7F1D1D", // Z7
];

function zoneColor(zoneNumber: number, total: number): string {
  // Map zone 1..total onto the shade array from lightest to darkest
  const idx = Math.round(((zoneNumber - 1) / Math.max(total - 1, 1)) * (RED_SHADES.length - 1));
  return RED_SHADES[Math.min(idx, RED_SHADES.length - 1)];
}

export function HRZonesChart({ distribution }: HRZonesChartProps) {
  if (!distribution || distribution.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-background-subtle p-6 text-center">
        <p className="text-sm text-foreground-muted">
          No HR zone data — set up your HR zones in your profile and re-upload
          the activity.
        </p>
      </div>
    );
  }

  const maxPct = Math.max(...distribution.map((z) => z.percentage), 1);
  // Highest intensity on top → reverse render order
  const sorted = [...distribution].sort((a, b) => b.zone_number - a.zone_number);
  const total = distribution.length;

  return (
    <div>
      <h3 className="mb-3 text-sm font-medium text-foreground-muted">
        Heart Rate Zone Distribution
      </h3>
      <div className="rounded-lg border border-border bg-background-subtle p-6">
      <div className="space-y-2">
        {sorted.map((zone) => {
          const color = zoneColor(zone.zone_number, total);
          const barPct = (zone.percentage / maxPct) * 100;
          const bpmRange = zone.max_bpm
            ? `${zone.min_bpm}–${zone.max_bpm} bpm`
            : `≥${zone.min_bpm} bpm`;
          const hasTime = zone.seconds > 0;

          return (
            <div key={zone.zone_number} className="flex items-center gap-3">
              {/* Zone label */}
              <div className="w-32 shrink-0">
                <p className="text-xs font-semibold text-foreground">
                  <span className="mr-1.5 font-mono text-foreground-muted">
                    Z{zone.zone_number}
                  </span>
                  {zone.name}
                </p>
                <p className="font-mono text-[10px] text-foreground-muted">
                  {bpmRange}
                </p>
              </div>

              {/* Time column */}
              <div className="w-20 shrink-0 text-right">
                <p className="font-mono text-sm font-semibold text-foreground">
                  {hasTime ? formatTime(zone.seconds) : "—"}
                </p>
              </div>

              {/* Percentage column */}
              <div className="w-14 shrink-0 text-right">
                <p className="font-mono text-sm font-semibold" style={{ color }}>
                  {hasTime ? `${zone.percentage.toFixed(1)}%` : "—"}
                </p>
              </div>

              {/* Bar track */}
              <div className="relative h-6 flex-1 overflow-hidden rounded-sm bg-background">
                <div
                  className="h-full rounded-sm transition-all duration-500"
                  style={{
                    width: `${barPct}%`,
                    backgroundColor: color,
                  }}
                />
              </div>
            </div>
          );
        })}
      </div>
      </div>
    </div>
  );
}
