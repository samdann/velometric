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
import { api, Activity } from "@/lib/api";

interface PowerCurvePoint {
  durationSeconds: number;
  bestPower: number;
  avgHeartRate?: number;
}

interface PowerTabProps {
  activityId: string;
  activity: Activity;
}

function formatDuration(seconds: number): string {
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

// Standard power curve durations to display in table
const TABLE_DURATIONS = [5, 15, 30, 60, 300, 600, 1200, 1800, 2700, 3600];
const DURATION_LABELS: Record<number, string> = {
  5: "5s",
  15: "15s",
  30: "30s",
  60: "1m",
  300: "5m",
  600: "10m",
  1200: "20m",
  1800: "30m",
  2700: "45m",
  3600: "1h",
};

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
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [curveView, setCurveView] = useState<"chart" | "table">("chart");

  useEffect(() => {
    async function fetchData() {
      try {
        const data = await api.getPowerCurve(activityId);
        // Sort by duration
        const sorted = data.sort((a, b) => a.durationSeconds - b.durationSeconds);
        setPowerCurve(sorted);
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

  const hasHR = tableRows.some((r) => r.avgHeartRate != null);

  return (
    <div className="mt-6 space-y-6">
      {/* Key Power Outputs */}
      <div>
        <h3 className="mb-3 text-sm font-medium text-foreground-muted">Key Power Outputs</h3>
        <div className="space-y-3">
          {/* Row 1: Power data */}
          <div className="grid grid-cols-3 gap-3">
            {activity.avgPower && (
              <div className="rounded-lg border border-border bg-background-subtle p-4">
                <p className="text-xs text-foreground-muted">Avg Power</p>
                <p className="mt-1 font-mono text-2xl text-power">
                  {activity.avgPower}<span className="ml-1 text-sm text-foreground-muted">w</span>
                </p>
              </div>
            )}
            {activity.normalizedPower && (
              <div className="rounded-lg border border-border bg-background-subtle p-4">
                <p className="text-xs text-foreground-muted">Normalized Power</p>
                <p className="mt-1 font-mono text-2xl text-power">
                  {activity.normalizedPower}<span className="ml-1 text-sm text-foreground-muted">w</span>
                </p>
              </div>
            )}
            {activity.maxPower && (
              <div className="rounded-lg border border-border bg-background-subtle p-4">
                <p className="text-xs text-foreground-muted">Max Power</p>
                <p className="mt-1 font-mono text-2xl text-power">
                  {activity.maxPower}<span className="ml-1 text-sm text-foreground-muted">w</span>
                </p>
              </div>
            )}
          </div>
          {/* Row 2: Derived metrics */}
          <div className="grid grid-cols-3 gap-3">
            {activity.intensityFactor && (
              <div className="rounded-lg border border-border bg-background-subtle p-4">
                <p className="text-xs text-foreground-muted">Intensity Factor</p>
                <p className="mt-1 font-mono text-2xl">
                  {activity.intensityFactor.toFixed(2)}
                </p>
              </div>
            )}
            {activity.tss && (
              <div className="rounded-lg border border-border bg-background-subtle p-4">
                <p className="text-xs text-foreground-muted">TSS</p>
                <p className="mt-1 font-mono text-2xl">
                  {activity.tss.toFixed(0)}
                </p>
              </div>
            )}
            {activity.variabilityIndex && (
              <div className="rounded-lg border border-border bg-background-subtle p-4">
                <p className="text-xs text-foreground-muted">Variability Index</p>
                <p className="mt-1 font-mono text-2xl">
                  {activity.variabilityIndex.toFixed(2)}
                </p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Key Power Outputs */}
      {keyPowers.length > 0 && (
        <div>
          <h3 className="mb-3 text-sm font-medium text-foreground-muted">Peak Power</h3>
          <div className="grid grid-cols-3 gap-3 sm:grid-cols-6">
            {keyPowers.map(({ duration, label, power }) => (
              <div
                key={duration}
                className="rounded-lg border border-border bg-background-subtle p-3 text-center"
              >
                <p className="text-xs text-foreground-muted">{label}</p>
                <p className="mt-1 font-mono text-lg text-power">{power}w</p>
              </div>
            ))}
          </div>
        </div>
      )}

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
                  <thead className="bg-background-subtle">
                    <tr>
                      <th className="px-4 py-2 text-left font-medium text-foreground-muted">
                        Duration
                      </th>
                      <th className="px-4 py-2 text-right font-medium text-foreground-muted">
                        Best Power
                      </th>
                      {hasHR && (
                        <th className="px-4 py-2 text-right font-medium text-foreground-muted">
                          Avg HR
                        </th>
                      )}
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border">
                    {tableRows.map((point) => (
                      <tr key={point.durationSeconds} className="hover:bg-background-subtle/50">
                        <td className="px-4 py-2 font-mono">
                          {DURATION_LABELS[point.durationSeconds]}
                        </td>
                        <td className="px-4 py-2 text-right font-mono text-power">
                          {point.bestPower}w
                        </td>
                        {hasHR && (
                          <td className="px-4 py-2 text-right font-mono text-heart-rate">
                            {point.avgHeartRate != null ? `${point.avgHeartRate} bpm` : "—"}
                          </td>
                        )}
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
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.06)" />
                    <XAxis
                      dataKey="index"
                      type="number"
                      domain={[0, CHART_DURATIONS.length - 1]}
                      ticks={CHART_DURATIONS.map((_, i) => i)}
                      tickFormatter={(i) => formatDuration(CHART_DURATIONS[i])}
                      tick={{ fill: "#6b7280", fontSize: 11, fontFamily: "DM Mono, monospace" }}
                      axisLine={false}
                      tickLine={false}
                      interval={1}
                    />
                    <YAxis
                      tickFormatter={(v) => `${v}w`}
                      tick={{ fill: "#6b7280", fontSize: 11, fontFamily: "DM Mono, monospace" }}
                      axisLine={false}
                      tickLine={false}
                      width={52}
                    />
                    <Tooltip
                      contentStyle={{ backgroundColor: "var(--color-background-subtle)", border: "1px solid var(--color-border)", borderRadius: "6px", fontSize: 12, color: "var(--color-foreground)" }}
                      labelFormatter={(i) => formatDuration(CHART_DURATIONS[Number(i)])}
                      formatter={(value: number | undefined) => [`${value ?? "—"}w`, "Best Power"]}
                    />
                    <Line
                      type="monotone"
                      dataKey="bestPower"
                      stroke="#F97316"
                      strokeWidth={2}
                      dot={false}
                      activeDot={{ r: 4, fill: "#F97316" }}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
