package monitor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hpcloud/tail"
)

// Monitor stores information neccessary to monitor the activity from a logging file.
type Monitor struct {
	db         *LoggingDatabase
	wg         *sync.WaitGroup
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

	wg := &sync.WaitGroup{}
	ctx, cancelFunc := context.WithCancel(context.Background())

	return &Monitor{db: db, wg: wg, filename: filename, ctx: ctx, cancelFunc: cancelFunc, alerts: alerts}, nil
}

// processLogs reads each line from the file, parses it and inserts it into the database.
func (m *Monitor) processLogs() {
	defer m.wg.Done()

	t, err := tail.TailFile(m.filename, tail.Config{Follow: true})
	if err != nil {
		// TODO return error (use channel errors)
		log.Fatal(err)
	}

	for {
		select {
		case line := <-t.Lines:
			entry, err := NewLoggingEntry(line.Text)
			if err != nil {
				fmt.Println(err)
			} else {
				err = m.db.AddEntry(entry)
				if err != nil {
					fmt.Println(err)
				}
			}
		case <-m.ctx.Done():
			fmt.Println("Stop processing logs")
			return
		}
	}
}

// monitorLogs collects stats from last 10 seconds and prints a summary.
func (m *Monitor) monitorLogs() {
	defer m.wg.Done()

	for {
		select {
		case <-time.After(10 * time.Second):
			now := time.Now()
			since := now.Add(-10 * time.Second)

			stats, err := NewStatsSummary(since.Unix(), now.Unix(), m.db)
			if err != nil {
				fmt.Printf("Failed to generate stats data: %v\n", err)
			} else {
				fmt.Print(stats)
			}
		case <-m.ctx.Done():
			fmt.Println("Stop monitoring logs")
			return
		}
	}
}

// Run is used to start the monitoring system in background.
// It processes the logs and display statistics about the traffic generated in the past 10 seconds.
// Besides that, it also starts to monitor the activiy for configured alerts.
// In order to stop the monitoring, the Stop method should be called.
func (m *Monitor) Run() {
	m.wg.Add(2)
	go m.processLogs()
	go m.monitorLogs()

	for _, a := range m.alerts {
		m.wg.Add(1)
		go a.Run(m.ctx, m.wg, m.db)
	}
}

// Stop is used to stop the monitoring and to do the cleanup.
func (m *Monitor) Stop() error {
	m.cancelFunc()
	m.wg.Wait()

	return m.db.Cleanup()
}
