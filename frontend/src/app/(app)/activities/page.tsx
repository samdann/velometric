"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { PageHeader } from "@/components/layout";
import { api, Activity } from "@/lib/api";

const PAGE_SIZE_OPTIONS = [10, 25, 50] as const;
type PageSize = (typeof PAGE_SIZE_OPTIONS)[number];

function formatDuration(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  if (h > 0) return `${h}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
  return `${m}:${String(s).padStart(2, "0")}`;
}

function formatDistance(meters: number): string {
  return `${(meters / 1000).toFixed(2)} km`;
}

function formatElevation(meters: number): string {
  return `${Math.round(meters)} m`;
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}


function formatSport(sport: string): string {
  return sport.charAt(0).toUpperCase() + sport.slice(1).toLowerCase();
}

function SportBadge({ sport }: { sport: string }) {
  const label = formatSport(sport);
  return (
    <span className="inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium bg-background border border-border text-foreground-muted">
      {label}
    </span>
  );
}

export default function ActivitiesPage() {
  const [activities, setActivities] = useState<Activity[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState<PageSize>(25);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchActivities = useCallback(async (p: number, l: PageSize) => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.getActivities(p, l);
      setActivities(data.activities ?? []);
      setTotal(data.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load activities");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchActivities(page, limit);
  }, [page, limit, fetchActivities]);

  const totalPages = Math.max(1, Math.ceil(total / limit));

  function handleLimitChange(newLimit: PageSize) {
    setLimit(newLimit);
    setPage(1);
  }

  return (
    <div>
      <PageHeader
        title="My Activities"
        description="Your ride history"
      />
      <div className="p-6">
        {error && (
          <div className="rounded-lg bg-heart-rate/10 p-4 text-center mb-4">
            <p className="text-sm text-heart-rate">{error}</p>
          </div>
        )}

        {!loading && !error && activities.length === 0 && (
          <div className="rounded-lg border border-border bg-background-subtle p-8 text-center">
            <p className="text-foreground-muted">No activities yet.</p>
            <Link
              href="/upload"
              className="mt-2 inline-block text-sm text-primary hover:underline"
            >
              Upload your first ride
            </Link>
          </div>
        )}

        {(loading || activities.length > 0) && (
          <>
            {/* Table */}
            <div className="rounded-lg border border-border overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border bg-background-subtle">
                    <th className="px-4 py-3 text-left text-xs font-medium text-foreground-muted uppercase tracking-wider">Sport</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-foreground-muted uppercase tracking-wider">Title</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-foreground-muted uppercase tracking-wider">Date</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-foreground-muted uppercase tracking-wider">Distance</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-foreground-muted uppercase tracking-wider">Time</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-foreground-muted uppercase tracking-wider">Elevation</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-foreground-muted uppercase tracking-wider">Avg Power</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {loading
                    ? Array.from({ length: limit }).map((_, i) => (
                        <tr key={i} className="animate-pulse">
                          {Array.from({ length: 7 }).map((_, j) => (
                            <td key={j} className="px-4 py-3">
                              <div className="h-4 rounded bg-background-subtle" />
                            </td>
                          ))}
                        </tr>
                      ))
                    : activities.map((activity) => (
                        <tr
                          key={activity.id}
                          className="hover:bg-background-subtle transition-colors"
                        >
                          <td className="px-4 py-3">
                            <SportBadge sport={activity.sport} />
                          </td>
                          <td className="px-4 py-3">
                            <Link
                              href={`/activities/${activity.id}`}
                              className="font-medium text-foreground hover:text-primary transition-colors"
                            >
                              {activity.name}
                            </Link>
                          </td>
                          <td className="px-4 py-3 text-foreground-muted">
                            {formatDate(activity.startTime)}
                          </td>
                          <td className="px-4 py-3 text-right font-mono text-foreground">
                            {formatDistance(activity.distance)}
                          </td>
                          <td className="px-4 py-3 text-right font-mono text-foreground">
                            {formatDuration(activity.duration)}
                          </td>
                          <td className="px-4 py-3 text-right font-mono text-foreground">
                            {formatElevation(activity.elevationGain)}
                          </td>
                          <td className="px-4 py-3 text-right font-mono">
                            {activity.avgPower != null ? (
                              <span className="text-power">{activity.avgPower}w</span>
                            ) : (
                              <span className="text-foreground-muted">—</span>
                            )}
                          </td>
                        </tr>
                      ))}
                </tbody>
              </table>
            </div>

            {/* Pagination controls */}
            <div className="mt-4 flex items-center justify-between">
              <div className="flex items-center gap-2 text-sm text-foreground-muted">
                <span>Show</span>
                {PAGE_SIZE_OPTIONS.map((size) => (
                  <button
                    key={size}
                    onClick={() => handleLimitChange(size)}
                    className={`px-2 py-1 rounded text-xs font-mono border transition-colors ${
                      limit === size
                        ? "border-primary text-primary bg-primary/10"
                        : "border-border text-foreground-muted hover:border-border-hover"
                    }`}
                  >
                    {size}
                  </button>
                ))}
                <span>per page</span>
                {!loading && (
                  <span className="ml-3">
                    {Math.min((page - 1) * limit + 1, total)}–{Math.min(page * limit, total)} of {total}
                  </span>
                )}
              </div>

              <div className="flex items-center gap-1">
                <button
                  onClick={() => setPage(1)}
                  disabled={page === 1 || loading}
                  className="px-2 py-1 rounded border border-border text-xs text-foreground-muted hover:border-border-hover disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                >
                  «
                </button>
                <button
                  onClick={() => setPage((p) => p - 1)}
                  disabled={page === 1 || loading}
                  className="px-2 py-1 rounded border border-border text-xs text-foreground-muted hover:border-border-hover disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                >
                  ‹
                </button>

                {/* Page numbers */}
                {Array.from({ length: totalPages }, (_, i) => i + 1)
                  .filter((p) => p === 1 || p === totalPages || Math.abs(p - page) <= 1)
                  .reduce<(number | "…")[]>((acc, p, i, arr) => {
                    if (i > 0 && p - (arr[i - 1] as number) > 1) acc.push("…");
                    acc.push(p);
                    return acc;
                  }, [])
                  .map((item, i) =>
                    item === "…" ? (
                      <span key={`ellipsis-${i}`} className="px-2 py-1 text-xs text-foreground-muted">…</span>
                    ) : (
                      <button
                        key={item}
                        onClick={() => setPage(item as number)}
                        disabled={loading}
                        className={`px-2 py-1 rounded border text-xs font-mono transition-colors disabled:cursor-not-allowed ${
                          page === item
                            ? "border-primary text-primary bg-primary/10"
                            : "border-border text-foreground-muted hover:border-border-hover"
                        }`}
                      >
                        {item}
                      </button>
                    )
                  )}

                <button
                  onClick={() => setPage((p) => p + 1)}
                  disabled={page === totalPages || loading}
                  className="px-2 py-1 rounded border border-border text-xs text-foreground-muted hover:border-border-hover disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                >
                  ›
                </button>
                <button
                  onClick={() => setPage(totalPages)}
                  disabled={page === totalPages || loading}
                  className="px-2 py-1 rounded border border-border text-xs text-foreground-muted hover:border-border-hover disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                >
                  »
                </button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
