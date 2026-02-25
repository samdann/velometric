"use client";

import { useEffect, useState } from "react";
import { api, Activity } from "@/lib/api";

interface PowerCurvePoint {
  durationSeconds: number;
  bestPower: number;
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

// Standard power curve durations to highlight
const KEY_DURATIONS = [5, 30, 60, 300, 1200, 3600];
const DURATION_LABELS: Record<number, string> = {
  5: "5 sec",
  30: "30 sec",
  60: "1 min",
  300: "5 min",
  1200: "20 min",
  3600: "1 hour",
};

export function PowerTab({ activityId, activity }: PowerTabProps) {
  const [powerCurve, setPowerCurve] = useState<PowerCurvePoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

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

  // Get key power values
  const keyPowers = KEY_DURATIONS.map((duration) => {
    const point = powerCurve.find((p) => p.durationSeconds === duration);
    return {
      duration,
      label: DURATION_LABELS[duration],
      power: point?.bestPower ?? null,
    };
  }).filter((p) => p.power !== null);

  return (
    <div className="mt-6 space-y-6">
      {/* Power Summary */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
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

      {/* Full Power Curve Table */}
      {powerCurve.length > 0 && (
        <div>
          <h3 className="mb-3 text-sm font-medium text-foreground-muted">Power Curve</h3>
          <div className="overflow-hidden rounded-lg border border-border">
            <table className="w-full">
              <thead className="bg-background-subtle">
                <tr>
                  <th className="px-4 py-2 text-left text-xs font-medium text-foreground-muted">
                    Duration
                  </th>
                  <th className="px-4 py-2 text-right text-xs font-medium text-foreground-muted">
                    Best Power
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {powerCurve.map((point) => (
                  <tr key={point.durationSeconds} className="hover:bg-background-subtle/50">
                    <td className="px-4 py-2 text-sm">
                      {formatDuration(point.durationSeconds)}
                    </td>
                    <td className="px-4 py-2 text-right font-mono text-sm text-power">
                      {point.bestPower}w
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
