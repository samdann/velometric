"use client";

import { useEffect, useState } from "react";
import { api, AnnualPowerStats } from "@/lib/api";
import { PowerCurveWidget } from "./power-curve-widget";
import { PowerDistributionWidget } from "./power-distribution-widget";

function WidgetCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="p-1">
      <h3 className="text-sm font-medium text-zinc-400 mb-4">{title}</h3>
      {children}
    </div>
  );
}

type ZoneMode = "avg" | "best";

export function StatsClient() {
  const [years, setYears] = useState<number[]>([]);
  const [selectedYear, setSelectedYear] = useState<number | null>(null);
  const [zoneMode, setZoneMode] = useState<ZoneMode>("avg");
  const [stats, setStats] = useState<AnnualPowerStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.getStatisticsYears().then((y) => {
      setYears(y);
      if (y.length > 0) setSelectedYear(y[0]);
      else setLoading(false);
    }).catch((e: Error) => {
      setError(e.message);
      setLoading(false);
    });
  }, []);

  useEffect(() => {
    if (!selectedYear) return;
    setLoading(true);
    setError(null);
    api
      .getStatisticsPower(selectedYear, zoneMode)
      .then(setStats)
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false));
  }, [selectedYear, zoneMode]);

  if (!loading && years.length === 0) {
    return (
      <div className="flex items-center justify-center h-48 text-zinc-500 text-sm">
        No activities with power data yet. Upload rides with a power meter to see statistics.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Controls */}
      <div className="flex items-center gap-6">
        <div className="flex items-center gap-3">
          <label className="text-sm text-zinc-400">Year</label>
          <select
            value={selectedYear ?? ""}
            onChange={(e) => setSelectedYear(Number(e.target.value))}
            className="bg-background-subtle border border-border rounded-lg px-3 py-1.5 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-power"
          >
            {years.map((y) => (
              <option key={y} value={y}>
                {y}
              </option>
            ))}
          </select>
        </div>
        <div className="flex items-center gap-3">
          <label className="text-sm text-zinc-400">Zone distribution</label>
          <div className="flex rounded-lg border border-border overflow-hidden text-sm">
            {(["avg", "best"] as ZoneMode[]).map((m) => (
              <button
                key={m}
                onClick={() => setZoneMode(m)}
                className={`px-3 py-1.5 capitalize transition-colors ${
                  zoneMode === m
                    ? "bg-power text-white"
                    : "bg-background-subtle text-zinc-400 hover:text-foreground"
                }`}
              >
                {m === "avg" ? "Avg" : "Best"}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Widget grid */}
      {loading ? (
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
          {[0, 1].map((i) => (
            <div key={i} className="bg-zinc-900 border border-zinc-800 rounded-xl p-5 h-64 animate-pulse" />
          ))}
        </div>
      ) : error ? (
        <div className="text-red-400 text-sm">{error}</div>
      ) : stats ? (
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
          <WidgetCard title={`Median Power Curve — ${selectedYear}`}>
            <PowerCurveWidget data={stats.powerCurve} year={selectedYear!} />
          </WidgetCard>
          <WidgetCard title={`${zoneMode === "avg" ? "Median" : "Best"} Zone Distribution — ${selectedYear}`}>
            <PowerDistributionWidget data={stats.zoneDistribution} year={selectedYear!} />
          </WidgetCard>
        </div>
      ) : null}
    </div>
  );
}
