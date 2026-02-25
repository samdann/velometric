"use client";

import { useSidebar } from "@/hooks/use-sidebar";
import { SIDEBAR_WIDTH } from "@/lib/constants";

interface MainContentProps {
  children: React.ReactNode;
}

export function MainContent({ children }: MainContentProps) {
  const { isCollapsed } = useSidebar();

  return (
    <main
      className="min-h-screen bg-background transition-[margin-left] duration-200"
      style={{
        marginLeft: isCollapsed
          ? SIDEBAR_WIDTH.collapsed
          : SIDEBAR_WIDTH.expanded,
      }}
    >
      {children}
    </main>
  );
}
