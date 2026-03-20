export const STATISTICS_HINTS = {
  powerCurveAvg:
    "Median power curve across all rides in the year. For each duration (5s → 1h), the value shown is the median of your best efforts — half your rides produced higher, half lower. A reliable picture of your typical power ceiling.",
  powerCurveBest:
    "Best power curve for the year. For each duration independently, shows the single highest power you produced across all rides — your absolute peak at each effort length.",
  zoneDistributionAvg:
    "Median time-in-zone across all rides in the year. Each zone percentage is the median across all rides — a picture of how a typical ride was distributed between zones.",
  zoneDistributionBest:
    "Zone distribution of your highest-quality ride of the year, identified by the ride with the highest Normalized Power (NP). Reflects how you spent your time on your most demanding effort.",
} as const;

export type StatisticsHintKey = keyof typeof STATISTICS_HINTS;
