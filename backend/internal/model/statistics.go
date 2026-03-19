package model

// AnnualPowerCurvePoint holds the median best power at a given duration across all activities in a year.
type AnnualPowerCurvePoint struct {
	DurationSeconds int `json:"durationSeconds"`
	MedianPower     int `json:"medianPower"`
}

// AnnualZoneDistributionPoint holds the median percentage of time spent in a power zone across rides.
type AnnualZoneDistributionPoint struct {
	ZoneNumber       int     `json:"zoneNumber"`
	Name             string  `json:"name"`
	Color            string  `json:"color"`
	MinWatts         int     `json:"minWatts"`
	MaxWatts         *int    `json:"maxWatts"`
	MedianPercentage float64 `json:"medianPercentage"`
}

// AnnualPowerStats is the combined response for the statistics power endpoint.
type AnnualPowerStats struct {
	PowerCurve       []AnnualPowerCurvePoint       `json:"powerCurve"`
	ZoneDistribution []AnnualZoneDistributionPoint `json:"zoneDistribution"`
}
