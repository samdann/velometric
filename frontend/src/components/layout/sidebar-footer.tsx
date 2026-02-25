"use client";

import { useSidebar } from "@/hooks/use-sidebar";
import { ChevronLeftIcon } from "@/components/icons";
import { cn } from "@/lib/utils";

export function SidebarFooter() {
  const { isCollapsed, toggle } = useSidebar();

  return (
    <div className="border-t border-border p-2">
      <button
        onClick={toggle}
        className={cn(
          "flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-foreground-muted transition-colors",
          "hover:bg-background-subtle hover:text-foreground",
          isCollapsed && "justify-center px-2"
        )}
      >
        <ChevronLeftIcon
          className={cn(
            "transition-transform duration-200",
            isCollapsed && "rotate-180"
          )}
        />
        {!isCollapsed && <span>Collapse</span>}
      </button>
    </div>
  );
}
