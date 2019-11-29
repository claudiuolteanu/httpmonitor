package main

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrInvalidFormatLine is used to signal that line supplied doesn't match the expected format.
	ErrInvalidFormatLine = errors.New("unsupported timestamp")
)

const (
	// Fields with missing data are represented as "-" (https://en.wikipedia.org/wiki/Common_Log_Format).
	missingData = "-"
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

// LoggingLine holds parsed information about a w3c-formatted HTTP access log (https://www.w3.org/Daemon/User/Config/Logging.html#common-logfile-format).
type LoggingLine struct {
	RemoteHost string
	RemoteUser string
	AuthUser   string
	Date       time.Time
	Request    Request
	Status     int
	Bytes      int
}

func (l *LoggingLine) String() string {
	loggingDate := missingData
	if !l.Date.IsZero() {
		loggingDate = l.Date.String() // TODO: we should use the same format as the one used in reading
	}

	return fmt.Sprintf("%s %s %s [%s] \"%s\" %d %d",
		l.RemoteHost,
		l.RemoteUser,
		l.AuthUser,
		loggingDate,
		l.Request,
		l.Status,
		l.Bytes,
	)
}
