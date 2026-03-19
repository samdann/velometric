"use client";

import { useEffect, useState } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { api, Activity, PowerZoneDistributionPoint } from "@/lib/api";
import { PowerZonesChart } from "./power-zones-chart";
import {
  CHART_GRID_STROKE,
  CHART_TICK_STYLE,
  CHART_TOOLTIP_CONTENT_STYLE,
  CHART_COLORS,
  POWER_CURVE_DURATIONS,
  DURATION_LABELS,
  formatDuration,
} from "@/lib/chart-config";

interface PowerCurvePoint {
  durationSeconds: number;
  bestPower: number;
  avgHeartRate?: number;
  avgSpeed?: number;
  avgGradient?: number;
  avgCadence?: number;
  avgLRBalance?: number;
  avgTorqueEffectiveness?: number;
  wattsPerKg?: number;
}

interface PowerTabProps {
  activityId: string;
  activity: Activity;
}

// Standard power curve durations to display in table
const TABLE_DURATIONS = POWER_CURVE_DURATIONS;

// Key durations for the summary cards
const KEY_DURATIONS = [5, 30, 60, 300, 1200, 3600];
const KEY_DURATION_LABELS: Record<number, string> = {
  5: "5 sec",
  30: "30 sec",
  60: "1 min",
  300: "5 min",
  1200: "20 min",
  3600: "1 hour",
};

// All durations the backend computes — used for the chart
const CHART_DURATIONS = [1, 5, 10, 15, 20, 30, 45, 60, 90, 120, 180, 300, 600, 900, 1200, 1800, 2700, 3600, 5400, 7200];

// Log-scale x-axis ticks with readable labels
const X_TICKS = [1, 5, 30, 60, 300, 600, 1800, 3600, 7200];

export function PowerTab({ activityId, activity }: PowerTabProps) {
  const [powerCurve, setPowerCurve] = useState<PowerCurvePoint[]>([]);
  const [powerZones, setPowerZones] = useState<PowerZoneDistributionPoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [curveView, setCurveView] = useState<"chart" | "table">("chart");

  useEffect(() => {
    async function fetchData() {
      try {
        const [curveData, zonesData] = await Promise.all([
          api.getPowerCurve(activityId),
          api.getPowerZoneDistribution(activityId),
        ]);
        setPowerCurve(curveData.sort((a, b) => a.durationSeconds - b.durationSeconds));
        setPowerZones(zonesData);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load power data");
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, [activityId]);

  if (loading) {
    return (
      <div className="mt-6 flex h-32 items-center justify-center">
        <p className="text-foreground-muted">Loading power data...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="mt-6 rounded-lg bg-heart-rate/10 p-4">
        <p className="text-sm text-heart-rate">{error}</p>
      </div>
    );
  }

  // Get key power values for summary cards
  const keyPowers = KEY_DURATIONS.map((duration) => {
    const point = powerCurve.find((p) => p.durationSeconds === duration);
    return {
      duration,
      label: KEY_DURATION_LABELS[duration],
      power: point?.bestPower ?? null,
    };
  }).filter((p) => p.power !== null);

  // Get table rows filtered to standard durations
  const tableRows = TABLE_DURATIONS.map((duration) => {
    return powerCurve.find((p) => p.durationSeconds === duration) ?? null;
  }).filter((p) => p !== null) as PowerCurvePoint[];

  const hasHR       = tableRows.some((r) => r.avgHeartRate != null);
  const hasWkg      = tableRows.some((r) => r.wattsPerKg != null);
  const hasSpeed    = tableRows.some((r) => r.avgSpeed != null);
  const hasGradient = tableRows.some((r) => r.avgGradient != null);
  const hasCadence  = tableRows.some((r) => r.avgCadence != null);
  const hasLR       = tableRows.some((r) => r.avgLRBalance != null);
  const hasTE       = tableRows.some((r) => r.avgTorqueEffectiveness != null);

  return (
    <div className="mt-6 space-y-6">
      {/* Key Power Outputs */}
      <div>
        <h3 className="mb-3 text-sm font-medium text-foreground-muted">Key Power Outputs</h3>
        <div className="grid grid-cols-6 gap-3">
          {activity.avgPower && (
            <div className="rounded-lg border border-border bg-background-subtle p-4">
              <p className="truncate text-xs text-foreground-muted">Avg Power</p>
              <p className="mt-1 font-mono text-base text-power">
                {activity.avgPower}<span className="ml-1 text-sm text-foreground-muted">w</span>
              </p>
            </div>
          )}
          {activity.normalizedPower && (
            <div className="rounded-lg border border-border bg-background-subtle p-4">
              <p className="truncate text-xs text-foreground-muted">Norm. Power</p>
              <p className="mt-1 font-mono text-base text-power">
                {activity.normalizedPower}<span className="ml-1 text-sm text-foreground-muted">w</span>
              </p>
            </div>
          )}
          {activity.maxPower && (
            <div className="rounded-lg border border-border bg-background-subtle p-4">
              <p className="truncate text-xs text-foreground-muted">Max Power</p>
              <p className="mt-1 font-mono text-base text-power">
                {activity.maxPower}<span className="ml-1 text-sm text-foreground-muted">w</span>
              </p>
            </div>
          )}
          {activity.intensityFactor && (
            <div className="rounded-lg border border-border bg-background-subtle p-4">
              <p className="truncate text-xs text-foreground-muted">Int. Factor</p>
              <p className="mt-1 font-mono text-base">
                {activity.intensityFactor.toFixed(2)}
              </p>
            </div>
          )}
          {activity.tss && (
            <div className="rounded-lg border border-border bg-background-subtle p-4">
              <p className="truncate text-xs text-foreground-muted">TSS</p>
              <p className="mt-1 font-mono text-base">
                {activity.tss.toFixed(0)}
              </p>
            </div>
          )}
          {activity.variabilityIndex && (
            <div className="rounded-lg border border-border bg-background-subtle p-4">
              <p className="truncate text-xs text-foreground-muted">Variability</p>
              <p className="mt-1 font-mono text-base">
                {activity.variabilityIndex.toFixed(2)}
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Power Curve — chart or table */}
      {(tableRows.length > 0 || powerCurve.length > 0) && (
        <div>
          <div className="mb-3 flex items-center justify-between">
            <h3 className="text-sm font-medium text-foreground-muted">Power Curve</h3>
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
          </div>

          <div className="relative">
            {/* Table always rendered — sets the natural height */}
            <div className={curveView === "chart" ? "invisible" : ""}>
              <div className="overflow-hidden rounded-lg border border-border">
                <table className="w-full">
                  <thead className="bg-background-subtle text-xs font-medium text-foreground-muted">
                    <tr>
                      <th className="px-4 py-2 text-left">Duration</th>
                      <th className="px-4 py-2 text-right">Power</th>
                      {hasWkg      && <th className="px-4 py-2 text-right">W/kg</th>}
                      {hasHR       && <th className="px-4 py-2 text-right">Avg HR</th>}
                      {hasSpeed    && <th className="px-4 py-2 text-right">Avg Speed</th>}
                      {hasGradient && <th className="px-4 py-2 text-right">Avg Grad</th>}
                      {hasCadence  && <th className="px-4 py-2 text-right">Avg Cad</th>}
                      {hasLR       && <th className="px-4 py-2 text-right">L/R Bal</th>}
                      {hasTE       && <th className="px-4 py-2 text-right">Torque Eff</th>}
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border text-xs">
                    {tableRows.map((point) => (
                      <tr key={point.durationSeconds} className="hover:bg-background-subtle/50">
                        <td className="px-4 py-2 font-mono">{DURATION_LABELS[point.durationSeconds]}</td>
                        <td className="px-4 py-2 text-right font-mono text-power">{point.bestPower}w</td>
                        {hasWkg      && <td className="px-4 py-2 text-right font-mono text-power">{point.wattsPerKg != null ? point.wattsPerKg.toFixed(2) : "—"}</td>}
                        {hasHR       && <td className="px-4 py-2 text-right font-mono text-heart-rate">{point.avgHeartRate != null ? `${point.avgHeartRate}` : "—"}</td>}
                        {hasSpeed    && <td className="px-4 py-2 text-right font-mono text-speed">{point.avgSpeed != null ? `${(point.avgSpeed * 3.6).toFixed(1)} km/h` : "—"}</td>}
                        {hasGradient && <td className="px-4 py-2 text-right font-mono">{point.avgGradient != null ? `${point.avgGradient.toFixed(1)}%` : "—"}</td>}
                        {hasCadence  && <td className="px-4 py-2 text-right font-mono text-cadence">{point.avgCadence != null ? `${point.avgCadence} rpm` : "—"}</td>}
                        {hasLR       && <td className="px-4 py-2 text-right font-mono">{point.avgLRBalance != null ? `${point.avgLRBalance.toFixed(1)}/${(100 - point.avgLRBalance).toFixed(1)}` : "—"}</td>}
                        {hasTE       && <td className="px-4 py-2 text-right font-mono">{point.avgTorqueEffectiveness != null ? `${point.avgTorqueEffectiveness.toFixed(1)}%` : "—"}</td>}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Chart overlays when active, fills exact same space */}
            {curveView === "chart" && (
              <div className="absolute inset-0 rounded-lg border border-border bg-background-subtle p-4">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart
                    data={powerCurve
                      .filter((p) => CHART_DURATIONS.includes(p.durationSeconds))
                      .map((p, i) => ({ ...p, index: i }))}
                    margin={{ top: 4, right: 8, left: 0, bottom: 4 }}
                  >
                    <CartesianGrid strokeDasharray="3 3" stroke={CHART_GRID_STROKE} />
                    <XAxis
                      dataKey="index"
                      type="number"
                      domain={[0, CHART_DURATIONS.length - 1]}
                      ticks={CHART_DURATIONS.map((_, i) => i)}
                      tickFormatter={(i) => formatDuration(CHART_DURATIONS[i])}
                      tick={CHART_TICK_STYLE}
                      axisLine={false}
                      tickLine={false}
                      interval={1}
                    />
                    <YAxis
                      tickFormatter={(v) => `${v}w`}
                      tick={CHART_TICK_STYLE}
                      axisLine={false}
                      tickLine={false}
                      width={52}
                    />
                    <Tooltip
                      contentStyle={CHART_TOOLTIP_CONTENT_STYLE}
                      labelFormatter={(i) => formatDuration(CHART_DURATIONS[Number(i)])}
                      formatter={(value: number | undefined) => [`${value ?? "—"}w`, "Best Power"]}
                    />
                    <Line
                      type="monotone"
                      dataKey="bestPower"
                      stroke={CHART_COLORS.power}
                      strokeWidth={2}
                      dot={false}
                      activeDot={{ r: 4, fill: CHART_COLORS.power }}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Power Zone Distribution */}
      <PowerZonesChart distribution={powerZones} />
    </div>
  );
}
