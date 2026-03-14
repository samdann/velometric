"use client";

import { useEffect, useRef, useState } from "react";
import { api, type FeedActivity } from "@/lib/api";
import { ActivityFeedCard } from "@/components/dashboard/activity-feed-card";

const LIMIT = 5;

export default function DashboardPage() {
  const [activities, setActivities] = useState<FeedActivity[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const sentinelRef = useRef<HTMLDivElement>(null);
  const pageRef = useRef(1);

  // Keep a ref to the latest load-more callback to avoid stale closures in the observer
  const loadMoreFn = useRef<() => void>(() => {});
  useEffect(() => {
    loadMoreFn.current = () => {
      if (loadingMore || loading || activities.length >= total) return;
      const nextPage = pageRef.current + 1;
      pageRef.current = nextPage;
      setLoadingMore(true);
      new Promise((r) => setTimeout(r, 1000))
        .then(() => api.getFeed(nextPage, LIMIT))
        .then((data) => {
          setActivities((prev) => [...prev, ...data.activities]);
          setTotal(data.total);
        })
        .catch((err) => setError(err.message))
        .finally(() => setLoadingMore(false));
    };
  }, [loadingMore, loading, activities.length, total]);

  // Initial load
  useEffect(() => {
    api
      .getFeed(1, LIMIT)
      .then((data) => {
        setActivities(data.activities);
        setTotal(data.total);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  // Attach sentinel observer once initial load is done
  useEffect(() => {
    if (loading) return;
    const sentinel = sentinelRef.current;
    if (!sentinel) return;
    const observer = new IntersectionObserver(() => loadMoreFn.current(), {
      rootMargin: "300px",
    });
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [loading]);

  const hasMore = activities.length < total;

  return (
    <div className="p-6">
      <div className="mx-auto space-y-4" style={{ width: 500 }}>
        {loading && (
          <p className="text-sm text-foreground-muted">Loading...</p>
        )}
        {error && <p className="text-sm text-red-500">{error}</p>}
        {!loading && !error && activities.length === 0 && (
          <p className="text-sm text-foreground-muted">
            No activities yet. Upload a FIT file to get started.
          </p>
        )}
        {activities.map((activity) => (
          <ActivityFeedCard key={activity.id} activity={activity} />
        ))}
        <div ref={sentinelRef} />
        {loadingMore && (
          <p className="text-sm text-foreground-muted">Loading more...</p>
        )}
        {!loading && !hasMore && activities.length > 0 && (
          <p className="py-4 text-center text-xs text-foreground-muted">
            All caught up
          </p>
        )}
      </div>
    </div>
  );
}
