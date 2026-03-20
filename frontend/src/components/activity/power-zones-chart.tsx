"use client";

import { PowerZoneDistributionPoint } from "@/lib/api";
import { ZoneDistributionBars, ZoneRow } from "@/components/charts/ZoneDistributionBars";

interface PowerZonesChartProps {
  distribution: PowerZoneDistributionPoint[];
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

  const rows: ZoneRow[] = distribution.map((z) => ({
    zoneNumber: z.zone_number,
    name: z.name,
    rangeLabel: z.max_watts ? `${z.min_watts}–${z.max_watts}w` : `≥${z.min_watts}w`,
    percentage: z.percentage,
    seconds: z.seconds,
  }));

  return (
    <div>
      <h3 className="mb-3 text-sm font-medium text-foreground-muted">
        Power Zone Distribution
      </h3>
      <div className="rounded-lg border border-border bg-background-subtle p-6">
        <ZoneDistributionBars data={rows} />
      </div>
    </div>
  );
}
