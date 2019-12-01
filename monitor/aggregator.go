package monitor

import (
	"github.com/prometheus/tsdb"
)

func GroupByLabel(series tsdb.SeriesSet, label string) (map[string]float64, error) {
	m := make(map[string]float64)

	for series.Next() {
		s := series.At()
		hits := 0.0

		it := s.Iterator()
		for it.Next() {
			_, v := it.At()
			hits += v
		}
		if err := it.Err(); err != nil {
			return nil, err
		}

		labelValue := s.Labels().Get(label)
		m[labelValue] += hits
	}

	return m, nil
}
