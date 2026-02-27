"use client";

import { useEffect, useState } from "react";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from "recharts";
import { api } from "@/lib/api";

interface ElevationChartProps {
  activityId: string;
}

export function ElevationChart({ activityId }: ElevationChartProps) {
  const [data, setData] = useState<{ distance: number; altitude: number }[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.getElevationProfile(activityId)
      .then(setData)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load elevation"))
      .finally(() => setLoading(false));
  }, [activityId]);

  if (loading) {
    return (
      <div className="flex h-40 items-center justify-center">
        <p className="text-sm text-foreground-muted">Loading elevation...</p>
      </div>
    );
  }

  if (error || data.length === 0) return null;

  const minAlt = Math.min(...data.map((d) => d.altitude));
  const maxAlt = Math.max(...data.map((d) => d.altitude));
  const maxDist = Math.max(...data.map((d) => d.distance));

  function niceStep(range: number, steps: number[]): number {
    for (const s of steps) if (range / s <= 8) return s;
    return steps[steps.length - 1];
  }

  const altStep = niceStep(maxAlt - minAlt, [10, 20, 50, 100, 200, 500]);
  const altMin = Math.floor(minAlt / altStep) * altStep;
  const altMax = Math.ceil(maxAlt / altStep) * altStep;
  const altTicks: number[] = [];
  for (let t = altMin; t <= altMax; t += altStep) altTicks.push(t);

  const distStep = niceStep(maxDist, [1, 2, 5, 10, 20, 50]);
  const distTicks: number[] = [];
  for (let t = 0; t <= maxDist; t += distStep) distTicks.push(parseFloat(t.toFixed(1)));

  return (
    <div className="mt-6">
      <h3 className="mb-3 text-sm font-medium text-foreground-muted">Elevation Profile</h3>
      <div className="rounded-lg border border-border bg-background-subtle p-4">
        <ResponsiveContainer width="100%" height={160}>
          <AreaChart data={data} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
            <defs>
              <linearGradient id="elevationGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#22C55E" stopOpacity={0.3} />
                <stop offset="95%" stopColor="#22C55E" stopOpacity={0.02} />
              </linearGradient>
            </defs>
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
            <YAxis
              domain={[altMin, altMax]}
              ticks={altTicks}
              tickFormatter={(v) => `${v}m`}
              tick={{ fontSize: 10, fill: "var(--color-foreground-muted)" }}
              axisLine={false}
              tickLine={false}
              width={45}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: "var(--color-background-subtle)",
                border: "1px solid var(--color-border)",
                borderRadius: "6px",
                fontSize: "12px",
              }}
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              formatter={(value: any) => [`${Math.round(value ?? 0)}m`, "Elevation"]}
              labelFormatter={(label: any) => `${Number(label).toFixed(1)} km`}
            />
            <Area
              type="monotone"
              dataKey="altitude"
              stroke="#22C55E"
              strokeWidth={1.5}
              fill="url(#elevationGradient)"
              dot={false}
              isAnimationActive={false}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
