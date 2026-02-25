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

export const ACTIVITY_TABS: { id: ActivityTab; label: string }[] = [
  { id: "overview", label: "Overview" },
  { id: "power", label: "Power" },
  { id: "heart-rate", label: "Heart Rate" },
  { id: "map", label: "Map" },
  { id: "segments", label: "Segments" },
  { id: "laps", label: "Laps" },
  { id: "data", label: "Data" },
];
