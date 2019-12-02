package monitor

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRequestSection(t *testing.T) {
	requestWithoutDashes := &Request{URL: "pages"}
	require.Equal(t, "pages", requestWithoutDashes.Section(), "Unexpected section")

	requestWithOneDash := &Request{URL: "/pages"}
	require.Equal(t, "/pages", requestWithOneDash.Section(), "Unexpected section")

	requestWithMultipleDashes := &Request{URL: "/pages/title/create"}
	require.Equal(t, "/pages", requestWithMultipleDashes.Section(), "Unexpected section")
}

func TestLoggingEntryInvalidFormat(t *testing.T) {
	entry, err := NewLoggingEntry("invalid")

	require.Nil(t, entry, "Unexpected logging entry")
	require.Equal(t, ErrInvalidFormatLine, err, "Unexpected error")
}

func TestLoggingEntryInvalidDate(t *testing.T) {
	entry, err := NewLoggingEntry("127.0.0.1 - james [40/May/2018:16:00:39 ] \"GET /report HTTP/1.0\" 200 123")

	require.Nil(t, entry, "Unexpected logging entry")
	require.Equal(t, ErrInvalidFormatLine, err, "Unexpected error")
}

func TestLoggingEntry(t *testing.T) {
	expectedTime, err := time.Parse("02/Jan/2006:15:04:05 -0700", "09/May/2018:16:00:39 +0000")
	require.Nil(t, err, "Unexpected error raised")
	expectedEntry := &LoggingEntry{RemoteHost: "127.0.0.1", RemoteLogname: "-", AuthUser: "james", Date: expectedTime, Request: &Request{Method: "GET", URL: "/report", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}
	entry, err := NewLoggingEntry("127.0.0.1 - james [09/May/2018:16:00:39 +0000] \"GET /report HTTP/1.0\" 200 123")

	require.Equal(t, expectedEntry, entry, "Unexpected logging entry")
	require.Nil(t, err, "Unexpected error raised")
}

func TestLoggingEntryParseAny(t *testing.T) {
	expectedTime, err := time.Parse(time.RFC822Z, "09 May 18 16:00 +0000")
	require.Nil(t, err, "Unexpected error raised")
	expectedEntry := &LoggingEntry{RemoteHost: "127.0.0.1", RemoteLogname: "-", AuthUser: "james", Date: expectedTime, Request: &Request{Method: "GET", URL: "/report", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}
	entry, err := NewLoggingEntry("127.0.0.1 - james [09 May 2018 16:00 +0000] \"GET /report HTTP/1.0\" 200 123")

	require.Equal(t, expectedEntry, entry, "Unexpected logging entry")
	require.Nil(t, err, "Unexpected error raised")
}

func TestLoggingEntryInvalidRequest(t *testing.T) {
	entry, err := NewLoggingEntry("127.0.0.1 - james [09 May 2018 16:00 +0000] \"GET /report\" 200 123")

	require.Nil(t, entry, "Unexpected logging entry")
	require.Equal(t, ErrInvalidFormatLine, err, "Unexpected error")
}

func TestLoggingEntryInvalidStatus(t *testing.T) {
	entry, err := NewLoggingEntry("127.0.0.1 - james [09 May 2018 16:00 +0000] \"GET /report HTTP/1.0\" 20 123")

	require.Nil(t, entry, "Unexpected logging entry")
	require.Equal(t, ErrInvalidFormatLine, err, "Unexpected error")
}

func TestLoggingEntryInvalidSize(t *testing.T) {
	entry, err := NewLoggingEntry("127.0.0.1 - james [09 May 2018 16:00 +0000] \"GET /report HTTP/1.0\" 123 invalid")

	require.Nil(t, entry, "Unexpected logging entry")
	require.Equal(t, ErrInvalidFormatLine, err, "Unexpected error")
}

func TestLabels(t *testing.T) {
	entry := &LoggingEntry{RemoteHost: "127.0.0.1", RemoteLogname: "-", AuthUser: "james", Date: time.Now(), Request: &Request{Method: "GET", URL: "/report", Protocol: "HTTP/1.0"}, Status: 200, Bytes: 123}
	expectedLabels := map[string]string{
		HostLabel:            entry.RemoteHost,
		LogNameLabel:         entry.RemoteLogname,
		UserLabel:            entry.AuthUser,
		RequestMethodLabel:   entry.Request.Method,
		RequestURLLabel:      entry.Request.URL,
		RequestProtocolLabel: entry.Request.Protocol,
		StatusLabel:          strconv.Itoa(entry.Status),
	}

	require.True(t, reflect.DeepEqual(expectedLabels, entry.Labels()), "Unexpected labels")
}
