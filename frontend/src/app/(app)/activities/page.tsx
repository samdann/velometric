import { PageHeader } from "@/components/layout";

export default function ActivitiesPage() {
  return (
    <div>
      <PageHeader
        title="Activities"
        description="View and analyze your rides"
      />
      <div className="p-6">
        <div className="rounded-lg border border-border bg-background-subtle p-8 text-center">
          <p className="text-foreground-muted">
            Activity list coming soon...
          </p>
        </div>
      </div>
    </div>
  );
}
