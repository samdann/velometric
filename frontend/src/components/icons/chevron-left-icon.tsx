import { cn } from "@/lib/utils";

interface IconProps {
  className?: string;
}

export function ChevronLeftIcon({ className }: IconProps) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={cn("h-5 w-5", className)}
    >
      <polyline points="15 18 9 12 15 6" />
    </svg>
  );
}
