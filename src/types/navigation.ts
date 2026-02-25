export type IconName = "dashboard" | "activities" | "upload" | "settings" | "chevron-left";

export interface NavItem {
  label: string;
  href: string;
  icon: IconName;
}
