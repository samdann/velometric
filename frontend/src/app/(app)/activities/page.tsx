"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { PageHeader } from "@/components/layout";
import { api, Activity } from "@/lib/api";

function formatDuration(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  if (h > 0) {
    return `${h}h ${m}m`;
  }
  return `${m}m`;
}

function formatDistance(meters: number): string {
  const km = meters / 1000;
  return `${km.toFixed(1)} km`;
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export default function ActivitiesPage() {
  const [activities, setActivities] = useState<Activity[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchActivities() {
      try {
        const data = await api.getActivities();
        setActivities(data || []);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load activities");
      } finally {
        setLoading(false);
      }
    }
    fetchActivities();
  }, []);

  return (
    <div>
      <PageHeader
        title="Activities"
        description="View and analyze your rides"
      />
      <div className="p-6">
        {loading && (
          <div className="text-center text-foreground-muted">Loading...</div>
        )}

        {error && (
          <div className="rounded-lg bg-heart-rate/10 p-4 text-center">
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

        {!loading && activities.length > 0 && (
          <div className="space-y-3">
            {activities.map((activity) => (
              <Link
                key={activity.id}
                href={`/activities/${activity.id}`}
                className="block rounded-lg border border-border bg-background-subtle p-4 transition-colors hover:border-border-hover"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium text-foreground">
                      {activity.name}
                    </h3>
                    <p className="mt-1 text-sm text-foreground-muted">
                      {formatDate(activity.startTime)}
                    </p>
                  </div>
                  <div className="flex gap-6 text-right">
                    <div>
                      <p className="font-mono text-sm text-foreground">
                        {formatDistance(activity.distance)}
                      </p>
                      <p className="text-xs text-foreground-muted">Distance</p>
                    </div>
                    <div>
                      <p className="font-mono text-sm text-foreground">
                        {formatDuration(activity.duration)}
                      </p>
                      <p className="text-xs text-foreground-muted">Duration</p>
                    </div>
                    {activity.avgPower && (
                      <div>
                        <p className="font-mono text-sm text-power">
                          {activity.avgPower}w
                        </p>
                        <p className="text-xs text-foreground-muted">Avg Power</p>
                      </div>
                    )}
                    {activity.normalizedPower && (
                      <div>
                        <p className="font-mono text-sm text-power">
                          {activity.normalizedPower}w
                        </p>
                        <p className="text-xs text-foreground-muted">NP</p>
                      </div>
                    )}
                    {activity.tss && (
                      <div>
                        <p className="font-mono text-sm text-foreground">
                          {activity.tss.toFixed(0)}
                        </p>
                        <p className="text-xs text-foreground-muted">TSS</p>
                      </div>
                    )}
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
