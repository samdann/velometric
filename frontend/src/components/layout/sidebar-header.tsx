"use client";

import { useSidebar } from "@/hooks/use-sidebar";
import { cn } from "@/lib/utils";

export function SidebarHeader() {
  const { isCollapsed } = useSidebar();

  return (
    <div
      className={cn(
        "flex h-14 items-center border-b border-border px-4",
        isCollapsed && "justify-center px-2"
      )}
    >
      <div className="flex items-center gap-2">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
          <span className="font-mono text-sm font-bold text-background">V</span>
        </div>
        {!isCollapsed && (
          <span className="text-lg font-semibold text-foreground">
            Velometric
          </span>
        )}
      </div>
    </div>
  );
}
