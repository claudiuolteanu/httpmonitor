package monitor

import (
	"context"
	"fmt"
	"time"
)

// Status represents all supported states of an alert.
type Status int

const (
	// Unknown represents a status not set.
	Unknown Status = iota
	// OK represents the state when the threshold is not reached.
	OK
	// Critical rrepresents the state when the threshold is reached.
	Critical
)

// Alert is used to configure an alert.
type Alert struct {
	checkingInterval time.Duration
	dataInterval     time.Duration
	threshold        float64
	label            string
	pattern          string
	status           Status
	name             string
}

// NewAlert is used to create a new alert.
func NewAlert(name string, checkingInterval time.Duration, dataInterval time.Duration, threshold float64, label string, pattern string) *Alert {
	return &Alert{
		checkingInterval: checkingInterval,
		dataInterval:     dataInterval,
		threshold:        threshold,
		label:            label,
		pattern:          pattern,
		status:           OK,
		name:             name,
	}
}

// CheckStatus is used to update the status of the alert and to notify in case of changes.
// The data from last "dataInterval" seconds and stored under the specified "label" which is matching the "pattern", is
// aggregated each "checkingInterval" seconds. If the result exceeds the "threshold" for the first time, the state of
// alert is changed to Critical and a logging message is displayed. When the result goes below the "threshold" the
// state of alert is moved back to OK and a new logging message is displayed.
func (a *Alert) CheckStatus(db *LoggingDatabase) error {
	now := time.Now()
	since := now.Add(-a.dataInterval)

	// Get the entries from last dataInterval seconds that match the pattern.
	stats, err := db.GetEntries(a.label, a.pattern, since.Unix(), now.Unix())
	if err != nil {
		return err
	}

	// Count total number of hits.
	total := 0.0
	for _, e := range stats {
		total += e.Value
	}

	// Compute the average and check if the threshold was reached.
	average := total / float64(a.dataInterval.Seconds())
	if average >= a.threshold {
		if a.status == OK {
			a.status = Critical
			fmt.Printf("High traffic generated an alert - hits = %f, triggered at %s\n", average, now.Format("2019-12-03 23:48:12 +0200 EET"))
		}
	} else {
		if a.status == Critical {
			a.status = OK
			fmt.Printf("The traffic returned back to normal - hits = %f, at %s\n", average, now.Format("2019-12-03 23:48:12 +0200 EET"))
		}
	}

	return nil
}

// Run is used to monitor and raise an alert. The method is blocking.
func (a *Alert) Run(ctx context.Context, db *LoggingDatabase) error {
	for {
		select {
		case <-time.After(a.checkingInterval):
			err := a.CheckStatus(db)
			if err != nil {
				return err
			}

		case <-ctx.Done():
			fmt.Printf("Stop alert %s\n", a.name)
			return nil
		}
	}
}
