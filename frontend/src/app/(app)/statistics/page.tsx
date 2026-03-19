import { PageHeader } from "@/components/layout";
import { StatsClient } from "@/components/statistics/stats-client";

export default function StatisticsPage() {
  return (
    <div className="space-y-6">
      <PageHeader title="Statistics" />
      <StatsClient />
    </div>
  );
}
