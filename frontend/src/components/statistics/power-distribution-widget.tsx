"use client";

import { AnnualZoneDistributionPoint } from "@/lib/api";
import { zoneColor } from "@/lib/chart-config";

interface Props {
  data: AnnualZoneDistributionPoint[];
  year: number;
}

export function PowerDistributionWidget({ data, year }: Props) {
  if (data.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-background-subtle p-6 text-center">
        <p className="text-sm text-foreground-muted">
          No zone data for {year}. Set your FTP and power zones in Settings.
        </p>
      </div>
    );
  }

  const maxPct = Math.max(...data.map((z) => z.medianPercentage), 1);
  const sorted = [...data].sort((a, b) => b.zoneNumber - a.zoneNumber);
  const total = data.length;

  return (
    <div className="rounded-lg border border-border bg-background-subtle p-6">
      <div className="space-y-2">
        {sorted.map((zone) => {
          const color = zoneColor(zone.zoneNumber, total);
          const barPct = (zone.medianPercentage / maxPct) * 100;
          const wattRange = zone.maxWatts
            ? `${zone.minWatts}–${zone.maxWatts}w`
            : `≥${zone.minWatts}w`;
          const hasPct = zone.medianPercentage > 0;

          return (
            <div key={zone.zoneNumber} className="flex items-center gap-3">
              {/* Zone label */}
              <div className="w-32 shrink-0">
                <p className="text-xs font-semibold text-foreground">
                  <span className="mr-1.5 font-mono text-foreground-muted">
                    Z{zone.zoneNumber}
                  </span>
                  {zone.name}
                </p>
                <p className="font-mono text-[10px] text-foreground-muted">{wattRange}</p>
              </div>

              {/* Percentage column */}
              <div className="w-14 shrink-0 text-right">
                <p className="font-mono text-sm font-semibold" style={{ color }}>
                  {hasPct ? `${zone.medianPercentage.toFixed(1)}%` : "—"}
                </p>
              </div>

              {/* Bar track */}
              <div className="relative h-6 flex-1 overflow-hidden rounded-sm bg-background">
                <div
                  className="h-full rounded-sm transition-all duration-500"
                  style={{ width: `${barPct}%`, backgroundColor: color }}
                />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
