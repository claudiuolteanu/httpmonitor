package monitor

import (
	"bytes"
	"fmt"
	"time"
)

const (
	limit = 3
)

// StatsSummary is used to represent traffic statistics from a given interval.
type StatsSummary struct {
	Since           int64
	Until           int64
	Size            int64
	TopSections     EntryList
	TopUsers        EntryList
	RequestMethods  EntryList
	RequestStatuses EntryList
}

// NewStatsSummary is used to generate traffic statistics from a given interval.
func NewStatsSummary(since int64, until int64, db *LoggingDatabase) (*StatsSummary, error) {
	topSections, err := db.TopEntries(RequestURLSectionLabel, AllEntriesPattern, since, until, limit)
	if err != nil {
		return nil, err
	}

	topUsers, err := db.TopEntries(UserLabel, AllEntriesPattern, since, until, limit)
	if err != nil {
		return nil, err
	}

	requestMethods, err := db.TopEntries(RequestMethodLabel, AllEntriesPattern, since, until, 0)
	if err != nil {
		return nil, err
	}

	requestStatuses, err := db.TopEntries(StatusLabel, AllEntriesPattern, since, until, 0)
	if err != nil {
		return nil, err
	}

	return &StatsSummary{
		Since:           since,
		Until:           until,
		TopSections:     topSections,
		TopUsers:        topUsers,
		RequestMethods:  requestMethods,
		RequestStatuses: requestStatuses}, nil
}

func (s *StatsSummary) String() string {
	var stats bytes.Buffer

	var total, clientErrors, serverErrors, redirections, successful int
	for _, e := range s.RequestStatuses {
		total += int(e.Value)
		if "200" <= e.Key && e.Key < "300" {
			successful += int(e.Value)
		} else if "300" <= e.Key && e.Key < "400" {
			redirections += int(e.Value)
		} else if "400" <= e.Key && e.Key < "500" {
			clientErrors += int(e.Value)
		} else if e.Key >= "500" {
			serverErrors += int(e.Value)
		}
	}

	stats.WriteString("------------------------------------------------------------------------------------------------------------------------\n")
	stats.WriteString(fmt.Sprintf("Traffic stats between [%s, %s]:\n", time.Unix(s.Since, 0), time.Unix(s.Until, 0)))
	stats.WriteString(fmt.Sprintf("- From a total of %d requests, there were %d successful calls, %d redirections, %d client errors and %d server errors\n",
		total, successful, redirections, clientErrors, serverErrors))
	stats.WriteString(fmt.Sprintf("- Requests by method: %v\n", s.RequestMethods))
	stats.WriteString(fmt.Sprintf("- Top %d sections: %v\n", limit, s.TopSections))
	stats.WriteString(fmt.Sprintf("- First %d users: %v\n", limit, s.TopUsers))
	stats.WriteString("------------------------------------------------------------------------------------------------------------------------\n")

	return stats.String()
}
