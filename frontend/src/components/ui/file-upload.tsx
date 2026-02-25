"use client";

import { useCallback, useState } from "react";
import { cn } from "@/lib/utils";
import { UploadIcon } from "@/components/icons";

interface FileUploadProps {
  onFileSelect: (file: File) => void;
  accept?: string;
  disabled?: boolean;
  className?: string;
}

export function FileUpload({
  onFileSelect,
  accept = ".fit",
  disabled = false,
  className,
}: FileUploadProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [fileName, setFileName] = useState<string | null>(null);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (!disabled) {
      setIsDragging(true);
    }
  }, [disabled]);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      setIsDragging(false);

      if (disabled) return;

      const file = e.dataTransfer.files[0];
      if (file && file.name.endsWith(".fit")) {
        setFileName(file.name);
        onFileSelect(file);
      }
    },
    [disabled, onFileSelect]
  );

  const handleFileInput = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (file) {
        setFileName(file.name);
        onFileSelect(file);
      }
    },
    [onFileSelect]
  );

  return (
    <div
      className={cn(
        "relative flex flex-col items-center justify-center rounded-lg border-2 border-dashed p-8 transition-colors",
        isDragging
          ? "border-primary bg-primary/5"
          : "border-border hover:border-border-hover",
        disabled && "cursor-not-allowed opacity-50",
        className
      )}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <input
        type="file"
        accept={accept}
        onChange={handleFileInput}
        disabled={disabled}
        className="absolute inset-0 cursor-pointer opacity-0"
      />

      <div className="flex flex-col items-center gap-4 text-center">
        <div
          className={cn(
            "flex h-16 w-16 items-center justify-center rounded-full",
            isDragging ? "bg-primary/10" : "bg-background-subtle"
          )}
        >
          <UploadIcon
            className={cn(
              "h-8 w-8",
              isDragging ? "text-primary" : "text-foreground-muted"
            )}
          />
        </div>

        {fileName ? (
          <div>
            <p className="text-sm font-medium text-foreground">{fileName}</p>
            <p className="mt-1 text-xs text-foreground-muted">
              Click or drop another file to replace
            </p>
          </div>
        ) : (
          <div>
            <p className="text-sm font-medium text-foreground">
              Drop your FIT file here
            </p>
            <p className="mt-1 text-xs text-foreground-muted">
              or click to browse
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
