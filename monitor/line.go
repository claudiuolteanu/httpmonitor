package monitor

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

	lineRegex = regexp.MustCompile(`^(\S+)\s` + // remote host
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
	// HostLabel is the label used to store hosts into dabatase.
	HostLabel = "host"
	// LogNameLabel is the label used to store lognames into dabatase.
	LogNameLabel = "logname"
	// UserLabel is the label used to store auth users into dabatase.
	UserLabel = "user"
	// StatusLabel is the label used to store requests' statuses into dabatase.
	StatusLabel = "status"
	// RequestMethodLabel is the label used to store requests' methods into dabatase.
	RequestMethodLabel = "method"
	// RequestURLLabel is the label used to store requests' URLs into dabatase.
	RequestURLLabel = "url"
	// RequestURLSectionLabel is the label used to store requests' sections into dabatase.
	RequestURLSectionLabel = "section"
	// RequestProtocolLabel is the label used to store requests' protocols into dabatase.
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

// Section computes what is before the second '/' in the URL's path.
// For example, the section for "/pages/create" is "/pages".
func (r *Request) Section() string {
	firstSeparator := strings.Index(r.URL, "/")
	if firstSeparator == -1 {
		return r.URL
	}

	secondSeparator := strings.Index(r.URL[firstSeparator+1:], "/")
	if secondSeparator == -1 {
		return r.URL
	}

	return r.URL[:firstSeparator+secondSeparator+1]
}

// NewRequest is used to create a request from a raw string.
func NewRequest(raw string) (*Request, error) {
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
	Request       *Request
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

// Labels returns a map between labels used to store into database and their logging values.
func (l *LoggingEntry) Labels() map[string]string {
	return map[string]string{
		HostLabel:              l.RemoteHost,
		LogNameLabel:           l.RemoteLogname,
		UserLabel:              l.AuthUser,
		RequestMethodLabel:     l.Request.Method,
		RequestURLLabel:        l.Request.URL,
		RequestProtocolLabel:   l.Request.Protocol,
		RequestURLSectionLabel: l.Request.Section(),
		StatusLabel:            strconv.Itoa(l.Status),
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
				// TODO wrap the error
				return nil, ErrInvalidFormatLine
			}
		}
	}

	request, err := NewRequest(entries[5])
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
		Request:       request,
		Status:        status,
		Bytes:         bytes,
	}, nil
}
