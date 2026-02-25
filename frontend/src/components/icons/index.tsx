import type { IconName } from "@/types/navigation";
import { DashboardIcon } from "./dashboard-icon";
import { ActivitiesIcon } from "./activities-icon";
import { UploadIcon } from "./upload-icon";
import { SettingsIcon } from "./settings-icon";
import { ChevronLeftIcon } from "./chevron-left-icon";

export { DashboardIcon } from "./dashboard-icon";
export { ActivitiesIcon } from "./activities-icon";
export { UploadIcon } from "./upload-icon";
export { SettingsIcon } from "./settings-icon";
export { ChevronLeftIcon } from "./chevron-left-icon";

interface IconProps {
  name: IconName;
  className?: string;
}

export function Icon({ name, className }: IconProps) {
  switch (name) {
    case "dashboard":
      return <DashboardIcon className={className} />;
    case "activities":
      return <ActivitiesIcon className={className} />;
    case "upload":
      return <UploadIcon className={className} />;
    case "settings":
      return <SettingsIcon className={className} />;
    case "chevron-left":
      return <ChevronLeftIcon className={className} />;
  }
}
