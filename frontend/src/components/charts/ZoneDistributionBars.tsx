"use client";

import { zoneColor, formatZoneTime } from "@/lib/chart-config";

export interface ZoneRow {
  zoneNumber: number;
  name: string;
  /** e.g. "110–150w" or "120–160 bpm" */
  rangeLabel: string;
  percentage: number;
  /** If provided, a time column is shown */
  seconds?: number;
}

interface Props {
  data: ZoneRow[];
}

export function ZoneDistributionBars({ data }: Props) {
  const maxPct = Math.max(...data.map((z) => z.percentage), 1);
  const sorted = [...data].sort((a, b) => b.zoneNumber - a.zoneNumber);
  const total = data.length;
  const showTime = data.some((z) => z.seconds != null);

  return (
    <div className="space-y-2">
      {sorted.map((zone) => {
        const color = zoneColor(zone.zoneNumber, total);
        const barPct = (zone.percentage / maxPct) * 100;
        const hasPct = zone.percentage > 0;

        return (
          <div key={zone.zoneNumber} className="flex items-center gap-3">
            <div className="w-32 shrink-0">
              <p className="text-xs font-semibold text-foreground">
                <span className="mr-1.5 font-mono text-foreground-muted">Z{zone.zoneNumber}</span>
                {zone.name}
              </p>
              <p className="font-mono text-[10px] text-foreground-muted">{zone.rangeLabel}</p>
            </div>

            {showTime && (
              <div className="w-20 shrink-0 text-right">
                <p className="font-mono text-sm font-semibold text-foreground">
                  {zone.seconds ? formatZoneTime(zone.seconds) : "—"}
                </p>
              </div>
            )}

            <div className="w-14 shrink-0 text-right">
              <p className="font-mono text-sm font-semibold" style={{ color }}>
                {hasPct ? `${zone.percentage.toFixed(1)}%` : "—"}
              </p>
            </div>

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
  );
}
