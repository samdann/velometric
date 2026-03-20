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
import {
  CHART_GRID_STROKE,
  CHART_TICK_STYLE,
  CHART_TOOLTIP_CONTENT_STYLE,
  CHART_COLORS,
  formatDuration,
} from "@/lib/chart-config";

export interface PowerCurveDataPoint {
  durationSeconds: number;
  power: number;
}

interface Props {
  data: PowerCurveDataPoint[];
  tooltipLabel?: string;
  /** XAxis interval — use 0 for sparse data (~10 pts), 1 to skip every other label for denser sets */
  labelInterval?: number;
}

export function PowerCurveChart({ data, tooltipLabel = "Power", labelInterval = 0 }: Props) {
  const chartData = data.map((p, i) => ({ ...p, index: i }));

  return (
    <ResponsiveContainer width="100%" height="100%">
      <LineChart data={chartData} margin={{ top: 4, right: 8, left: 0, bottom: 4 }}>
        <CartesianGrid strokeDasharray="3 3" stroke={CHART_GRID_STROKE} />
        <XAxis
          dataKey="index"
          type="number"
          domain={[0, chartData.length - 1]}
          ticks={chartData.map((_, i) => i)}
          tickFormatter={(i) => formatDuration(chartData[i]?.durationSeconds ?? 0)}
          tick={CHART_TICK_STYLE}
          axisLine={false}
          tickLine={false}
          interval={labelInterval}
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
          labelFormatter={(i) => formatDuration(chartData[Number(i)]?.durationSeconds ?? 0)}
          formatter={(value: number | undefined) => [`${value ?? "—"}w`, tooltipLabel]}
        />
        <Line
          type="monotone"
          dataKey="power"
          stroke={CHART_COLORS.power}
          strokeWidth={2}
          dot={false}
          activeDot={{ r: 4, fill: CHART_COLORS.power }}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}
