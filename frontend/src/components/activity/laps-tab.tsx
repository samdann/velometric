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
function LapOverviewChart({ laps, activityId }: { laps: Lap[]; activityId: string }) {
  const [elevation, setElevation] = useState<{ distance: number; altitude: number }[]>([]);
  const [hoveredLap, setHoveredLap] = useState<number | null>(null);

  useEffect(() => {
    api.getElevationProfile(activityId).then(setElevation).catch(() => {});
  }, [activityId]);

  const totalDuration = laps.reduce((s, l) => s + l.duration, 0);
  const maxPower = Math.max(...laps.map((l) => l.avgPower ?? 0));
  if (totalDuration === 0 || maxPower === 0) return null;

  // Elevation SVG path stretched across full width (decorative background)
  let elevationPath = "";
  if (elevation.length > 1) {
    const minAlt = Math.min(...elevation.map((e) => e.altitude));
    const maxAlt = Math.max(...elevation.map((e) => e.altitude));
    const altRange = maxAlt - minAlt || 1;
    // Elevation fills full chart height — highest point reaches the top
    const yTop = 0;
    const yBottom = 100;
    const pts = elevation.map((e, i) => {
      const x = (i / (elevation.length - 1)) * 100;
      const y = yBottom - ((e.altitude - minAlt) / altRange) * (yBottom - yTop);
      return `${x.toFixed(1)} ${y.toFixed(1)}`;
    });
    elevationPath = `M0 ${yBottom} L${pts.join(" L")} L100 ${yBottom} Z`;
  }

  // Bars — time-based widths, power-based heights
  let cumTime = 0;
  const bars = laps.map((lap) => {
    const xPct = (cumTime / totalDuration) * 100;
    const wPct = (lap.duration / totalDuration) * 100;
    const power = lap.avgPower ?? 0;
    const hPct = (power / maxPower) * 100;
    const opacity = 0.2 + (power / maxPower) * 0.4;
    cumTime += lap.duration;
    return { xPct, wPct, hPct, power, opacity, lap };
  });

  return (
    <div className="mb-6">
      <div
        className="relative rounded-lg border border-border bg-background-subtle overflow-visible"
        style={{ height: 120 }}
      >
        {/* Elevation background */}
        {elevationPath && (
          <svg
            className="absolute inset-[3px] w-[calc(100%-6px)] h-[calc(100%-6px)] rounded-lg"
            viewBox="0 0 100 100"
            preserveAspectRatio="none"
          >
            <path d={elevationPath} fill="var(--color-foreground-muted)" fillOpacity={0.08} />
          </svg>
        )}

        {/* Lap bars */}
        <div className="absolute inset-[3px] flex items-end">
          {bars.map(({ xPct: _x, wPct, hPct, power, opacity, lap }) => {
            const isHovered = hoveredLap === lap.lapNumber;
            return (
              <div
                key={lap.lapNumber}
                className="relative flex-shrink-0 h-full"
                style={{ width: `${wPct}%` }}
                onMouseEnter={() => setHoveredLap(lap.lapNumber)}
                onMouseLeave={() => setHoveredLap(null)}
              >
                {/* Bar fill */}
                <div
                  className="absolute bottom-0 w-full transition-opacity duration-150 rounded-sm"
                  style={{
                    height: `${hPct}%`,
                    backgroundColor: "#EF4444",
                    opacity: isHovered ? Math.min(opacity + 0.15, 1) : opacity,
                    borderRight: "2px solid var(--color-background-subtle)",
                  }}
                />

                {/* Tooltip */}
                {isHovered && (
                  <div
                    className="absolute bottom-full mb-2 left-1/2 -translate-x-1/2 z-20 rounded border border-border bg-background-subtle px-3 py-2 shadow-lg pointer-events-none"
                    style={{ whiteSpace: "nowrap", fontSize: 12 }}
                  >
                    <p style={{ color: "var(--color-foreground-muted)", marginBottom: 4 }}>
                      Lap {lap.lapNumber}
                    </p>
                    <p style={{ color: "var(--color-foreground)" }}>
                      Time: <span style={{ color: "var(--color-foreground)" }}>{formatDuration(lap.duration)}</span>
                    </p>
                    {power > 0 && (
                      <p style={{ color: "#F97316", marginTop: 2 }}>
                        Power: <span style={{ color: "var(--color-foreground)" }}>{power} W</span>
                      </p>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

// Opacity scales from 0.25 (near 0%) to 1.0 (at 120%+)
function effortStyle(pct: number): { background: string; opacity: number } {
  const capped = Math.min(pct, 120);
  const opacity = 0.25 + (capped / 120) * 0.75;
  return { background: "#EF4444", opacity };
}

function EffortBar({ power, ftp }: { power?: number; ftp?: number }) {
  const [hovered, setHovered] = useState(false);

  if (!power || !ftp || ftp === 0) {
    return <span className="font-mono text-sm text-foreground-muted">—</span>;
  }

  const ftpPct = (power / ftp) * 100;
  const { background, opacity } = effortStyle(ftpPct);
  const barWidth = Math.min((ftpPct / 120) * 100, 100);

  return (
    <div
      className="relative"
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <div className="w-28 overflow-hidden rounded-full bg-background" style={{ height: 4 }}>
        <div
          className="h-full rounded-full transition-all duration-300"
          style={{ width: `${barWidth}%`, background, opacity }}
        />
      </div>
      {hovered && (
        <div
          className="absolute bottom-full left-0 z-10 mb-1.5 rounded border border-border bg-background-subtle px-2.5 py-1.5 text-xs shadow-lg"
          style={{ whiteSpace: "nowrap" }}
        >
          <span className="font-mono font-semibold text-foreground">{power} W</span>
          <span className="mx-1.5 text-foreground-muted">·</span>
          <span className="font-mono font-semibold" style={{ color: "#EF4444", opacity: 1 }}>
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
      <LapOverviewChart laps={laps} activityId={activityId} />
      <div className="overflow-x-auto rounded-lg border border-border">
        <table className="w-full min-w-[900px]">
          <thead className="bg-background-subtle">
            <tr>
              <th className="px-3 py-2.5 text-left font-medium text-foreground-muted">Lap</th>
              <th className="px-3 py-2.5 text-right font-medium text-foreground-muted">Distance</th>
              <th className="px-3 py-2.5 text-right font-medium text-foreground-muted">Time</th>
              <th className="px-3 py-2.5 text-right font-medium text-foreground-muted">Elevation</th>
              <th className="px-3 py-2.5 text-right font-medium text-foreground-muted">Speed</th>
              <th className="px-3 py-2.5 text-right font-medium text-foreground-muted">HR</th>
              <th className="px-3 py-2.5 text-right font-medium text-foreground-muted">Cad</th>
              <th className="px-3 py-2.5 text-right font-medium text-foreground-muted">Power</th>
              <th className="px-3 py-2.5 text-left font-medium text-foreground-muted">Effort Intensity</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {laps.map((lap) => (
              <tr key={lap.lapNumber} className="hover:bg-background-subtle/50">
                <td className="px-3 py-2 font-mono font-medium text-foreground">
                  {lap.lapNumber}
                </td>
                <td className="px-3 py-2 text-right font-mono text-foreground">
                  {formatDistance(lap.distance)}
                </td>
                <td className="px-3 py-2 text-right font-mono text-foreground">
                  {formatDuration(lap.duration)}
                </td>
                <td className="px-3 py-2 text-right font-mono text-elevation">
                  {lap.ascent ? `+${lap.ascent.toFixed(0)} m` : "—"}
                </td>
                <td className="px-3 py-2 text-right font-mono text-speed">
                  {lap.avgSpeed ? formatSpeed(lap.avgSpeed) : "—"}
                </td>
                <td className="px-3 py-2 text-right font-mono text-heart-rate">
                  {lap.avgHeartRate ?? "—"}
                </td>
                <td className="px-3 py-2 text-right font-mono text-cadence">
                  {lap.avgCadence ?? "—"}
                </td>
                <td className="px-3 py-2 text-right font-mono text-power">
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
