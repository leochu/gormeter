package stats

type Summary struct {
	Min                         float64   `json:"min"`
	Max                         float64   `json:"max"`
	Sum                         float64   `json:"sum"`
	Mean                        float64   `json:"mean"`
	Median                      float64   `json:"median"`
	Mode                        []float64 `json:"mode"`
	PopulationVariance          float64   `json:"population_variance"`
	SampleVariance              float64   `json:"sample_variance"`
	StandardDeviationPopulation float64   `json:standard_deviation_population`
	StandardDeviationSample     float64   `json:standard_deviation_sample`
	PercentileOf99              float64   `json:percentile_of_99`
	PercentileNearestRankOf99   float64   `json:percentile_nearest_rank_of_99`
	PercentileOf95              float64   `json:percentile_of_95`
	PercentileNearestRankOf95   float64   `json:percentile_nearest_rank_of_95`
	PercentileOf90              float64   `json:percentile_of_90`
	PercentileNearestRankOf90   float64   `json:percentile_nearest_rank_of_90`
	PercentileOf85              float64   `json:percentile_of_85`
	PercentileNearestRankOf85   float64   `json:percentile_nearest_rank_of_85`
}
