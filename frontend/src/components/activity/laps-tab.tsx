"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";

interface Lap {
  lapNumber: number;
  startTime: string;
  duration: number;
  distance: number;
  avgPower?: number;
  maxPower?: number;
  avgHeartRate?: number;
  maxHeartRate?: number;
  avgCadence?: number;
  avgSpeed?: number;
  maxSpeed?: number;
  ascent?: number;
  descent?: number;
  trigger?: string;
}

interface LapsTabProps {
  activityId: string;
}

function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m}:${s.toString().padStart(2, "0")}`;
}

function formatDistance(meters: number): string {
  return (meters / 1000).toFixed(2);
}

function formatSpeed(mps: number): string {
  return (mps * 3.6).toFixed(1);
}

export function LapsTab({ activityId }: LapsTabProps) {
  const [laps, setLaps] = useState<Lap[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchLaps() {
      try {
        const response = await fetch(
          `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8081"}/api/activities/${activityId}/laps`
        );
        if (!response.ok) throw new Error("Failed to fetch laps");
        const data = await response.json();
        setLaps(data || []);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load laps");
      } finally {
        setLoading(false);
      }
    }
    fetchLaps();
  }, [activityId]);

  if (loading) {
    return (
      <div className="mt-6 flex h-32 items-center justify-center">
        <p className="text-foreground-muted">Loading laps...</p>
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

  if (laps.length === 0) {
    return (
      <div className="mt-6 rounded-lg border border-border bg-background-subtle p-6 text-center">
        <p className="text-foreground-muted">No lap data available</p>
      </div>
    );
  }

  return (
    <div className="mt-6">
      <div className="overflow-x-auto rounded-lg border border-border">
        <table className="w-full min-w-[800px]">
          <thead className="bg-background-subtle">
            <tr>
              <th className="px-3 py-2 text-left text-xs font-medium text-foreground-muted">Lap</th>
              <th className="px-3 py-2 text-right text-xs font-medium text-foreground-muted">Time</th>
              <th className="px-3 py-2 text-right text-xs font-medium text-foreground-muted">Dist (km)</th>
              <th className="px-3 py-2 text-right text-xs font-medium text-foreground-muted">Avg Pwr</th>
              <th className="px-3 py-2 text-right text-xs font-medium text-foreground-muted">Max Pwr</th>
              <th className="px-3 py-2 text-right text-xs font-medium text-foreground-muted">Avg HR</th>
              <th className="px-3 py-2 text-right text-xs font-medium text-foreground-muted">Avg Cad</th>
              <th className="px-3 py-2 text-right text-xs font-medium text-foreground-muted">Avg Spd</th>
              <th className="px-3 py-2 text-right text-xs font-medium text-foreground-muted">Elev</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {laps.map((lap) => (
              <tr key={lap.lapNumber} className="hover:bg-background-subtle/50">
                <td className="px-3 py-2 text-sm font-medium">{lap.lapNumber}</td>
                <td className="px-3 py-2 text-right font-mono text-sm">{formatDuration(lap.duration)}</td>
                <td className="px-3 py-2 text-right font-mono text-sm">{formatDistance(lap.distance)}</td>
                <td className="px-3 py-2 text-right font-mono text-sm text-power">
                  {lap.avgPower ?? "-"}
                </td>
                <td className="px-3 py-2 text-right font-mono text-sm text-power">
                  {lap.maxPower ?? "-"}
                </td>
                <td className="px-3 py-2 text-right font-mono text-sm text-heart-rate">
                  {lap.avgHeartRate ?? "-"}
                </td>
                <td className="px-3 py-2 text-right font-mono text-sm text-cadence">
                  {lap.avgCadence ?? "-"}
                </td>
                <td className="px-3 py-2 text-right font-mono text-sm text-speed">
                  {lap.avgSpeed ? formatSpeed(lap.avgSpeed) : "-"}
                </td>
                <td className="px-3 py-2 text-right font-mono text-sm text-elevation">
                  {lap.ascent ? `+${lap.ascent.toFixed(0)}m` : "-"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
