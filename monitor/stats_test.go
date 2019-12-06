package monitor

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StatsTestSuite struct {
	suite.Suite
	db *LoggingDatabase
}

func (suite *StatsTestSuite) SetupTest() {
	db, err := NewLoggingDatabase()
	if err != nil {
		log.Fatal()
	}

	suite.db = db
}

func (suite *StatsTestSuite) TearDownTest() {
	suite.db.Cleanup()
}

func (suite *StatsTestSuite) TestStatsSummary() {
	t := suite.T()
	now := time.Now()
	entry1 := &LoggingEntry{RemoteHost: "127.0.0.1", RemoteLogname: "-", AuthUser: "james1", Date: now, Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}
	entry2 := &LoggingEntry{RemoteHost: "127.0.0.1", RemoteLogname: "-", AuthUser: "james2", Date: now.Add(time.Second), Request: &Request{Method: "GET", URL: "/home", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}
	entry3 := &LoggingEntry{RemoteHost: "172.16.0.1", RemoteLogname: "-", AuthUser: "james3", Date: now, Request: &Request{Method: "POST", URL: "/report/summary", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}

	err := suite.db.AddEntry(entry1)
	require.Nil(t, err, "No error should be returned while adding an entry.")

	err = suite.db.AddEntry(entry2)
	require.Nil(t, err, "No error should be returned while adding an entry.")

	err = suite.db.AddEntry(entry3)
	require.Nil(t, err, "No error should be returned while adding an entry.")

	since, until := now.Unix(), now.Add(time.Second).Unix()

	expectedSummary := &StatsSummary{
		Since:           since,
		Until:           until,
		TopSections:     []Entry{{Key: "/report", Value: 2.0}, {Key: "/home", Value: 1.0}},
		TopUsers:        []Entry{{Key: "james1", Value: 1.0}, {Key: "james2", Value: 1.0}, {Key: "james3", Value: 1.0}},
		RequestMethods:  []Entry{{Key: "GET", Value: 2.0}, {Key: "POST", Value: 1.0}},
		RequestStatuses: []Entry{{Key: "200", Value: 3.0}},
	}
	summary, err := NewStatsSummary(since, until, suite.db)
	require.Nil(t, err, "No error should be returned while computing stats.")

	require.Equal(t, since, summary.Since)
	require.Equal(t, until, summary.Until)
	require.ElementsMatch(t, expectedSummary.TopSections, summary.TopSections)
	require.ElementsMatch(t, expectedSummary.TopUsers, summary.TopUsers)
	require.ElementsMatch(t, expectedSummary.RequestMethods, summary.RequestMethods)
	require.ElementsMatch(t, expectedSummary.RequestStatuses, summary.RequestStatuses)
}

func TestStatsTestSuite(t *testing.T) {
	suite.Run(t, new(StatsTestSuite))
}
