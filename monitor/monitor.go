package monitor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hpcloud/tail"
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

	for {
		select {
		case <-time.After(10 * time.Second):
			now := time.Now()
			since := now.Add(-10 * time.Second)

			err := printLoggingStats(since.Unix(), now.Unix(), m.db)
			if err != nil {
				fmt.Printf("Failed to generate stats data: %v", err)
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
