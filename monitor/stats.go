package monitor

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/prometheus/tsdb/labels"
)

const (
	limit = 3
)

// Entry represents a point from a timeseries set.
type Entry struct {
	Key   string
	Value float64
}

// EntryList represents a list of timeseries points.
type EntryList []Entry

func (p EntryList) Len() int           { return len(p) }
func (p EntryList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p EntryList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func mapToEntryList(m map[string]float64) EntryList {
	p := make(EntryList, len(m))

	i := 0
	for k, v := range m {
		p[i] = Entry{k, v}
		i++
	}
	return p
}

// GetEntries can be used to collect all entries from a given label that match the pattern.
func GetEntries(label string, pattern string, since int64, until int64, db *LoggingDatabase) (EntryList, error) {
	// Collect the data
	matcher, err := labels.NewRegexpMatcher(label, pattern)
	if err != nil {
		return nil, err
	}

	series, err := db.Query(since, until, matcher)
	if err != nil {
		return nil, err
	}

	// Group the data by label
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

	return mapToEntryList(m), nil
}

// TopEntries can be used to collect top entries from a given label that match the pattern.
func TopEntries(label string, since int64, until int64, limit int, db *LoggingDatabase) (EntryList, error) {
	entries, err := GetEntries(label, ".*", since, until, db)
	if err != nil {
		return nil, err
	}

	sort.Sort(sort.Reverse(entries))

	if limit > 0 && limit < entries.Len() {
		return entries[:limit], nil
	}

	return entries, nil
}

// StatsSummary is used to represent traffic statistics from a given interval.
type StatsSummary struct {
	Since           int64
	Until           int64
	Size            int64
	TopSections     EntryList
	TopUsers        EntryList
	RequestMethods  EntryList
	RequestStatuses EntryList
}

// NewStatsSummary is used to generate traffic statistics from a given interval.
func NewStatsSummary(since int64, until int64, db *LoggingDatabase) (*StatsSummary, error) {
	topSections, err := TopEntries(RequestURLSectionLabel, since, until, limit, db)
	if err != nil {
		return nil, err
	}

	topUsers, err := TopEntries(UserLabel, since, until, limit, db)
	if err != nil {
		return nil, err
	}

	requestMethods, err := TopEntries(RequestMethodLabel, since, until, 0, db)
	if err != nil {
		return nil, err
	}

	requestStatuses, err := TopEntries(StatusLabel, since, until, 0, db)
	if err != nil {
		return nil, err
	}

	return &StatsSummary{
		Since:           since,
		Until:           until,
		TopSections:     topSections,
		TopUsers:        topUsers,
		RequestMethods:  requestMethods,
		RequestStatuses: requestStatuses}, nil
}

func (s *StatsSummary) String() string {
	var stats bytes.Buffer

	var total, clientErrors, serverErrors, redirections, successful int
	for _, e := range s.RequestStatuses {
		total += int(e.Value)
		if "200" <= e.Key && e.Key < "300" {
			successful += int(e.Value)
		} else if "300" <= e.Key && e.Key < "400" {
			redirections += int(e.Value)
		} else if "400" <= e.Key && e.Key < "500" {
			clientErrors += int(e.Value)
		} else if e.Key >= "500" {
			serverErrors += int(e.Value)
		}
	}

	stats.WriteString("------------------------------------------------------------------------------------------------------------------------\n")
	stats.WriteString(fmt.Sprintf("Traffic stats between [%s, %s]:\n", time.Unix(s.Since, 0), time.Unix(s.Until, 0)))
	stats.WriteString(fmt.Sprintf("- From a total of %d requests, there were %d successful calls, %d redirections, %d client errors and %d server errors\n",
		total, successful, redirections, clientErrors, serverErrors))
	stats.WriteString(fmt.Sprintf("- Requests by method: %v\n", s.RequestMethods))
	stats.WriteString(fmt.Sprintf("- Top %d sections: %v\n", limit, s.TopSections))
	stats.WriteString(fmt.Sprintf("- First %d users: %v\n", limit, s.TopUsers))
	stats.WriteString("------------------------------------------------------------------------------------------------------------------------\n")

	return stats.String()
}
