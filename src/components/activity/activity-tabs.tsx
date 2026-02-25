"use client";

import { useState } from "react";
import { cn } from "@/lib/utils";
import { ACTIVITY_TABS, type ActivityTab } from "@/types/activity";

interface ActivityTabsProps {
  activityId: string;
}

export function ActivityTabs({ activityId }: ActivityTabsProps) {
  const [activeTab, setActiveTab] = useState<ActivityTab>("overview");

  return (
    <div className="border-b border-border">
      <nav className="-mb-px flex gap-4" aria-label="Activity tabs">
        {ACTIVITY_TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={cn(
              "border-b-2 px-1 py-3 text-sm font-medium transition-colors",
              activeTab === tab.id
                ? "border-primary text-primary"
                : "border-transparent text-foreground-muted hover:border-border-hover hover:text-foreground"
            )}
          >
            {tab.label}
          </button>
        ))}
      </nav>
    </div>
  );
}
