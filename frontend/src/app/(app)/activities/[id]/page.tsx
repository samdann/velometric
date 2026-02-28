"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { PageHeader } from "@/components/layout";
import { ActivityTabs } from "@/components/activity/activity-tabs";
import { PowerTab } from "@/components/activity/power-tab";
import { HeartRateTab } from "@/components/activity/heart-rate-tab";
import { MapTab } from "@/components/activity/map-tab";
import { SegmentsTab } from "@/components/activity/segments-tab";
import { LapsTab } from "@/components/activity/laps-tab";
import { DataTab } from "@/components/activity/data-tab";
import { ElevationChart } from "@/components/activity/elevation-chart";
import { SpeedChart } from "@/components/activity/speed-chart";
import { HRCadenceChart } from "@/components/activity/hr-cadence-chart";
import { api, Activity } from "@/lib/api";
import { ActivityTab } from "@/types/activity";

function formatDuration(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  if (h > 0) {
    return `${h}:${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
  }
  return `${m}:${s.toString().padStart(2, "0")}`;
}

function formatDistance(meters: number): string {
  const km = meters / 1000;
  return `${km.toFixed(2)} km`;
}

function formatSpeed(mps: number): string {
  const kph = mps * 3.6;
  return `${kph.toFixed(1)} km/h`;
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString("en-US", {
    weekday: "long",
    month: "long",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

interface StatCardProps {
  label: string;
  value: string | number;
  unit?: string;
  color?: string;
}

function StatCard({ label, value, unit, color }: StatCardProps) {
  return (
    <div className="rounded-lg border border-border bg-background-subtle p-4">
      <p className="text-xs text-foreground-muted">{label}</p>
      <p className={`mt-1 font-mono text-2xl ${color || "text-foreground"}`}>
        {value}
        {unit && <span className="ml-1 text-sm text-foreground-muted">{unit}</span>}
      </p>
    </div>
  );
}

export default function ActivityDetailPage() {
  const params = useParams();
  const id = params.id as string;

  const [activity, setActivity] = useState<Activity | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<ActivityTab>("overview");

  useEffect(() => {
    async function fetchActivity() {
      try {
        const data = await api.getActivity(id);
        setActivity(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load activity");
      } finally {
        setLoading(false);
      }
    }
    fetchActivity();
  }, [id]);

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <p className="text-foreground-muted">Loading...</p>
      </div>
    );
  }

  if (error || !activity) {
    return (
      <div className="p-6">
        <div className="rounded-lg bg-heart-rate/10 p-4 text-center">
          <p className="text-sm text-heart-rate">{error || "Activity not found"}</p>
        </div>
      </div>
    );
  }

  return (
    <div>
      <PageHeader
        title={activity.name}
        description={formatDate(activity.startTime)}
      />
      <div className="p-6">
        <ActivityTabs activeTab={activeTab} onTabChange={setActiveTab} />

        {/* Tab Content */}
        {activeTab === "overview" && (
        <div>
        <div className="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
          <StatCard
            label="Duration"
            value={formatDuration(activity.duration)}
          />
          <StatCard
            label="Distance"
            value={formatDistance(activity.distance)}
          />
          <StatCard
            label="Elevation"
            value={activity.elevationGain?.toFixed(0) || "0"}
            unit="m"
            color="text-elevation"
          />
          {activity.avgPower && (
            <StatCard
              label="Avg Power"
              value={activity.avgPower}
              unit="w"
              color="text-power"
            />
          )}
          {activity.normalizedPower && (
            <StatCard
              label="Normalized Power"
              value={activity.normalizedPower}
              unit="w"
              color="text-power"
            />
          )}
          {activity.tss && (
            <StatCard
              label="TSS"
              value={activity.tss.toFixed(0)}
            />
          )}
          {activity.avgHeartRate && (
            <StatCard
              label="Avg HR"
              value={activity.avgHeartRate}
              unit="bpm"
              color="text-heart-rate"
            />
          )}
          {activity.maxHeartRate && (
            <StatCard
              label="Max HR"
              value={activity.maxHeartRate}
              unit="bpm"
              color="text-heart-rate"
            />
          )}
          {activity.avgCadence && (
            <StatCard
              label="Avg Cadence"
              value={activity.avgCadence}
              unit="rpm"
              color="text-cadence"
            />
          )}
          {activity.avgSpeed && (
            <StatCard
              label="Avg Speed"
              value={formatSpeed(activity.avgSpeed)}
              color="text-speed"
            />
          )}
          {activity.intensityFactor && (
            <StatCard
              label="IF"
              value={activity.intensityFactor.toFixed(2)}
            />
          )}
          {activity.variabilityIndex && (
            <StatCard
              label="VI"
              value={activity.variabilityIndex.toFixed(2)}
            />
          )}
        </div>
          <ElevationChart activityId={id} />
          <SpeedChart activityId={id} />
          <HRCadenceChart activityId={id} />
        </div>
        )}

        {activeTab === "power" && (
          <PowerTab activityId={id} activity={activity} />
        )}

        {activeTab === "heart-rate" && (
          <HeartRateTab activity={activity} />
        )}

        {activeTab === "map" && (
          <MapTab activityId={id} />
        )}

        {activeTab === "segments" && (
          <SegmentsTab />
        )}

        {activeTab === "laps" && (
          <LapsTab activityId={id} />
        )}

        {activeTab === "data" && (
          <DataTab activityId={id} />
        )}
      </div>
    </div>
  );
}
