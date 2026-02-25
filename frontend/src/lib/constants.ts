import type { NavItem } from "@/types/navigation";

export const NAV_ITEMS: NavItem[] = [
  {
    label: "Dashboard",
    href: "/",
    icon: "dashboard",
  },
  {
    label: "Activities",
    href: "/activities",
    icon: "activities",
  },
  {
    label: "Upload",
    href: "/upload",
    icon: "upload",
  },
  {
    label: "Settings",
    href: "/settings",
    icon: "settings",
  },
];

export const SIDEBAR_WIDTH = {
  expanded: 240,
  collapsed: 64,
} as const;

export const STORAGE_KEYS = {
  sidebarCollapsed: "velometric-sidebar-collapsed",
} as const;
