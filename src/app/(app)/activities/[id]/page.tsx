import { PageHeader } from "@/components/layout";
import { ActivityTabs } from "@/components/activity/activity-tabs";

interface ActivityDetailPageProps {
  params: Promise<{ id: string }>;
}

export default async function ActivityDetailPage({
  params,
}: ActivityDetailPageProps) {
  const { id } = await params;

  return (
    <div>
      <PageHeader
        title={`Activity ${id}`}
        description="Detailed activity analysis"
      />
      <div className="p-6">
        <ActivityTabs activityId={id} />
        <div className="mt-6 rounded-lg border border-border bg-background-subtle p-8 text-center">
          <p className="text-foreground-muted">
            Activity detail content coming soon...
          </p>
        </div>
      </div>
    </div>
  );
}
