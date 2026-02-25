"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { Icon } from "@/components/icons";
import { useSidebar } from "@/hooks/use-sidebar";
import type { NavItem as NavItemType } from "@/types/navigation";

interface NavItemProps {
  item: NavItemType;
}

export function NavItem({ item }: NavItemProps) {
  const pathname = usePathname();
  const { isCollapsed } = useSidebar();

  const isActive =
    item.href === "/"
      ? pathname === "/"
      : pathname.startsWith(item.href);

  return (
    <Link
      href={item.href}
      className={cn(
        "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
        "hover:bg-background-subtle hover:text-foreground",
        isActive
          ? "bg-background-subtle text-primary"
          : "text-foreground-muted",
        isCollapsed && "justify-center px-2"
      )}
    >
      <Icon name={item.icon} className={cn(isActive && "text-primary")} />
      {!isCollapsed && <span>{item.label}</span>}
    </Link>
  );
}
