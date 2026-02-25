"use client";

import { useSidebar } from "@/hooks/use-sidebar";
import { cn } from "@/lib/utils";
import { SIDEBAR_WIDTH } from "@/lib/constants";
import { SidebarHeader } from "./sidebar-header";
import { SidebarNav } from "./sidebar-nav";
import { SidebarFooter } from "./sidebar-footer";

export function Sidebar() {
  const { isCollapsed } = useSidebar();

  return (
    <aside
      className={cn(
        "fixed left-0 top-0 z-40 flex h-screen flex-col bg-background-subtle transition-[width] duration-200"
      )}
      style={{
        width: isCollapsed ? SIDEBAR_WIDTH.collapsed : SIDEBAR_WIDTH.expanded,
      }}
    >
      <SidebarHeader />
      <div className="flex-1 overflow-y-auto py-4">
        <SidebarNav />
      </div>
      <SidebarFooter />
    </aside>
  );
}
