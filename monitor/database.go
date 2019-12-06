package monitor

import (
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/prometheus/tsdb"
	"github.com/prometheus/tsdb/labels"
	"github.com/prometheus/tsdb/wal"
)

const (
	// AllEntriesPattern is a pattern used to match all db entries.
	AllEntriesPattern = ".*"
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

// LoggingDatabase is used to store logging entries into a timeseries format.
type LoggingDatabase struct {
	db       *tsdb.DB
	appender tsdb.Appender
}

// NewLoggingDatabase is used to create a new logging database.
func NewLoggingDatabase() (*LoggingDatabase, error) {
	tempDir, err := ioutil.TempDir("", "accesslog")
	if err != nil {
		return nil, err
	}

	options := tsdb.Options{
		WALSegmentSize:         wal.DefaultSegmentSize,
		RetentionDuration:      60 * 60 * 1000, // 1 hour in milliseconds
		BlockRanges:            tsdb.ExponentialBlockRanges(int64(2*time.Hour)/1e6, 3, 5),
		NoLockfile:             false,
		AllowOverlappingBlocks: false,
		WALCompression:         false,
	}
	db, err := tsdb.Open(tempDir, nil, nil, &options)
	if err != nil {
		return nil, err
	}

	appender := db.Appender()

	return &LoggingDatabase{
		db:       db,
		appender: appender,
	}, nil
}

// AddEntry adds a new entry to database.
func (ld *LoggingDatabase) AddEntry(entry *LoggingEntry) error {
	labels := labels.FromMap(entry.Labels())
	ld.appender.Add(labels, entry.Date.Unix(), 1.0)
	return ld.appender.Commit()
}

// Query is used to get all the series between since-until interval that match the the given label matchers.
func (ld *LoggingDatabase) Query(since int64, until int64, matchers ...labels.Matcher) (tsdb.SeriesSet, error) {
	q, err := ld.db.Querier(since, until)
	if err != nil {
		return nil, nil
	}
	defer q.Close()

	return q.Select(matchers...)
}

// GetEntries can be used to collect all entries from a given label that match the pattern.
func (ld *LoggingDatabase) GetEntries(label string, pattern string, since int64, until int64) (EntryList, error) {
	// Collect the data
	matcher, err := labels.NewRegexpMatcher(label, pattern)
	if err != nil {
		return nil, err
	}

	query, err := ld.db.Querier(since, until)
	if err != nil {
		return nil, err
	}
	defer query.Close()

	series, err := query.Select(matcher)
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
func (ld *LoggingDatabase) TopEntries(label string, pattern string, since int64, until int64, limit int) (EntryList, error) {
	entries, err := ld.GetEntries(label, pattern, since, until)
	if err != nil {
		return nil, err
	}

	sort.Sort(sort.Reverse(entries))

	if limit > 0 && limit < entries.Len() {
		return entries[:limit], nil
	}

	return entries, nil
}

// Cleanup is used to drop all stored data.
func (ld *LoggingDatabase) Cleanup() error {
	err := os.RemoveAll(ld.db.Dir())
	if err != nil {
		return err
	}

	return ld.db.Close()
}
