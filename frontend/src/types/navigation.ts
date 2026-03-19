export type IconName = "dashboard" | "activities" | "upload" | "settings" | "chevron-left" | "statistics";

export interface NavItem {
  label: string;
  href: string;
  icon: IconName;
}
