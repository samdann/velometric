"use client";

import { SidebarProvider } from "@/hooks/use-sidebar";
import { Sidebar } from "./sidebar";
import { MainContent } from "./main-content";

interface AppLayoutProps {
  children: React.ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
  return (
    <SidebarProvider>
      <div className="flex min-h-screen">
        <Sidebar />
        <MainContent>{children}</MainContent>
      </div>
    </SidebarProvider>
  );
}
