"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import Link from "next/link";
import { PageHeader } from "@/components/layout";
import { api, Activity, ActivityFilters } from "@/lib/api";

const PAGE_SIZE_OPTIONS = [10, 25, 50] as const;
type PageSize = (typeof PAGE_SIZE_OPTIONS)[number];

type SortKey = "date" | "distance" | "duration" | "elevation";
type SortOrder = "asc" | "desc";


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

function TrashIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="15"
      height="15"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <polyline points="3 6 5 6 21 6" />
      <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6" />
      <path d="M10 11v6M14 11v6" />
      <path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2" />
    </svg>
  );
}

function SortIcon({ direction }: { direction: SortOrder | null }) {
  if (!direction) {
    return (
      <svg width="10" height="10" viewBox="0 0 10 10" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" className="opacity-30">
        <path d="M5 2v6M2 5l3-3 3 3" />
        <path d="M2 5l3 3 3-3" />
      </svg>
    );
  }
  if (direction === "asc") {
    return (
      <svg width="10" height="10" viewBox="0 0 10 10" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
        <path d="M5 8V2M2 5l3-3 3 3" />
      </svg>
    );
  }
  return (
    <svg width="10" height="10" viewBox="0 0 10 10" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
      <path d="M5 2v6M2 5l3 3 3-3" />
    </svg>
  );
}

interface DeleteDialogProps {
  activity: Activity;
  onConfirm: () => void;
  onCancel: () => void;
  deleting: boolean;
}

function DeleteDialog({ activity, onConfirm, onCancel, deleting }: DeleteDialogProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onCancel}
      />
      <div className="relative w-full max-w-sm rounded-xl border border-border bg-background p-6 shadow-xl">
        <h2 className="text-base font-semibold text-foreground">Delete Activity</h2>
        <p className="mt-2 text-sm text-foreground-muted">
          Are you sure you want to delete{" "}
          <span className="font-medium text-foreground">{activity.name}</span>?
          This will permanently remove the activity and all its data.
        </p>
        <div className="mt-6 flex gap-3 justify-end">
          <button
            onClick={onCancel}
            disabled={deleting}
            className="rounded-lg border border-border px-4 py-2 text-sm text-foreground-muted hover:border-border-hover hover:text-foreground disabled:opacity-50 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            disabled={deleting}
            className="rounded-lg bg-heart-rate px-4 py-2 text-sm font-medium text-white hover:bg-heart-rate/80 disabled:opacity-50 transition-colors"
          >
            {deleting ? "Deleting…" : "Delete"}
          </button>
        </div>
      </div>
    </div>
  );
}

export default function ActivitiesPage() {
  const [activities, setActivities] = useState<Activity[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState<PageSize>(10);
  const [sports, setSports] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [toDelete, setToDelete] = useState<Activity | null>(null);
  const [deleting, setDeleting] = useState(false);

  // Filter state
  const [search, setSearch] = useState("");
  const [sport, setSport] = useState("");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [distMin, setDistMin] = useState("");
  const [distMax, setDistMax] = useState("");

  // Sort state
  const [sortBy, setSortBy] = useState<SortKey>("date");
  const [sortOrder, setSortOrder] = useState<SortOrder>("desc");

  // Debounce search
  const searchDebounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [debouncedSearch, setDebouncedSearch] = useState("");

  useEffect(() => {
    if (searchDebounceRef.current) clearTimeout(searchDebounceRef.current);
    searchDebounceRef.current = setTimeout(() => setDebouncedSearch(search), 300);
    return () => {
      if (searchDebounceRef.current) clearTimeout(searchDebounceRef.current);
    };
  }, [search]);

  const activeFilterCount = [
    debouncedSearch,
    sport,
    dateFrom,
    dateTo,
    distMin,
    distMax,
  ].filter(Boolean).length;

  const filters: ActivityFilters = {
    q: debouncedSearch || undefined,
    sport: sport || undefined,
    dateFrom: dateFrom || undefined,
    dateTo: dateTo || undefined,
    distMin: distMin ? parseFloat(distMin) : undefined,
    distMax: distMax ? parseFloat(distMax) : undefined,
    sortBy,
    sortOrder,
  };

  const fetchActivities = useCallback(async (p: number, l: PageSize, f: ActivityFilters) => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.getActivities(p, l, f);
      setActivities(data.activities ?? []);
      setTotal(data.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load activities");
    } finally {
      setLoading(false);
    }
  }, []);

  // Reset to page 1 whenever filters or sort change
  useEffect(() => {
    setPage(1);
  }, [debouncedSearch, sport, dateFrom, dateTo, distMin, distMax, sortBy, sortOrder]);

  useEffect(() => {
    fetchActivities(page, limit, filters);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, limit, debouncedSearch, sport, dateFrom, dateTo, distMin, distMax, sortBy, sortOrder]);

  useEffect(() => {
    api.getSports().then(setSports).catch(() => {});
  }, []);

  const totalPages = Math.max(1, Math.ceil(total / limit));

  function handleLimitChange(newLimit: PageSize) {
    setLimit(newLimit);
    setPage(1);
  }

  function handleSort(key: SortKey) {
    if (sortBy === key) {
      setSortOrder((o) => (o === "asc" ? "desc" : "asc"));
    } else {
      setSortBy(key);
      setSortOrder(key === "date" ? "desc" : "asc");
    }
  }

  function clearFilters() {
    setSearch("");
    setSport("");
    setDateFrom("");
    setDateTo("");
    setDistMin("");
    setDistMax("");
  }

  async function handleDeleteConfirm() {
    if (!toDelete) return;
    setDeleting(true);
    try {
      await api.deleteActivity(toDelete.id);
      setToDelete(null);
      const newTotal = total - 1;
      const maxPage = Math.max(1, Math.ceil(newTotal / limit));
      const nextPage = Math.min(page, maxPage);
      if (nextPage !== page) {
        setPage(nextPage);
      } else {
        fetchActivities(nextPage, limit, filters);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete activity");
      setToDelete(null);
    } finally {
      setDeleting(false);
    }
  }

  const SortableHeader = ({
    label,
    colKey,
    align = "right",
  }: {
    label: string;
    colKey: SortKey;
    align?: "left" | "right";
  }) => (
    <th
      className={`px-4 py-3 font-medium text-foreground-muted uppercase tracking-wider cursor-pointer select-none group ${align === "right" ? "text-right" : "text-left"}`}
      onClick={() => handleSort(colKey)}
    >
      <span className={`inline-flex items-center gap-1.5 ${align === "right" ? "flex-row-reverse" : ""}`}>
        <span className="group-hover:text-foreground transition-colors">{label}</span>
        <span className={sortBy === colKey ? "text-primary" : "text-foreground-muted"}>
          <SortIcon direction={sortBy === colKey ? sortOrder : null} />
        </span>
      </span>
    </th>
  );

  return (
    <div>
      <PageHeader title="My Activities" description="Your ride history" />

      {toDelete && (
        <DeleteDialog
          activity={toDelete}
          onConfirm={handleDeleteConfirm}
          onCancel={() => setToDelete(null)}
          deleting={deleting}
        />
      )}

      <div className="p-6">
        {error && (
          <div className="rounded-lg bg-heart-rate/10 p-4 text-center mb-4">
            <p className="text-sm text-heart-rate">{error}</p>
          </div>
        )}

        {/* Filter bar */}
        <div className="mb-4 rounded-lg border border-border bg-background-subtle p-3">
          <div className="flex flex-wrap items-center gap-2">
            {/* Search */}
            <div className="relative flex-1 min-w-[160px]">
              <svg className="absolute left-2.5 top-1/2 -translate-y-1/2 text-foreground-muted/50" width="13" height="13" viewBox="0 0 13 13" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round">
                <circle cx="5.5" cy="5.5" r="4" />
                <path d="M8.5 8.5L11 11" />
              </svg>
              <input
                type="text"
                placeholder="Search activities…"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-full rounded-md border border-border bg-background pl-8 pr-3 py-1.5 text-sm text-foreground placeholder:text-foreground-muted/40 focus:outline-none focus:border-primary transition-colors"
              />
            </div>

            {/* Divider */}
            <div className="h-6 w-px bg-border hidden sm:block" />

            {/* Sport */}
            <select
              value={sport}
              onChange={(e) => setSport(e.target.value)}
              className="rounded-md border border-border bg-background px-2.5 py-1.5 text-sm text-foreground focus:outline-none focus:border-primary cursor-pointer transition-colors hover:border-border-hover"
            >
              <option value="">All sports</option>
              {sports.map((s) => (
                <option key={s} value={s}>{formatSport(s)}</option>
              ))}
            </select>

            {/* Divider */}
            <div className="h-6 w-px bg-border hidden sm:block" />

            {/* Date range */}
            <div className="flex items-center gap-1.5">
              <input
                type="date"
                value={dateFrom}
                onChange={(e) => setDateFrom(e.target.value)}
                title="From date"
                className="rounded-md border border-border bg-background px-2 py-1.5 text-sm text-foreground focus:outline-none focus:border-primary transition-colors [color-scheme:dark]"
              />
              <span className="text-xs text-foreground-muted/50">–</span>
              <input
                type="date"
                value={dateTo}
                onChange={(e) => setDateTo(e.target.value)}
                title="To date"
                className="rounded-md border border-border bg-background px-2 py-1.5 text-sm text-foreground focus:outline-none focus:border-primary transition-colors [color-scheme:dark]"
              />
            </div>

            {/* Divider */}
            <div className="h-6 w-px bg-border hidden sm:block" />

            {/* Distance */}
            <div className="flex items-center gap-1.5">
              <input
                type="number"
                placeholder="Min km"
                min={0}
                value={distMin}
                onChange={(e) => setDistMin(e.target.value)}
                className="w-20 rounded-md border border-border bg-background px-2 py-1.5 text-sm text-foreground placeholder:text-foreground-muted/40 focus:outline-none focus:border-primary transition-colors"
              />
              <span className="text-xs text-foreground-muted/50">–</span>
              <input
                type="number"
                placeholder="Max km"
                min={0}
                value={distMax}
                onChange={(e) => setDistMax(e.target.value)}
                className="w-20 rounded-md border border-border bg-background px-2 py-1.5 text-sm text-foreground placeholder:text-foreground-muted/40 focus:outline-none focus:border-primary transition-colors"
              />
            </div>

            {/* Clear */}
            {activeFilterCount > 0 && (
              <button
                onClick={clearFilters}
                className="ml-auto flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-xs text-foreground-muted hover:text-foreground hover:bg-surface-hover transition-colors"
              >
                <svg width="11" height="11" viewBox="0 0 11 11" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round">
                  <path d="M1 1l9 9M10 1L1 10" />
                </svg>
                Clear
                <span className="rounded-full bg-primary/20 text-primary px-1.5 py-0.5 text-[10px] font-semibold leading-none">
                  {activeFilterCount}
                </span>
              </button>
            )}
          </div>
        </div>

        {!loading && !error && activities.length === 0 && (
          <div className="rounded-lg border border-border bg-background-subtle p-8 text-center">
            <p className="text-foreground-muted">
              {activeFilterCount > 0 ? "No activities match your filters." : "No activities yet."}
            </p>
            {activeFilterCount > 0 ? (
              <button
                onClick={clearFilters}
                className="mt-2 text-sm text-primary hover:underline"
              >
                Clear filters
              </button>
            ) : (
              <Link
                href="/upload"
                className="mt-2 inline-block text-sm text-primary hover:underline"
              >
                Upload your first ride
              </Link>
            )}
          </div>
        )}

        {(loading || activities.length > 0) && (
          <>
            <div className="rounded-lg border border-border overflow-hidden">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-border bg-background-subtle text-xs">
                    <th className="px-4 py-3 text-left font-medium text-foreground-muted uppercase tracking-wider">Sport</th>
                    <th className="px-4 py-3 text-left font-medium text-foreground-muted uppercase tracking-wider">Title</th>
                    <SortableHeader label="Date" colKey="date" align="left" />
                    <SortableHeader label="Distance" colKey="distance" />
                    <SortableHeader label="Time" colKey="duration" />
                    <SortableHeader label="Elevation" colKey="elevation" />
                    <th className="px-4 py-3 text-right font-medium text-foreground-muted uppercase tracking-wider">Avg Power</th>
                    <th className="px-4 py-3 w-10" />
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {loading
                    ? Array.from({ length: limit }).map((_, i) => (
                        <tr key={i} className="animate-pulse">
                          {Array.from({ length: 8 }).map((_, j) => (
                            <td key={j} className="px-4 py-3">
                              <div className="h-4 rounded bg-background-subtle" />
                            </td>
                          ))}
                        </tr>
                      ))
                    : activities.map((activity) => (
                        <tr
                          key={activity.id}
                          className="hover:bg-background-subtle transition-colors group"
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
                          <td className="px-4 py-3 text-sm text-foreground-muted">
                            {formatDate(activity.startTime)}
                          </td>
                          <td className="px-4 py-3 text-right font-mono text-sm text-foreground">
                            {formatDistance(activity.distance)}
                          </td>
                          <td className="px-4 py-3 text-right font-mono text-sm text-foreground">
                            {formatDuration(activity.duration)}
                          </td>
                          <td className="px-4 py-3 text-right font-mono text-sm text-foreground">
                            {formatElevation(activity.elevationGain)}
                          </td>
                          <td className="px-4 py-3 text-right font-mono text-sm">
                            {activity.avgPower != null ? (
                              <span className="text-power">{activity.avgPower}w</span>
                            ) : (
                              <span className="text-foreground-muted">—</span>
                            )}
                          </td>
                          <td className="px-4 py-3 text-center">
                            <button
                              onClick={() => setToDelete(activity)}
                              className="text-foreground-muted opacity-0 group-hover:opacity-100 hover:text-heart-rate transition-all"
                              title="Delete activity"
                            >
                              <TrashIcon />
                            </button>
                          </td>
                        </tr>
                      ))}
                </tbody>
              </table>
            </div>

            {/* Pagination controls */}
            <div className="mt-6 flex items-center justify-between gap-4">
              {/* Left: rows per page + record range */}
              <div className="flex items-center gap-3 text-xs text-foreground-muted font-mono">
                <label className="flex items-center gap-2">
                  <span>Rows</span>
                  <select
                    value={limit}
                    onChange={(e) => handleLimitChange(Number(e.target.value) as PageSize)}
                    className="bg-surface border border-border rounded-md px-2 py-1 text-xs font-mono text-foreground focus:outline-none focus:border-primary cursor-pointer transition-colors hover:border-border-hover"
                  >
                    {PAGE_SIZE_OPTIONS.map((size) => (
                      <option key={size} value={size}>{size}</option>
                    ))}
                  </select>
                </label>
                {!loading && total > 0 && (
                  <span className="text-foreground-muted/60 tabular-nums">
                    {Math.min((page - 1) * limit + 1, total)}–{Math.min(page * limit, total)}{" "}
                    <span className="text-foreground-muted/40">of</span> {total}
                  </span>
                )}
              </div>

              {/* Right: page nav */}
              <div className="flex items-center gap-1">
                <button
                  onClick={() => setPage(1)}
                  disabled={page === 1 || loading}
                  title="First page"
                  className="h-7 w-7 flex items-center justify-center rounded-md text-foreground-muted hover:text-foreground hover:bg-surface-hover disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                >
                  <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M9 9L5 6l4-3"/><line x1="3" y1="3" x2="3" y2="9"/>
                  </svg>
                </button>
                <button
                  onClick={() => setPage((p) => p - 1)}
                  disabled={page === 1 || loading}
                  title="Previous page"
                  className="h-7 w-7 flex items-center justify-center rounded-md text-foreground-muted hover:text-foreground hover:bg-surface-hover disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                >
                  <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M7.5 9L4 6l3.5-3"/>
                  </svg>
                </button>

                <div className="flex items-center gap-1 mx-1">
                  {Array.from({ length: totalPages }, (_, i) => i + 1)
                    .filter((p) => p === 1 || p === totalPages || Math.abs(p - page) <= 1)
                    .reduce<(number | "…")[]>((acc, p, i, arr) => {
                      if (i > 0 && p - (arr[i - 1] as number) > 1) acc.push("…");
                      acc.push(p);
                      return acc;
                    }, [])
                    .map((item, i) =>
                      item === "…" ? (
                        <span key={`ellipsis-${i}`} className="w-7 text-center text-xs text-foreground-muted/40 font-mono">…</span>
                      ) : (
                        <button
                          key={item}
                          onClick={() => setPage(item as number)}
                          disabled={loading}
                          className={`h-7 min-w-7 px-1 rounded-md text-xs font-mono transition-colors disabled:cursor-not-allowed ${
                            page === item
                              ? "bg-primary/15 text-primary font-semibold"
                              : "text-foreground-muted hover:text-foreground hover:bg-surface-hover"
                          }`}
                        >
                          {item}
                        </button>
                      )
                    )}
                </div>

                <button
                  onClick={() => setPage((p) => p + 1)}
                  disabled={page === totalPages || loading}
                  title="Next page"
                  className="h-7 w-7 flex items-center justify-center rounded-md text-foreground-muted hover:text-foreground hover:bg-surface-hover disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                >
                  <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M4.5 9L8 6l-3.5-3"/>
                  </svg>
                </button>
                <button
                  onClick={() => setPage(totalPages)}
                  disabled={page === totalPages || loading}
                  title="Last page"
                  className="h-7 w-7 flex items-center justify-center rounded-md text-foreground-muted hover:text-foreground hover:bg-surface-hover disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                >
                  <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M3 9l4-3-4-3"/><line x1="9" y1="3" x2="9" y2="9"/>
                  </svg>
                </button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
