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
  return (meters / 1000).toFixed(2) + " km";
}

function formatSpeed(mps: number): string {
  return (mps * 3.6).toFixed(1) + " km/h";
}

// Shades of red matching the HR zones palette
const EFFORT_SHADES = [
  "#FECACA", // 0–20%
  "#FCA5A5", // 20–40%
  "#F87171", // 40–60%
  "#EF4444", // 60–80%
  "#DC2626", // 80–100%+
];

function effortColor(pct: number): string {
  const idx = Math.min(Math.floor(pct / 20), EFFORT_SHADES.length - 1);
  return EFFORT_SHADES[idx];
}

function EffortBar({ power, ftp }: { power?: number; ftp?: number }) {
  const [hovered, setHovered] = useState(false);

  if (!power || !ftp || ftp === 0) {
    return <span className="font-mono text-sm text-foreground-muted">—</span>;
  }

  const ftpPct = (power / ftp) * 100;
  const color = effortColor(ftpPct);
  // Cap bar fill at 150% FTP = 100% width
  const barWidth = Math.min((ftpPct / 150) * 100, 100);

  return (
    <div
      className="relative"
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <div className="h-5 w-28 overflow-hidden rounded-sm bg-background">
        <div
          className="h-full rounded-sm transition-all duration-300"
          style={{ width: `${barWidth}%`, backgroundColor: color }}
        />
      </div>
      {hovered && (
        <div
          className="absolute bottom-full left-0 z-10 mb-1.5 rounded border border-border bg-background-subtle px-2.5 py-1.5 text-xs shadow-lg"
          style={{ whiteSpace: "nowrap" }}
        >
          <span className="font-mono font-semibold text-foreground">{power} W</span>
          <span className="mx-1.5 text-foreground-muted">·</span>
          <span className="font-mono font-semibold" style={{ color }}>
            {ftpPct.toFixed(0)}% FTP
          </span>
        </div>
      )}
    </div>
  );
}

export function LapsTab({ activityId }: LapsTabProps) {
  const [laps, setLaps] = useState<Lap[]>([]);
  const [ftp, setFtp] = useState<number | undefined>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([
      api.getLaps(activityId),
      api.getUserProfile(),
    ])
      .then(([lapData, profile]) => {
        setLaps(lapData || []);
        setFtp(profile.ftp);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load laps"))
      .finally(() => setLoading(false));
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
        <table className="w-full min-w-[900px]">
          <thead className="bg-background-subtle">
            <tr>
              <th className="px-3 py-2.5 text-left text-xs font-medium text-foreground-muted">Lap</th>
              <th className="px-3 py-2.5 text-right text-xs font-medium text-foreground-muted">Distance</th>
              <th className="px-3 py-2.5 text-right text-xs font-medium text-foreground-muted">Time</th>
              <th className="px-3 py-2.5 text-right text-xs font-medium text-foreground-muted">Elevation</th>
              <th className="px-3 py-2.5 text-right text-xs font-medium text-foreground-muted">Speed</th>
              <th className="px-3 py-2.5 text-right text-xs font-medium text-foreground-muted">HR</th>
              <th className="px-3 py-2.5 text-right text-xs font-medium text-foreground-muted">Cad</th>
              <th className="px-3 py-2.5 text-right text-xs font-medium text-foreground-muted">Power</th>
              <th className="px-3 py-2.5 text-left text-xs font-medium text-foreground-muted">Effort Intensity</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {laps.map((lap) => (
              <tr key={lap.lapNumber} className="hover:bg-background-subtle/50">
                <td className="px-3 py-2 font-mono text-sm font-medium text-foreground">
                  {lap.lapNumber}
                </td>
                <td className="px-3 py-2 text-right font-mono text-sm text-foreground">
                  {formatDistance(lap.distance)}
                </td>
                <td className="px-3 py-2 text-right font-mono text-sm text-foreground">
                  {formatDuration(lap.duration)}
                </td>
                <td className="px-3 py-2 text-right font-mono text-sm text-elevation">
                  {lap.ascent ? `+${lap.ascent.toFixed(0)} m` : "—"}
                </td>
                <td className="px-3 py-2 text-right font-mono text-sm text-speed">
                  {lap.avgSpeed ? formatSpeed(lap.avgSpeed) : "—"}
                </td>
                <td className="px-3 py-2 text-right font-mono text-sm text-heart-rate">
                  {lap.avgHeartRate ?? "—"}
                </td>
                <td className="px-3 py-2 text-right font-mono text-sm text-cadence">
                  {lap.avgCadence ?? "—"}
                </td>
                <td className="px-3 py-2 text-right font-mono text-sm text-power">
                  {lap.avgPower ? `${lap.avgPower} W` : "—"}
                </td>
                <td className="px-3 py-2">
                  <EffortBar power={lap.avgPower} ftp={ftp} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
