package monitor

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/prometheus/tsdb"
	"github.com/prometheus/tsdb/labels"
)

type Entry struct {
	Key   string
	Value float64
}

type EntryList []Entry

func (p EntryList) Len() int           { return len(p) }
func (p EntryList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p EntryList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func mapToPairList(m map[string]float64) EntryList {
	p := make(EntryList, len(m))

	i := 0
	for k, v := range m {
		p[i] = Entry{k, v}
		i++
	}
	return p
}

func EntriesByLabel(series tsdb.SeriesSet, label string) (map[string]float64, error) {
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

func getEntries(label string, pattern string, since int64, until int64, db *LoggingDatabase) (map[string]float64, error) {
	matcher, err := labels.NewRegexpMatcher(label, pattern)
	if err != nil {
		return nil, err
	}

	series, err := db.Query(since, until, matcher)
	if err != nil {
		return nil, err
	}

	entries, err := EntriesByLabel(series, label)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func TopHitsByLabel(label string, since int64, until int64, limit int, db *LoggingDatabase) (EntryList, error) {
	entries, err := getEntries(label, ".*", since, until, db)
	if err != nil {
		return nil, err
	}

	hits := mapToPairList(entries)
	sort.Sort(sort.Reverse(hits))

	if limit > 0 && limit < hits.Len() {
		return hits[:limit], nil
	}

	return hits, nil
}

func generateTotalTraffic(since int64, until int64, db *LoggingDatabase) (string, error) {
	entries, err := getEntries(StatusLabel, ".*", since, until, db)
	if err != nil {
		return "", err
	}

	var total, clientErrors, serverErrors, redirections, successful int

	for k, v := range entries {
		total += int(v)
		if "200" <= k && k < "300" {
			successful += int(v)
		} else if "300" <= k && k < "400" {
			redirections += int(v)
		} else if "400" <= k && k < "500" {
			clientErrors += int(v)
		} else if k >= "500" {
			serverErrors += int(v)
		}
	}

	return fmt.Sprintf("Total requests: %d, Successful: %d, Redirections: %d, Client errors: %d, Server errors: %d",
		total, successful, redirections, clientErrors, serverErrors), nil
}

func generateSectionStats(since int64, until int64, db *LoggingDatabase) (string, error) {
	hits, err := TopHitsByLabel(RequestURLSectionLabel, since, until, 3, db)
	if err != nil {
		return "", err
	}

	if hits.Len() == 0 {
		return "N/A", nil
	}

	var sectionStats bytes.Buffer

	for i := 0; i < hits.Len(); i++ {
		sectionStats.WriteString(fmt.Sprintf("%s %d hits,", hits[i].Key, int64(hits[i].Value)))
	}

	return sectionStats.String(), nil
}

func generateUserStats(since int64, until int64, db *LoggingDatabase) (string, error) {
	hits, err := TopHitsByLabel(UserLabel, since, until, 1, db)
	if err != nil {
		return "", err
	}

	if hits.Len() == 0 {
		return "N/A", nil
	}

	return fmt.Sprintf("%s - %d requests", hits[0].Key, int64(hits[0].Value)), nil
}

func generateMethodStats(since int64, until int64, db *LoggingDatabase) (string, error) {
	hits, err := TopHitsByLabel(RequestMethodLabel, since, until, 0, db)
	if err != nil {
		return "", err
	}

	if hits.Len() == 0 {
		return "N/A", nil
	}

	var methodStats bytes.Buffer

	for i := 0; i < hits.Len(); i++ {
		methodStats.WriteString(fmt.Sprintf("%d %s, ", int64(hits[i].Value), hits[i].Key))
	}

	return methodStats.String(), nil
}

func printLoggingStats(since int64, until int64, db *LoggingDatabase) error {
	trafficStats, err := generateTotalTraffic(since, until, db)
	if err != nil {
		return nil
	}

	sectionStats, err := generateSectionStats(since, until, db)
	if err != nil {
		return err
	}

	userStats, err := generateUserStats(since, until, db)
	if err != nil {
		return err
	}

	methodStats, err := generateMethodStats(since, until, db)
	if err != nil {
		return err
	}

	fmt.Printf("Summary stats from last %d seconds:\n", (until - since))
	fmt.Printf("- Operations details [%s\n", methodStats)
	fmt.Printf("- Traffic details [%s]\n", trafficStats)
	fmt.Printf("- Top 3 sections  [%s]\n", sectionStats)
	fmt.Printf("- User with biggest number of requests [%s]\n", userStats)

	return nil
}
