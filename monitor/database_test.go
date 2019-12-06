package monitor

import (
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/prometheus/tsdb/labels"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DatabaseTestSuite struct {
	suite.Suite
	db *LoggingDatabase
}

func (suite *DatabaseTestSuite) SetupTest() {
	db, err := NewLoggingDatabase()
	if err != nil {
		log.Fatal()
	}

	suite.db = db
}

func (suite *DatabaseTestSuite) TearDownTest() {
	suite.db.Cleanup()
}

func (suite *DatabaseTestSuite) TestAddEntryl() {
	t := suite.T()
	now := time.Now()
	entry := &LoggingEntry{RemoteHost: "127.0.0.1", RemoteLogname: "-", AuthUser: "james", Date: now, Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}

	err := suite.db.AddEntry(entry)
	require.Nil(t, err, "No error should be returned while adding an entry.")

	matcher, err := labels.NewRegexpMatcher(HostLabel, ".*")
	require.Nil(t, err, "No error should be returned while creating the label matcher.")

	query, err := suite.db.db.Querier(now.Unix(), now.Unix())
	require.Nil(t, err, "No error should be returned while creating the querier.")
	defer query.Close()

	series, err := query.Select(matcher)
	require.Nil(t, err, "No error should be returned while selecting data.")

	expectedLabels := []labels.Label{
		{Name: HostLabel, Value: entry.RemoteHost},
		{Name: LogNameLabel, Value: entry.RemoteLogname},
		{Name: UserLabel, Value: entry.AuthUser},
		{Name: RequestMethodLabel, Value: entry.Request.Method},
		{Name: RequestURLLabel, Value: entry.Request.URL},
		{Name: RequestProtocolLabel, Value: entry.Request.Protocol},
		{Name: RequestURLSectionLabel, Value: entry.Request.Section()},
		{Name: StatusLabel, Value: strconv.Itoa(entry.Status)},
	}

	// Check that the labels were inserted correctly.
	hits := 0.0
	for series.Next() {
		s := series.At()

		it := s.Iterator()
		for it.Next() {
			_, v := it.At()
			hits += v
		}
		if err := it.Err(); err != nil {
			require.Nil(t, err, "No error should be returned while iterating though the series")
		}

		require.ElementsMatch(t, expectedLabels, s.Labels())
	}

	require.Equal(t, 1.0, hits)
}

func (suite *DatabaseTestSuite) TestGetEntries() {
	t := suite.T()
	now := time.Now()
	entry := &LoggingEntry{RemoteHost: "127.0.0.1", RemoteLogname: "-", AuthUser: "james", Date: now, Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}

	err := suite.db.AddEntry(entry)
	require.Nil(t, err, "No error should be returned while adding an entry.")

	entries, err := suite.db.GetEntries(HostLabel, ".*", now.Unix(), now.Unix())
	require.Nil(t, err, "No error should be returned while getting the entries.")

	expectedEntries := []Entry{
		{Key: "127.0.0.1", Value: 1.0},
	}

	require.ElementsMatch(t, expectedEntries, entries)
}

func (suite *DatabaseTestSuite) TestTopEntries() {
	t := suite.T()
	now := time.Now()
	entry1 := &LoggingEntry{RemoteHost: "127.0.0.1", RemoteLogname: "-", AuthUser: "james", Date: now, Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}
	entry2 := &LoggingEntry{RemoteHost: "127.0.0.1", RemoteLogname: "-", AuthUser: "james", Date: now.Add(time.Second), Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}
	entry3 := &LoggingEntry{RemoteHost: "172.16.0.1", RemoteLogname: "-", AuthUser: "james", Date: now, Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}

	err := suite.db.AddEntry(entry1)
	require.Nil(t, err, "No error should be returned while adding an entry.")

	err = suite.db.AddEntry(entry2)
	require.Nil(t, err, "No error should be returned while adding an entry.")

	err = suite.db.AddEntry(entry3)
	require.Nil(t, err, "No error should be returned while adding an entry.")

	entries, err := suite.db.TopEntries(HostLabel, AllEntriesPattern, now.Unix(), now.Add(time.Second).Unix(), 0)
	require.Nil(t, err, "No error should be returned while getting the entries.")

	expectedEntries := []Entry{
		{Key: "127.0.0.1", Value: 2.0},
		{Key: "172.16.0.1", Value: 1.0},
	}

	require.ElementsMatch(t, expectedEntries, entries)
}

func (suite *DatabaseTestSuite) TestTopEntriesWithLimitInBounds() {
	t := suite.T()
	now := time.Now()
	entry1 := &LoggingEntry{RemoteHost: "127.0.0.1", RemoteLogname: "-", AuthUser: "james", Date: now, Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}
	entry2 := &LoggingEntry{RemoteHost: "127.0.0.1", RemoteLogname: "-", AuthUser: "james", Date: now.Add(time.Second), Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}
	entry3 := &LoggingEntry{RemoteHost: "172.16.0.1", RemoteLogname: "-", AuthUser: "james", Date: now, Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}

	err := suite.db.AddEntry(entry1)
	require.Nil(t, err, "No error should be returned while adding an entry.")

	err = suite.db.AddEntry(entry2)
	require.Nil(t, err, "No error should be returned while adding an entry.")

	err = suite.db.AddEntry(entry3)
	require.Nil(t, err, "No error should be returned while adding an entry.")

	entries, err := suite.db.TopEntries(HostLabel, AllEntriesPattern, now.Unix(), now.Add(time.Second).Unix(), 1)
	require.Nil(t, err, "No error should be returned while getting the entries.")

	expectedEntries := []Entry{
		{Key: "127.0.0.1", Value: 2.0},
	}

	require.ElementsMatch(t, expectedEntries, entries)
}

func (suite *DatabaseTestSuite) TestTopEntriesWithLimitOutBounds() {
	t := suite.T()
	now := time.Now()
	entry1 := &LoggingEntry{RemoteHost: "127.0.0.1", RemoteLogname: "-", AuthUser: "james", Date: now, Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}
	entry2 := &LoggingEntry{RemoteHost: "127.0.0.1", RemoteLogname: "-", AuthUser: "james", Date: now.Add(time.Second), Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}
	entry3 := &LoggingEntry{RemoteHost: "172.16.0.1", RemoteLogname: "-", AuthUser: "james", Date: now, Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}

	err := suite.db.AddEntry(entry1)
	require.Nil(t, err, "No error should be returned while adding an entry.")

	err = suite.db.AddEntry(entry2)
	require.Nil(t, err, "No error should be returned while adding an entry.")

	err = suite.db.AddEntry(entry3)
	require.Nil(t, err, "No error should be returned while adding an entry.")

	entries, err := suite.db.TopEntries(HostLabel, AllEntriesPattern, now.Unix(), now.Add(time.Second).Unix(), 5)
	require.Nil(t, err, "No error should be returned while getting the entries.")

	expectedEntries := []Entry{
		{Key: "127.0.0.1", Value: 2.0},
		{Key: "172.16.0.1", Value: 1.0},
	}

	require.ElementsMatch(t, expectedEntries, entries)
}

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}
