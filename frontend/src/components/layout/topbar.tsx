"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { UploadIcon, SettingsIcon } from "@/components/icons";

function NavLink({
  href,
  exact,
  children,
}: {
  href: string;
  exact?: boolean;
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const isActive = exact ? pathname === href : pathname.startsWith(href);

  return (
    <Link
      href={href}
      className={cn(
        "px-3 py-1.5 text-sm font-medium rounded-md transition-colors",
        isActive
          ? "text-primary"
          : "text-foreground-muted hover:text-foreground"
      )}
    >
      {children}
    </Link>
  );
}

function IconLink({
  href,
  label,
  children,
}: {
  href: string;
  label: string;
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const isActive = pathname.startsWith(href);

  return (
    <div className="relative group">
      <Link
        href={href}
        className={cn(
          "flex items-center justify-center w-9 h-9 rounded-md transition-colors",
          isActive
            ? "text-primary"
            : "text-foreground-muted hover:text-foreground hover:bg-background-subtle"
        )}
      >
        {children}
      </Link>
      <div className="absolute right-0 top-full mt-1.5 px-2 py-1 text-xs text-foreground bg-surface border border-border rounded-md opacity-0 group-hover:opacity-100 pointer-events-none whitespace-nowrap transition-opacity duration-150 shadow-sm">
        {label}
      </div>
    </div>
  );
}

export function Topbar() {
  return (
    <header className="sticky top-0 z-40 flex h-14 items-center border-b border-border bg-background-subtle px-6 gap-8">
      <div className="flex items-center gap-2">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
          <span className="font-mono text-sm font-bold text-background">V</span>
        </div>
        <span className="text-lg font-semibold text-foreground">Velometric</span>
      </div>

      <nav className="flex items-center gap-1">
        <NavLink href="/" exact>Dashboard</NavLink>
        <NavLink href="/activities">Activities</NavLink>
      </nav>

      <div className="ml-auto flex items-center gap-1">
        <IconLink href="/upload" label="Upload">
          <UploadIcon />
        </IconLink>
        <IconLink href="/settings" label="Settings">
          <SettingsIcon />
        </IconLink>
      </div>
    </header>
  );
}
