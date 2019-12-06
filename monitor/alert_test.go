package monitor

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AlertTestSuite struct {
	suite.Suite
	db *LoggingDatabase
}

func (suite *AlertTestSuite) SetupTest() {
	db, err := NewLoggingDatabase()
	if err != nil {
		log.Fatal()
	}

	suite.db = db
}

func (suite *AlertTestSuite) TearDownTest() {
	suite.db.Cleanup()
}

func (suite *AlertTestSuite) TestAlertTriggered() {
	t := suite.T()

	alert := NewAlert("test", time.Second, 5*time.Second, 1.0, HostLabel, AllEntriesPattern)
	entries := 10 // double the elements to be sure that the threshold is reached

	for i := 0; i < entries; i++ {
		now := time.Now()
		entry := &LoggingEntry{RemoteHost: fmt.Sprintf("127.0.0.%d", i), RemoteLogname: "-", AuthUser: "james", Date: now, Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}
		err := suite.db.AddEntry(entry)
		require.Nil(t, err, "No error should be returned while adding an entry.")
	}

	require.Equal(t, OK, alert.status, "The initial status must be ok.")
	err := alert.CheckStatus(suite.db)
	require.Nil(t, err, "No error should be returned while checking the status.")
	require.Equal(t, Critical, alert.status, "The final status must be critical.")
}

func (suite *AlertTestSuite) TestAlertNotTriggered() {
	t := suite.T()

	alert := NewAlert("test", time.Second, 5*time.Second, 10.0, HostLabel, AllEntriesPattern)
	entries := 1

	for i := 0; i < entries; i++ {
		now := time.Now()
		entry := &LoggingEntry{RemoteHost: fmt.Sprintf("127.0.0.%d", i), RemoteLogname: "-", AuthUser: "james", Date: now, Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}
		err := suite.db.AddEntry(entry)
		require.Nil(t, err, "No error should be returned while adding an entry.")
	}

	require.Equal(t, OK, alert.status, "The initial status must be ok.")
	err := alert.CheckStatus(suite.db)
	require.Nil(t, err, "No error should be returned while checking the status.")
	require.Equal(t, OK, alert.status, "The final status must be ok.")
}

func (suite *AlertTestSuite) TestAlertBackToNormal() {
	t := suite.T()

	alert := NewAlert("test", time.Second, 5*time.Second, 10.0, HostLabel, AllEntriesPattern)
	alert.status = Critical // Manually change the state of the alert to critical
	entries := 1

	for i := 0; i < entries; i++ {
		now := time.Now()
		entry := &LoggingEntry{RemoteHost: fmt.Sprintf("127.0.0.%d", i), RemoteLogname: "-", AuthUser: "james", Date: now, Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}
		err := suite.db.AddEntry(entry)
		require.Nil(t, err, "No error should be returned while adding an entry.")
	}

	require.Equal(t, Critical, alert.status, "The initial status must be critical.")
	err := alert.CheckStatus(suite.db)
	require.Nil(t, err, "No error should be returned while checking the status.")
	require.Equal(t, OK, alert.status, "The final status must be ok.")
}

func (suite *AlertTestSuite) TestAlertRemainsCritical() {
	t := suite.T()

	alert := NewAlert("test", time.Second, 5*time.Second, 1.0, HostLabel, AllEntriesPattern)
	alert.status = Critical // Manually change the state of the alert to critical
	entries := 10           // double the elements to be sure that the threshold is reached

	for i := 0; i < entries; i++ {
		now := time.Now()
		entry := &LoggingEntry{RemoteHost: fmt.Sprintf("127.0.0.%d", i), RemoteLogname: "-", AuthUser: "james", Date: now, Request: &Request{Method: "GET", URL: "/report/user", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}
		err := suite.db.AddEntry(entry)
		require.Nil(t, err, "No error should be returned while adding an entry.")
	}

	require.Equal(t, Critical, alert.status, "The initial status must be critical.")
	err := alert.CheckStatus(suite.db)
	require.Nil(t, err, "No error should be returned while checking the status.")
	require.Equal(t, Critical, alert.status, "The final status must be critical.")
}

func TestAlertTestSuite(t *testing.T) {
	suite.Run(t, new(AlertTestSuite))
}
