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

interface SpeedChartProps {
  activityId: string;
}

export function SpeedChart({ activityId }: SpeedChartProps) {
  const [data, setData] = useState<{ distance: number; speed: number }[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.getSpeedProfile(activityId)
      .then(setData)
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load speed"))
      .finally(() => setLoading(false));
  }, [activityId]);

  if (loading) {
    return (
      <div className="flex h-40 items-center justify-center">
        <p className="text-sm text-foreground-muted">Loading speed...</p>
      </div>
    );
  }

  if (error || data.length === 0) return null;

  const minSpeed = Math.min(...data.map((d) => d.speed));
  const maxSpeed = Math.max(...data.map((d) => d.speed));
  const maxDist = Math.max(...data.map((d) => d.distance));

  function niceStep(range: number, steps: number[]): number {
    for (const s of steps) if (range / s <= 8) return s;
    return steps[steps.length - 1];
  }

  const speedStep = niceStep(maxSpeed - minSpeed, [1, 2, 5, 10, 20]);
  const speedMin = Math.floor(minSpeed / speedStep) * speedStep;
  const speedMax = Math.ceil(maxSpeed / speedStep) * speedStep;
  const speedTicks: number[] = [];
  for (let t = speedMin; t <= speedMax; t += speedStep) speedTicks.push(parseFloat(t.toFixed(1)));

  const distStep = niceStep(maxDist, [1, 2, 5, 10, 20, 50]);
  const distTicks: number[] = [];
  for (let t = 0; t <= maxDist; t += distStep) distTicks.push(parseFloat(t.toFixed(1)));

  return (
    <div className="mt-6">
      <h3 className="mb-3 text-sm font-medium text-foreground-muted">Speed Profile</h3>
      <div className="rounded-lg border border-border bg-background-subtle p-4">
        <ResponsiveContainer width="100%" height={160}>
          <AreaChart data={data} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
            <defs>
              <linearGradient id="speedGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#3B82F6" stopOpacity={0.3} />
                <stop offset="95%" stopColor="#3B82F6" stopOpacity={0.02} />
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
              domain={[speedMin, speedMax]}
              ticks={speedTicks}
              tickFormatter={(v) => `${v}`}
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
              formatter={(value: any) => [`${Number(value ?? 0).toFixed(1)} km/h`, "Speed"]}
              labelFormatter={(label: any) => `${Number(label).toFixed(1)} km`}
            />
            <Area
              type="monotone"
              dataKey="speed"
              stroke="#3B82F6"
              strokeWidth={1.5}
              fill="url(#speedGradient)"
              dot={false}
              isAnimationActive={false}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
