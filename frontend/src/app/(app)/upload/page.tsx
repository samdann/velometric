"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { PageHeader } from "@/components/layout";
import { FileUpload } from "@/components/ui/file-upload";
import { api } from "@/lib/api";

type UploadState = "idle" | "uploading" | "success" | "error";

export default function UploadPage() {
  const router = useRouter();
  const [state, setState] = useState<UploadState>("idle");
  const [error, setError] = useState<string | null>(null);

  const handleFileSelect = async (file: File) => {
    setState("uploading");
    setError(null);

    try {
      const result = await api.uploadActivity(file);
      setState("success");

      // Redirect to the activity detail page after a brief delay
      setTimeout(() => {
        router.push(`/activities/${result.id}`);
      }, 1000);
    } catch (err) {
      setState("error");
      setError(err instanceof Error ? err.message : "Upload failed");
    }
  };

  return (
    <div>
      <PageHeader
        title="Upload"
        description="Import FIT files from your rides"
      />
      <div className="p-6">
        <div className="mx-auto max-w-xl">
          <FileUpload
            onFileSelect={handleFileSelect}
            disabled={state === "uploading"}
            className="min-h-[240px]"
          />

          {state === "uploading" && (
            <div className="mt-4 text-center">
              <div className="inline-flex items-center gap-2 text-sm text-foreground-muted">
                <svg
                  className="h-4 w-4 animate-spin"
                  viewBox="0 0 24 24"
                  fill="none"
                >
                  <circle
                    className="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    strokeWidth="4"
                  />
                  <path
                    className="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                  />
                </svg>
                Processing your ride...
              </div>
            </div>
          )}

          {state === "success" && (
            <div className="mt-4 rounded-lg bg-elevation/10 p-4 text-center">
              <p className="text-sm font-medium text-elevation">
                Activity uploaded successfully!
              </p>
              <p className="mt-1 text-xs text-foreground-muted">
                Redirecting to activity details...
              </p>
            </div>
          )}

          {state === "error" && error && (
            <div className="mt-4 rounded-lg bg-heart-rate/10 p-4 text-center">
              <p className="text-sm font-medium text-heart-rate">
                Upload failed
              </p>
              <p className="mt-1 text-xs text-foreground-muted">{error}</p>
            </div>
          )}

          <div className="mt-8">
            <h3 className="text-sm font-medium text-foreground">
              Supported formats
            </h3>
            <ul className="mt-2 space-y-1 text-sm text-foreground-muted">
              <li>• FIT files from Garmin, Wahoo, and other devices</li>
              <li>• Files up to 32MB in size</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}
