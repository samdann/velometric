"use client";

import { AnnualZoneDistributionPoint } from "@/lib/api";
import { ZoneDistributionBars, ZoneRow } from "@/components/charts/ZoneDistributionBars";
import { HelpIcon } from "@/components/ui/HelpIcon";
import { STATISTICS_HINTS } from "@/lib/statistics-hints";

interface Props {
  data: AnnualZoneDistributionPoint[];
  year: number;
  mode: "avg" | "best";
}

export function PowerDistributionWidget({ data, year, mode }: Props) {
  const hint = mode === "avg" ? STATISTICS_HINTS.zoneDistributionAvg : STATISTICS_HINTS.zoneDistributionBest;

  const rows: ZoneRow[] = data.map((z) => ({
    zoneNumber: z.zoneNumber,
    name: z.name,
    rangeLabel: z.maxWatts ? `${z.minWatts}–${z.maxWatts}w` : `≥${z.minWatts}w`,
    percentage: z.medianPercentage,
  }));

  return (
    <div className="rounded-xl border border-border bg-background-subtle p-5">
      {/* Header: title + hint */}
      <div className="flex items-center gap-2 mb-4">
        <h3 className="text-sm font-medium text-foreground-muted">Zone Distribution</h3>
        <HelpIcon hint={hint} />
      </div>

      {data.length === 0 ? (
        <p className="text-sm text-foreground-muted text-center py-6">
          No zone data for {year}. Set your FTP and power zones in Settings.
        </p>
      ) : (
        <ZoneDistributionBars data={rows} />
      )}
    </div>
  );
}
