"use client";

import { useEffect, useState } from "react";
import { api, type FeedActivity } from "@/lib/api";
import { ActivityFeedCard } from "@/components/dashboard/activity-feed-card";

export default function DashboardPage() {
  const [activities, setActivities] = useState<FeedActivity[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .getFeed(1, 25)
      .then((data) => setActivities(data.activities))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="p-6">
      <div className="mx-auto space-y-4" style={{ width: 500 }}>
        {loading && (
          <p className="text-sm text-foreground-muted">Loading...</p>
        )}
        {error && (
          <p className="text-sm text-red-500">{error}</p>
        )}
        {!loading && !error && activities.length === 0 && (
          <p className="text-sm text-foreground-muted">No activities yet. Upload a FIT file to get started.</p>
        )}
        {activities.map((activity) => (
          <ActivityFeedCard key={activity.id} activity={activity} />
        ))}
      </div>
    </div>
  );
}
