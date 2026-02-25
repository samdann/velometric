"use client";

import { useEffect, useState } from "react";

interface Record {
  lat?: number;
  lon?: number;
}

interface MapTabProps {
  activityId: string;
}

export function MapTab({ activityId }: MapTabProps) {
  const [hasGPS, setHasGPS] = useState<boolean | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function checkGPS() {
      try {
        const response = await fetch(
          `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8081"}/api/activities/${activityId}/records`
        );
        if (!response.ok) throw new Error("Failed to fetch records");
        const records: Record[] = await response.json();

        // Check if any records have GPS coordinates
        const gpsRecords = records.filter((r) => r.lat && r.lon);
        setHasGPS(gpsRecords.length > 0);
      } catch {
        setHasGPS(false);
      } finally {
        setLoading(false);
      }
    }
    checkGPS();
  }, [activityId]);

  if (loading) {
    return (
      <div className="mt-6 flex h-32 items-center justify-center">
        <p className="text-foreground-muted">Loading map data...</p>
      </div>
    );
  }

  if (!hasGPS) {
    return (
      <div className="mt-6 rounded-lg border border-border bg-background-subtle p-6 text-center">
        <p className="text-foreground-muted">No GPS data available for this activity</p>
      </div>
    );
  }

  return (
    <div className="mt-6">
      <div className="flex h-96 items-center justify-center rounded-lg border border-border bg-background-subtle">
        <div className="text-center">
          <p className="text-foreground-muted">Map visualization coming soon</p>
          <p className="mt-2 text-sm text-foreground-muted">
            GPS data is available for this activity
          </p>
        </div>
      </div>
    </div>
  );
}
