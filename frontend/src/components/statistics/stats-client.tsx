"use client";

import { useEffect, useState } from "react";
import { api, AnnualPowerStats } from "@/lib/api";
import { PowerCurveWidget } from "./power-curve-widget";
import { PowerDistributionWidget } from "./power-distribution-widget";

type ZoneMode = "avg" | "best";

export function StatsClient() {
  const [years, setYears] = useState<number[]>([]);
  const [selectedYear, setSelectedYear] = useState<number | null>(null);
  const [zoneMode, setZoneMode] = useState<ZoneMode>("avg");
  const [stats, setStats] = useState<AnnualPowerStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [fetching, setFetching] = useState(false);
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
    if (stats === null) setLoading(true);
    else setFetching(true);
    setError(null);
    api
      .getStatisticsPower(selectedYear, zoneMode)
      .then(setStats)
      .catch((e: Error) => setError(e.message))
      .finally(() => { setLoading(false); setFetching(false); });
  }, [selectedYear, zoneMode]);

  if (!loading && years.length === 0) {
    return (
      <div className="flex items-center justify-center h-48 text-foreground-muted text-sm">
        No activities with power data yet. Upload rides with a power meter to see statistics.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Controls */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <label className="text-sm text-foreground-muted">Year</label>
          <select
            value={selectedYear ?? ""}
            onChange={(e) => setSelectedYear(Number(e.target.value))}
            className="bg-background-subtle border border-border rounded-lg px-3 py-1.5 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-power"
          >
            {years.map((y) => (
              <option key={y} value={y}>{y}</option>
            ))}
          </select>
          {fetching && (
            <svg className="animate-spin h-4 w-4 text-foreground-muted" viewBox="0 0 24 24" fill="none">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
            </svg>
          )}
        </div>
        <div className="flex rounded-lg border border-border overflow-hidden text-sm">
          {(["avg", "best"] as ZoneMode[]).map((m) => (
            <button
              key={m}
              onClick={() => setZoneMode(m)}
              className={`px-3 py-1.5 transition-colors ${
                zoneMode === m
                  ? "bg-power text-white"
                  : "bg-background-subtle text-foreground-muted hover:text-foreground"
              }`}
            >
              {m === "avg" ? "Avg" : "Best"}
            </button>
          ))}
        </div>
      </div>

      {/* Widget grid */}
      {loading ? (
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
          {[0, 1].map((i) => (
            <div key={i} className="bg-background-subtle border border-border rounded-xl p-5 h-64 animate-pulse" />
          ))}
        </div>
      ) : error ? (
        <div className="text-red-400 text-sm">{error}</div>
      ) : stats ? (
        <div className={`grid grid-cols-1 xl:grid-cols-2 gap-4 transition-opacity duration-150 ${fetching ? "opacity-40" : "opacity-100"}`}>
          <PowerCurveWidget data={stats.powerCurve} year={selectedYear!} mode={zoneMode} />
          <PowerDistributionWidget data={stats.zoneDistribution} year={selectedYear!} mode={zoneMode} />
        </div>
      ) : null}
    </div>
  );
}
