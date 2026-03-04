"use client";

import { useEffect, useState } from "react";
import { Activity, HRZoneDistributionPoint, api } from "@/lib/api";
import { HRZonesChart } from "./hr-zones-chart";

interface HeartRateTabProps {
  activity: Activity;
}

export function HeartRateTab({ activity }: HeartRateTabProps) {
  const hasHRData = activity.avgHeartRate || activity.maxHeartRate;
  const [distribution, setDistribution] = useState<HRZoneDistributionPoint[] | null>(null);

  useEffect(() => {
    if (!hasHRData) return;
    api.getHRZoneDistribution(activity.id).then(setDistribution).catch(() => setDistribution([]));
  }, [activity.id, hasHRData]);

  if (!hasHRData) {
    return (
      <div className="mt-6 rounded-lg border border-border bg-background-subtle p-6 text-center">
        <p className="text-foreground-muted">No heart rate data available for this activity</p>
      </div>
    );
  }

  return (
    <div className="mt-6 space-y-6">
      {/* HR Summary */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        {activity.avgHeartRate && (
          <div className="rounded-lg border border-border bg-background-subtle p-4">
            <p className="text-xs text-foreground-muted">Avg Heart Rate</p>
            <p className="mt-1 font-mono text-2xl text-heart-rate">
              {activity.avgHeartRate}<span className="ml-1 text-sm text-foreground-muted">bpm</span>
            </p>
          </div>
        )}
        {activity.maxHeartRate && (
          <div className="rounded-lg border border-border bg-background-subtle p-4">
            <p className="text-xs text-foreground-muted">Max Heart Rate</p>
            <p className="mt-1 font-mono text-2xl text-heart-rate">
              {activity.maxHeartRate}<span className="ml-1 text-sm text-foreground-muted">bpm</span>
            </p>
          </div>
        )}
      </div>

      {/* HR Zones Distribution */}
      {distribution === null ? (
        <div className="rounded-lg border border-border bg-background-subtle p-6">
          <div className="h-32 animate-pulse rounded bg-background" />
        </div>
      ) : (
        <HRZonesChart distribution={distribution} />
      )}
    </div>
  );
}
