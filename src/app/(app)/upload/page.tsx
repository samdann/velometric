import { PageHeader } from "@/components/layout";

export default function UploadPage() {
  return (
    <div>
      <PageHeader
        title="Upload"
        description="Import FIT files from your rides"
      />
      <div className="p-6">
        <div className="rounded-lg border border-border bg-background-subtle p-8 text-center">
          <p className="text-foreground-muted">
            FIT file upload coming soon...
          </p>
        </div>
      </div>
    </div>
  );
}
