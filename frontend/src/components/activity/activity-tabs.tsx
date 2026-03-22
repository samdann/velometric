"use client";

import { cn } from "@/lib/utils";
import { ALL_ACTIVITY_TABS, type ActivityTab } from "@/types/activity";

interface ActivityTabsProps {
  activeTab: ActivityTab;
  onTabChange: (tab: ActivityTab) => void;
  visibleTabs?: ActivityTab[];
}

export function ActivityTabs({ activeTab, onTabChange, visibleTabs }: ActivityTabsProps) {
  const tabs = visibleTabs
    ? ALL_ACTIVITY_TABS.filter((t) => visibleTabs.includes(t.id))
    : ALL_ACTIVITY_TABS;

  return (
    <div className="border-b border-border">
      <nav className="-mb-px flex gap-4" aria-label="Activity tabs">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => onTabChange(tab.id)}
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
