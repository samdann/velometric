"use client";

import { PowerZoneDistributionPoint } from "@/lib/api";

interface PowerZonesChartProps {
  distribution: PowerZoneDistributionPoint[];
}

const ORANGE_SHADES = [
  "#FED7AA", // Z1 — lightest
  "#FDBA74", // Z2
  "#FB923C", // Z3
  "#F97316", // Z4
  "#EA580C", // Z5
  "#C2410C", // Z6
  "#9A3412", // Z7 — darkest
];

function zoneColor(zoneNumber: number, total: number): string {
  const idx = Math.round(((zoneNumber - 1) / Math.max(total - 1, 1)) * (ORANGE_SHADES.length - 1));
  return ORANGE_SHADES[Math.min(idx, ORANGE_SHADES.length - 1)];
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

export function PowerZonesChart({ distribution }: PowerZonesChartProps) {
  if (!distribution || distribution.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-background-subtle p-6 text-center">
        <p className="text-sm text-foreground-muted">
          No power zone data — set your FTP and power zones in your profile and re-upload the activity.
        </p>
      </div>
    );
  }

  const maxPct = Math.max(...distribution.map((z) => z.percentage), 1);
  const sorted = [...distribution].sort((a, b) => b.zone_number - a.zone_number);
  const total = distribution.length;

  return (
    <div>
      <h3 className="mb-3 text-sm font-medium text-foreground-muted">
        Power Zone Distribution
      </h3>
      <div className="rounded-lg border border-border bg-background-subtle p-6">
        <div className="space-y-2">
          {sorted.map((zone) => {
            const color = zoneColor(zone.zone_number, total);
            const barPct = (zone.percentage / maxPct) * 100;
            const wattRange = zone.max_watts
              ? `${zone.min_watts}–${zone.max_watts}w`
              : `≥${zone.min_watts}w`;
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
                    {wattRange}
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
