import { PageHeader } from "@/components/layout";

export default function DashboardPage() {
  return (
    <div>
      <PageHeader
        title="Dashboard"
        description="Overview of your cycling performance"
      />
      <div className="p-6">
        <div className="rounded-lg border border-border bg-background-subtle p-8 text-center">
          <p className="text-foreground-muted">
            Dashboard content coming soon...
          </p>
        </div>
      </div>
    </div>
  );
}
