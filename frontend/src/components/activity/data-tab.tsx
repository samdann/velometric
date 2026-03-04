"use client";

import { useEffect, useState } from "react";

interface Record {
  timestamp: string;
  distance?: number;
  altitude?: number;
  power?: number;
  heartRate?: number;
  cadence?: number;
  speed?: number;
  temperature?: number;
}

interface DataTabProps {
  activityId: string;
}

function formatTime(timestamp: string, startTime: string): string {
  const elapsed = (new Date(timestamp).getTime() - new Date(startTime).getTime()) / 1000;
  const h = Math.floor(elapsed / 3600);
  const m = Math.floor((elapsed % 3600) / 60);
  const s = Math.floor(elapsed % 60);
  if (h > 0) return `${h.toString().padStart(2, "0")}:${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
  return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
}

export function DataTab({ activityId }: DataTabProps) {
  const [records, setRecords] = useState<Record[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(0);
  const pageSize = 100;

  useEffect(() => {
    async function fetchRecords() {
      try {
        const response = await fetch(
          `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8081"}/api/activities/${activityId}/records`
        );
        if (!response.ok) throw new Error("Failed to fetch records");
        const data = await response.json();
        setRecords(data || []);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load data");
      } finally {
        setLoading(false);
      }
    }
    fetchRecords();
  }, [activityId]);

  if (loading) {
    return (
      <div className="mt-6 flex h-32 items-center justify-center">
        <p className="text-foreground-muted">Loading data...</p>
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

  if (records.length === 0) {
    return (
      <div className="mt-6 rounded-lg border border-border bg-background-subtle p-6 text-center">
        <p className="text-foreground-muted">No data records available</p>
      </div>
    );
  }

  const startTime = records[0]?.timestamp;
  const totalPages = Math.ceil(records.length / pageSize);
  const displayedRecords = records.slice(page * pageSize, (page + 1) * pageSize);

  return (
    <div className="mt-6 space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-foreground-muted">
          {records.length.toLocaleString()} data points
        </p>
        {totalPages > 1 && (
          <div className="flex items-center gap-2">
            <button
              onClick={() => setPage((p) => Math.max(0, p - 1))}
              disabled={page === 0}
              className="rounded border border-border px-3 py-1 text-sm disabled:opacity-50"
            >
              Prev
            </button>
            <span className="text-sm text-foreground-muted">
              {page + 1} / {totalPages}
            </span>
            <button
              onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
              disabled={page === totalPages - 1}
              className="rounded border border-border px-3 py-1 text-sm disabled:opacity-50"
            >
              Next
            </button>
          </div>
        )}
      </div>

      <div className="overflow-x-auto rounded-lg border border-border">
        <table className="w-full min-w-[700px]">
          <thead className="bg-background-subtle">
            <tr>
              <th className="px-3 py-2 text-left font-medium text-foreground-muted">Time</th>
              <th className="px-3 py-2 text-right font-medium text-foreground-muted">Dist (m)</th>
              <th className="px-3 py-2 text-right font-medium text-foreground-muted">Alt (m)</th>
              <th className="px-3 py-2 text-right font-medium text-foreground-muted">Power</th>
              <th className="px-3 py-2 text-right font-medium text-foreground-muted">HR</th>
              <th className="px-3 py-2 text-right font-medium text-foreground-muted">Cad</th>
              <th className="px-3 py-2 text-right font-medium text-foreground-muted">Speed</th>
              <th className="px-3 py-2 text-right font-medium text-foreground-muted">Temp</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border text-sm">
            {displayedRecords.map((rec, idx) => (
              <tr key={idx} className="hover:bg-background-subtle/50">
                <td className="px-3 py-1.5 font-mono">{formatTime(rec.timestamp, startTime)}</td>
                <td className="px-3 py-1.5 text-right font-mono">
                  {rec.distance?.toFixed(0) ?? "-"}
                </td>
                <td className="px-3 py-1.5 text-right font-mono text-elevation">
                  {rec.altitude?.toFixed(1) ?? "-"}
                </td>
                <td className="px-3 py-1.5 text-right font-mono text-power">
                  {rec.power ?? "-"}
                </td>
                <td className="px-3 py-1.5 text-right font-mono text-heart-rate">
                  {rec.heartRate ?? "-"}
                </td>
                <td className="px-3 py-1.5 text-right font-mono text-cadence">
                  {rec.cadence ?? "-"}
                </td>
                <td className="px-3 py-1.5 text-right font-mono text-speed">
                  {rec.speed ? (rec.speed * 3.6).toFixed(1) : "-"}
                </td>
                <td className="px-3 py-1.5 text-right font-mono">
                  {rec.temperature?.toFixed(1) ?? "-"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
