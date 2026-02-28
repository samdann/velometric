"use client";

import { useEffect, useState } from "react";
import {
  ComposedChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from "recharts";
import { api } from "@/lib/api";

interface HRCadenceChartProps {
  activityId: string;
}

export function HRCadenceChart({ activityId }: HRCadenceChartProps) {
  const [data, setData] = useState<{ distance: number; heartRate?: number; cadence?: number }[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.getHRCadenceProfile(activityId)
      .then(setData)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load HR/cadence"))
      .finally(() => setLoading(false));
  }, [activityId]);

  if (loading) {
    return (
      <div className="flex h-40 items-center justify-center">
        <p className="text-sm text-foreground-muted">Loading heart rate...</p>
      </div>
    );
  }

  if (error || data.length === 0) return null;

  const hasHR = data.some((d) => d.heartRate != null);
  const hasCadence = data.some((d) => d.cadence != null);

  if (!hasHR && !hasCadence) return null;

  const maxDist = Math.max(...data.map((d) => d.distance));

  function niceStep(range: number, steps: number[]): number {
    for (const s of steps) if (range / s <= 8) return s;
    return steps[steps.length - 1];
  }

  const distStep = niceStep(maxDist, [1, 2, 5, 10, 20, 50]);
  const distTicks: number[] = [];
  for (let t = 0; t <= maxDist; t += distStep) distTicks.push(parseFloat(t.toFixed(1)));

  const hrValues = hasHR ? data.map((d) => d.heartRate ?? 0).filter((v) => v > 0) : [];
  const minHR = hasHR ? Math.min(...hrValues) : 0;
  const maxHR = hasHR ? Math.max(...hrValues) : 0;
  const hrStep = niceStep(maxHR - minHR, [5, 10, 20, 30]);
  const hrMin = Math.floor(minHR / hrStep) * hrStep;
  const hrMax = Math.ceil(maxHR / hrStep) * hrStep;
  const hrTicks: number[] = [];
  for (let t = hrMin; t <= hrMax; t += hrStep) hrTicks.push(t);

  const cadValues = hasCadence ? data.map((d) => d.cadence ?? 0).filter((v) => v > 0) : [];
  const minCad = hasCadence ? Math.min(...cadValues) : 0;
  const maxCad = hasCadence ? Math.max(...cadValues) : 0;
  const cadStep = niceStep(maxCad - minCad, [5, 10, 20]);
  const cadMin = Math.floor(minCad / cadStep) * cadStep;
  const cadMax = Math.ceil(maxCad / cadStep) * cadStep;
  const cadTicks: number[] = [];
  for (let t = cadMin; t <= cadMax; t += cadStep) cadTicks.push(t);

  return (
    <div className="mt-6">
      <div className="mb-3 flex items-center gap-4">
        <h3 className="text-sm font-medium text-foreground-muted">Heart Rate &amp; Cadence</h3>
        <div className="flex items-center gap-3 text-xs text-foreground-muted">
          {hasHR && (
            <span className="flex items-center gap-1">
              <span className="inline-block h-0.5 w-4 bg-[#EF4444]" />
              Heart Rate
            </span>
          )}
          {hasCadence && (
            <span className="flex items-center gap-1">
              <span className="inline-block h-0.5 w-4 bg-[#A855F7]" />
              Cadence
            </span>
          )}
        </div>
      </div>
      <div className="rounded-lg border border-border bg-background-subtle p-4">
        <ResponsiveContainer width="100%" height={160}>
          <ComposedChart data={data} margin={{ top: 4, right: hasCadence ? 48 : 8, left: 0, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" />
            <XAxis
              dataKey="distance"
              type="number"
              domain={[0, maxDist]}
              ticks={distTicks}
              tickFormatter={(v) => `${v}km`}
              tick={{ fontSize: 10, fill: "var(--color-foreground-muted)" }}
              axisLine={false}
              tickLine={false}
            />
            {hasHR && (
              <YAxis
                yAxisId="hr"
                domain={[hrMin, hrMax]}
                ticks={hrTicks}
                tickFormatter={(v) => `${v}`}
                tick={{ fontSize: 10, fill: "var(--color-foreground-muted)" }}
                axisLine={false}
                tickLine={false}
                width={45}
              />
            )}
            {hasCadence && (
              <YAxis
                yAxisId="cadence"
                orientation="right"
                domain={[cadMin, cadMax]}
                ticks={cadTicks}
                tickFormatter={(v) => `${v}`}
                tick={{ fontSize: 10, fill: "var(--color-foreground-muted)" }}
                axisLine={false}
                tickLine={false}
                width={45}
              />
            )}
            <Tooltip
              contentStyle={{
                backgroundColor: "var(--color-background-subtle)",
                border: "1px solid var(--color-border)",
                borderRadius: "6px",
                fontSize: "12px",
              }}
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              formatter={(value: any, name: any) => {
                if (name === "heartRate") return [`${Math.round(value ?? 0)} bpm`, "Heart Rate"];
                if (name === "cadence") return [`${Math.round(value ?? 0)} rpm`, "Cadence"];
                return [value, name];
              }}
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              labelFormatter={(label: any) => `${Number(label).toFixed(1)} km`}
            />
            {hasHR && (
              <Line
                yAxisId="hr"
                type="monotone"
                dataKey="heartRate"
                stroke="#EF4444"
                strokeWidth={1.5}
                dot={false}
                isAnimationActive={false}
                connectNulls={false}
              />
            )}
            {hasCadence && (
              <Line
                yAxisId="cadence"
                type="monotone"
                dataKey="cadence"
                stroke="#A855F7"
                strokeWidth={1.5}
                dot={false}
                isAnimationActive={false}
                connectNulls={false}
              />
            )}
          </ComposedChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
