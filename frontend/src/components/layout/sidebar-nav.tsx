"use client";

import { NAV_ITEMS } from "@/lib/constants";
import { NavItem } from "./nav-item";

export function SidebarNav() {
  return (
    <nav className="flex flex-col gap-1 px-2">
      {NAV_ITEMS.map((item) => (
        <NavItem key={item.href} item={item} />
      ))}
    </nav>
  );
}
