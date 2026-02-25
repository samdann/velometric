import { PageHeader } from "@/components/layout";

export default function SettingsPage() {
  return (
    <div>
      <PageHeader
        title="Settings"
        description="Configure your preferences"
      />
      <div className="p-6">
        <div className="rounded-lg border border-border bg-background-subtle p-8 text-center">
          <p className="text-foreground-muted">
            Settings content coming soon...
          </p>
        </div>
      </div>
    </div>
  );
}
