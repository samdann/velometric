"use client";

interface Props {
  hint: string;
}

export function HelpIcon({ hint }: Props) {
  return (
    <div className="relative group">
      <button className="leading-none text-foreground-muted hover:text-foreground transition-colors">
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <circle cx="12" cy="12" r="10" />
          <line x1="12" y1="8" x2="12" y2="12" />
          <line x1="12" y1="16" x2="12.01" y2="16" />
        </svg>
      </button>
      <div className="pointer-events-none absolute z-20 left-0 top-5 w-72 rounded-lg border border-border bg-background-subtle p-3 text-xs leading-relaxed text-foreground-muted shadow-xl opacity-0 group-hover:opacity-100 transition-opacity duration-150">
        {hint}
      </div>
    </div>
  );
}
