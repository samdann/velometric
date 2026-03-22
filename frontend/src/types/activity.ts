export interface Activity {
  id: string;
  name: string;
  date: string;
  duration: number; // seconds
  distance: number; // meters
  elevationGain: number; // meters
  averagePower?: number; // watts
  normalizedPower?: number; // watts
  averageHeartRate?: number; // bpm
  averageCadence?: number; // rpm
  averageSpeed: number; // m/s
}

export type ActivityTab =
  | "overview"
  | "power"
  | "heart-rate"
  | "map"
  | "segments"
  | "laps"
  | "data";

export type ChartKey = "elevation" | "speed" | "hr-cadence";

export interface SportConfig {
  tabs: ActivityTab[];
  charts: ChartKey[];
}

const ALL_CHARTS: ChartKey[] = ["elevation", "speed", "hr-cadence"];

export const SPORT_CONFIG: Record<string, SportConfig> = {
  Ride:    { tabs: ["overview", "power", "heart-rate", "map", "laps", "data"], charts: ALL_CHARTS },
  Run:     { tabs: ["overview", "heart-rate", "map", "laps", "data"],          charts: ALL_CHARTS },
  Swim:    { tabs: ["overview", "laps", "data"],                               charts: ["speed", "hr-cadence"] },
  Hike:    { tabs: ["overview", "heart-rate", "map", "laps", "data"],          charts: ALL_CHARTS },
  Walk:    { tabs: ["overview", "heart-rate", "map", "laps", "data"],          charts: ALL_CHARTS },
  Workout: { tabs: ["overview", "heart-rate", "laps", "data"],                 charts: ["hr-cadence"] },
  Other:   { tabs: ["overview", "map", "laps", "data"],                        charts: ALL_CHARTS },
};

export const DEFAULT_SPORT_CONFIG: SportConfig = SPORT_CONFIG.Other;

/** All tab definitions — SportConfig.tabs controls which subset is shown. */
export const ALL_ACTIVITY_TABS: { id: ActivityTab; label: string }[] = [
  { id: "overview",   label: "Overview" },
  { id: "power",      label: "Power" },
  { id: "heart-rate", label: "Heart Rate" },
  { id: "map",        label: "Map" },
  { id: "segments",   label: "Segments" },
  { id: "laps",       label: "Laps" },
  { id: "data",       label: "Data" },
];

// Keep backward-compatible export used by ActivityTabs before sport config was introduced.
export const ACTIVITY_TABS = ALL_ACTIVITY_TABS;
