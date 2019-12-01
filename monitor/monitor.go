package monitor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hpcloud/tail"
	"github.com/prometheus/tsdb/labels"
)

type Monitor struct {
	db         *LoggingDatabase
	wg         *sync.WaitGroup
	ctx        context.Context
	cancelFunc context.CancelFunc
	filename   string
}

func NewMonitor(filename string) (*Monitor, error) {
	db, err := NewLoggingDatabase()
	if err != nil {
		return nil, err
	}

	wg := &sync.WaitGroup{}
	ctx, cancelFunc := context.WithCancel(context.Background())

	return &Monitor{db: db, wg: wg, filename: filename, ctx: ctx, cancelFunc: cancelFunc}, nil
}

func (m *Monitor) processLogs() {
	defer m.wg.Done()

	t, err := tail.TailFile(m.filename, tail.Config{Follow: true})
	if err != nil {
		// TODO return error (use errors channel)
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

func (m *Monitor) monitorLogs() {
	defer m.wg.Done()

	matcher, err := labels.NewRegexpMatcher(RequestURLLabel, ".*")
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-time.After(10 * time.Second):
			now := time.Now().Unix()
			series, err := m.db.Query(now-int64(10*time.Second), now, matcher)

			if err != nil {
				fmt.Printf("Failed to sync data: %v", err)
			}

			stats, err := GroupByLabel(series, RequestURLLabel)
			if err != nil {
				fmt.Printf("Failed to sync data: %v", err)
			} else {
				fmt.Printf("Stats about %s: %v\n", RequestURLLabel, stats)
			}
		case <-m.ctx.Done():
			fmt.Println("Stop monitoring logs")
			return
		}
	}
}

func (m *Monitor) Start() {
	m.wg.Add(2)
	go m.processLogs()
	go m.monitorLogs()
}

func (m *Monitor) Stop() error {
	m.cancelFunc()
	m.wg.Wait()

	return m.db.Cleanup()
}
