package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/prometheus/tsdb"
	"github.com/prometheus/tsdb/labels"
	"github.com/prometheus/tsdb/wal"
)

// LoggingDatabase is used to blabla //TODO
type LoggingDatabase struct {
	db       *tsdb.DB
	appender tsdb.Appender
}

// NewLoggingDatabase is used to create a new logging database
func NewLoggingDatabase() (*LoggingDatabase, error) {
	tempDir, err := ioutil.TempDir("", "accesslog")
	if err != nil {
		return nil, err
	}

	db, err := tsdb.Open(tempDir, nil, nil, &tsdb.Options{
		WALSegmentSize:         wal.DefaultSegmentSize,
		RetentionDuration:      60 * 60 * 1000, // 1 hour in milliseconds
		BlockRanges:            tsdb.ExponentialBlockRanges(int64(2*time.Hour)/1e6, 3, 5),
		NoLockfile:             false,
		AllowOverlappingBlocks: false,
		WALCompression:         false,
	})

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
func (m *LoggingDatabase) AddEntry(entry *LoggingEntry) error {
	labels := labels.FromMap(entry.Labels())
	m.appender.Add(labels, entry.Date.Unix(), 1.0)
	fmt.Printf("Adding %v at %d\n", labels, entry.Date.Unix())
	return m.appender.Commit()
}

// Query TODO
func (m *LoggingDatabase) Query(since int64, until int64, matcher labels.Matcher) error {
	q, err := m.db.Querier(since, until)
	if err != nil {
		return nil
	}
	defer q.Close()

	series, err := q.Select(matcher)
	if err != nil {
		return err
	}

	fmt.Printf("Series %v\n", series)
	for series.Next() {
		// Get each Series
		s := series.At()
		fmt.Println("Labels:", s.Labels())
		fmt.Println("Data:")
		it := s.Iterator()
		for it.Next() {
			ts, v := it.At()
			fmt.Println("ts =", ts, "v =", v)
		}
		if err := it.Err(); err != nil {
			panic(err)
		}
	}

	return nil
}

// Cleanup is used to drop all stored data.
func (m *LoggingDatabase) Cleanup() error {
	err := os.RemoveAll(m.db.Dir())
	if err != nil {
		return err
	}

	return m.db.Close()
}
