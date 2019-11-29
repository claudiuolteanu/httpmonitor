package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

var (
	// ErrInvalidFormatLine is used to signal that line supplied doesn't match the expected format.
	ErrInvalidFormatLine = errors.New("invalid format line")
	lineRegex            = regexp.MustCompile(`^(\S+)\s` + // remote host
		`([^ ]*)\s` + // remote logname
		`([^ ]*)\s` + // authuser
		`(?:-|\[([^\]]*)\])\s` + // date
		`(?:-|\"(.*)\")\s` + // request
		`(-|[\d]{3})\s` + //status
		`(-|[\d]+)$`) // size
)

const (
	// Fields with missing data are represented as "-" (https://en.wikipedia.org/wiki/Common_Log_Format).
	missingData = "-"

	HostLabel            = "host"
	UserLabel            = "user"
	StatusLabel          = "status"
	SizeLabel            = "size"
	RequestMethodLabel   = "method"
	RequestURLLabel      = "url"
	RequestProtocolLabel = "protocol"
)

// Request contains information about the method used, the URL and the protocol.
type Request struct {
	Method   string
	URL      string
	Protocol string
}

func (r *Request) String() string {
	return fmt.Sprintf("%s %s %s", r.Method, r.URL, r.Protocol)
}

func newRequest(raw string) (*Request, error) {
	entries := strings.Split(raw, " ")
	if len(entries) != 3 {
		return nil, ErrInvalidFormatLine
	}

	return &Request{
		Method:   entries[0],
		URL:      entries[1],
		Protocol: entries[2],
	}, nil
}

// LoggingEntry holds parsed information about a w3c-formatted HTTP access log (https://www.w3.org/Daemon/User/Config/Logging.html#common-logfile-format).
type LoggingEntry struct {
	RemoteHost    string
	RemoteLogname string
	AuthUser      string
	Date          time.Time
	Request       Request
	Status        int
	Bytes         int
}

func (l *LoggingEntry) String() string {
	loggingDate := missingData
	if !l.Date.IsZero() {
		loggingDate = l.Date.String() // TODO: we should use the same format as the one used in reading
	}

	return fmt.Sprintf("%s %s %s [%s] \"%s\" %d %d",
		l.RemoteHost,
		l.RemoteLogname,
		l.AuthUser,
		loggingDate,
		l.Request,
		l.Status,
		l.Bytes,
	)
}

// TODO continue
// Labels returns a map between l
func (l *LoggingEntry) Labels() map[string]string {
	return map[string]string{
		HostLabel:            l.RemoteHost,
		UserLabel:            l.AuthUser,
		RequestMethodLabel:   l.Request.Method,
		RequestURLLabel:      l.Request.URL,
		RequestProtocolLabel: l.Request.Protocol,
	}
}

// NewLoggingEntry creates a LoggingLine from a raw string.
func NewLoggingEntry(raw string) (*LoggingEntry, error) {
	entries := lineRegex.FindStringSubmatch(raw)
	if len(entries) != 8 {
		return nil, ErrInvalidFormatLine
	}

	var err error
	var date time.Time
	if entries[4] != missingData {
		// First check the format "02/Jan/2006:15:04:05 -0700"
		date, err = time.Parse("02/Jan/2006:15:04:05 -0700", entries[4])
		if err != nil {
			// Otherwise try to use dateparse library that is supposed to support multiple formats except the one above :).
			date, err = dateparse.ParseAny(entries[4])
			if err != nil {
				return nil, err
			}
		}
	}

	request, err := newRequest(entries[5])
	if err != nil {
		return nil, err
	}

	var status int
	if entries[6] != missingData {
		status, err = strconv.Atoi(entries[6])
		if err != nil {
			return nil, err
		}
	}

	var bytes int
	if entries[7] != missingData {
		bytes, err = strconv.Atoi(entries[7])
		if err != nil {
			return nil, err
		}
	}

	return &LoggingEntry{
		RemoteHost:    entries[1],
		RemoteLogname: entries[2],
		AuthUser:      entries[3],
		Date:          date,
		Request:       *request,
		Status:        status,
		Bytes:         bytes,
	}, nil
}
