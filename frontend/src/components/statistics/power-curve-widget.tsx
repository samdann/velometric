"use client";

import { useState } from "react";
import { AnnualPowerCurvePoint } from "@/lib/api";
import { PowerCurveChart, PowerCurveDataPoint } from "@/components/charts/PowerCurveChart";
import { HelpIcon } from "@/components/ui/HelpIcon";
import { STATISTICS_HINTS } from "@/lib/statistics-hints";
import { POWER_CURVE_DURATIONS, DURATION_LABELS } from "@/lib/chart-config";

interface Props {
  data: AnnualPowerCurvePoint[];
  year: number;
  mode: "avg" | "best";
}

export function PowerCurveWidget({ data, year, mode }: Props) {
  const [curveView, setCurveView] = useState<"chart" | "table">("chart");

  const hint = mode === "avg" ? STATISTICS_HINTS.powerCurveAvg : STATISTICS_HINTS.powerCurveBest;
  const tooltipLabel = mode === "avg" ? "Median Power" : "Best Power";

  const tableRows = POWER_CURVE_DURATIONS
    .map((dur) => data.find((p) => p.durationSeconds === dur) ?? null)
    .filter(Boolean) as AnnualPowerCurvePoint[];

  const chartData: PowerCurveDataPoint[] = tableRows.map((p) => ({
    durationSeconds: p.durationSeconds,
    power: p.medianPower,
  }));

  return (
    <div className="rounded-xl border border-border bg-background-subtle p-5">
      {/* Header: title + hint left, toggle right */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <h3 className="text-sm font-medium text-foreground-muted">Power Curve</h3>
          <HelpIcon hint={hint} />
        </div>
        {data.length > 0 && (
          <div className="flex rounded-md border border-border text-xs font-medium overflow-hidden">
            <button
              onClick={() => setCurveView("chart")}
              className={`px-3 py-1 transition-colors ${curveView === "chart" ? "bg-power text-white" : "text-foreground-muted hover:text-foreground"}`}
            >
              Chart
            </button>
            <button
              onClick={() => setCurveView("table")}
              className={`px-3 py-1 transition-colors ${curveView === "table" ? "bg-power text-white" : "text-foreground-muted hover:text-foreground"}`}
            >
              Table
            </button>
          </div>
        )}
      </div>

      {data.length === 0 ? (
        <p className="text-sm text-foreground-muted text-center py-6">
          No power curve data for {year}
        </p>
      ) : (
        <div className="relative h-64">
          {/* Table fills fixed height with scroll if needed */}
          <div className={`h-full overflow-y-auto ${curveView === "chart" ? "invisible" : ""}`}>
            <div className="overflow-hidden rounded-lg border border-border">
              <table className="w-full">
                <thead className="bg-background-subtle text-xs font-medium text-foreground-muted">
                  <tr>
                    <th className="px-4 py-2 text-left">Duration</th>
                    <th className="px-4 py-2 text-right">Power</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border text-xs">
                  {tableRows.map((point) => (
                    <tr key={point.durationSeconds} className="hover:bg-background-subtle/50">
                      <td className="px-4 py-2 font-mono">{DURATION_LABELS[point.durationSeconds]}</td>
                      <td className="px-4 py-2 text-right font-mono text-power">{point.medianPower}w</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Chart overlays when active, fills exact same space */}
          {curveView === "chart" && (
            <div className="absolute inset-0 rounded-lg border border-border bg-background-subtle p-4">
              <PowerCurveChart data={chartData} tooltipLabel={tooltipLabel} />
            </div>
          )}
        </div>
      )}
    </div>
  );
}
