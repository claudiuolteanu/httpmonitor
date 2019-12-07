package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/hpcloud/tail"
	"golang.org/x/sync/errgroup"
)

// Monitor stores information neccessary to monitor the activity from a logging file.
type Monitor struct {
	db         *LoggingDatabase
	errg       *errgroup.Group
	ctx        context.Context
	cancelFunc context.CancelFunc
	filename   string
	alerts     []*Alert
}

// NewMonitor is used to create a new monitoring for a specifc file, with some configured alerts.
func NewMonitor(filename string, alerts []*Alert) (*Monitor, error) {
	db, err := NewLoggingDatabase()
	if err != nil {
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	errg, ctx := errgroup.WithContext(ctx)

	return &Monitor{db: db, errg: errg, filename: filename, ctx: ctx, cancelFunc: cancelFunc, alerts: alerts}, nil
}

// processLogs reads each line from the file, parses it and inserts it into the database.
func (m *Monitor) processLogs() error {
	t, err := tail.TailFile(m.filename, tail.Config{Follow: true})
	if err != nil {
		return err
	}

	for {
		select {
		case line := <-t.Lines:
			entry, err := NewLoggingEntry(line.Text)
			// Skip invalid entries but don't stop the whole process.
			if err != nil {
				fmt.Printf("Failed to parse one entry %v\n", err)
				continue
			}

			err = m.db.AddEntry(entry)
			if err != nil {
				return err
			}
		case <-m.ctx.Done():
			fmt.Println("Stop processing logs")
			return nil
		}
	}
}

// monitorLogs collects stats from last 10 seconds and prints a summary.
func (m *Monitor) monitorLogs() error {
	for {
		select {
		case <-time.After(10 * time.Second):
			now := time.Now()
			since := now.Add(-10 * time.Second)

			stats, err := NewStatsSummary(since.Unix(), now.Unix(), m.db)
			if err != nil {
				return err
			}

			fmt.Println(stats)
		case <-m.ctx.Done():
			fmt.Println("Stop monitoring logs")
			return nil
		}
	}
}

// Run is used to start the monitoring system.
// It processes the logs and display statistics about the traffic generated in the past 10 seconds.
// Besides that, it also starts to monitor the activiy for configured alerts.
// In order to stop the monitoring, the Stop method should be called.
func (m *Monitor) Run() error {
	m.errg.Go(m.processLogs)
	m.errg.Go(m.monitorLogs)

	for _, a := range m.alerts {
		m.errg.Go(func() error {
			return a.Run(m.ctx, m.db)
		})
	}

	return m.errg.Wait()
}

// Stop is used to stop the monitoring and to do the cleanup.
func (m *Monitor) Stop() error {
	m.cancelFunc()

	return m.db.Cleanup()
}
