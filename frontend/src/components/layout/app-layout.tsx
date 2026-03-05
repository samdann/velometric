"use client";

import { Topbar } from "./topbar";
import { MainContent } from "./main-content";

interface AppLayoutProps {
  children: React.ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
  return (
    <div className="flex min-h-screen flex-col w-full max-w-[1100px] mx-auto">
      <Topbar />
      <MainContent>{children}</MainContent>
    </div>
  );
}
