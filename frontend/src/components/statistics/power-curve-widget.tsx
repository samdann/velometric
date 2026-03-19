"use client";

import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { AnnualPowerCurvePoint } from "@/lib/api";
import {
  CHART_GRID_STROKE,
  CHART_TICK_STYLE,
  CHART_TOOLTIP_CONTENT_STYLE,
  CHART_COLORS,
  POWER_CURVE_DURATIONS,
  DURATION_LABELS,
} from "@/lib/chart-config";

interface Props {
  data: AnnualPowerCurvePoint[];
  year: number;
}

export function PowerCurveWidget({ data, year }: Props) {
  if (data.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-background-subtle p-6 text-center">
        <p className="text-sm text-foreground-muted">No power curve data for {year}</p>
      </div>
    );
  }

  // Index-based x-axis — same approach as the activity power tab chart
  const chartData = POWER_CURVE_DURATIONS
    .map((dur, index) => {
      const point = data.find((p) => p.durationSeconds === dur);
      return point ? { index, label: DURATION_LABELS[dur], medianPower: point.medianPower } : null;
    })
    .filter(Boolean) as { index: number; label: string; medianPower: number }[];

  return (
    <div className="rounded-lg border border-border bg-background-subtle p-4 h-64">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={chartData} margin={{ top: 4, right: 8, left: 0, bottom: 4 }}>
          <CartesianGrid strokeDasharray="3 3" stroke={CHART_GRID_STROKE} />
          <XAxis
            dataKey="index"
            type="number"
            domain={[0, chartData.length - 1]}
            ticks={chartData.map((_, i) => i)}
            tickFormatter={(i) => chartData[i]?.label ?? ""}
            tick={CHART_TICK_STYLE}
            axisLine={false}
            tickLine={false}
            interval={0}
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
            labelFormatter={(i) => chartData[Number(i)]?.label ?? ""}
            formatter={(value: number | undefined) => [`${value ?? "—"}w`, "Median Power"]}
          />
          <Line
            type="monotone"
            dataKey="medianPower"
            stroke={CHART_COLORS.power}
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 4, fill: CHART_COLORS.power }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
