package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hpcloud/tail"
	"github.com/prometheus/tsdb/labels"
)

func processLine(line string, db *LoggingDatabase) error {
	entry, err := NewLoggingEntry(line)
	if err != nil {
		return err
	}

	return db.AddEntry(entry)
}

func processLogs(ctx context.Context, wg *sync.WaitGroup, filename string, db *LoggingDatabase) {
	defer wg.Done()

	t, err := tail.TailFile(filename, tail.Config{Follow: true})
	if err != nil {
		// TODO return error
		log.Fatal(err)
	}

	for {
		select {
		case line := <-t.Lines:
			err := processLine(line.Text, db)
			if err != nil {
				fmt.Println(err)
			}
		case <-ctx.Done():
			fmt.Println("Stop processing logs")
			return
		}
	}
}

func monitorLogs(ctx context.Context, wg *sync.WaitGroup, db *LoggingDatabase) {
	defer wg.Done()

	m := labels.NewEqualMatcher(RequestMethodLabel, "GET")

	for {
		select {
		case <-time.After(10 * time.Second):
			err := db.Query(1571047299, 1576061842, m)

			if err != nil {
				fmt.Printf("Failed to sync data: %v", err)
			}
		case <-ctx.Done():
			fmt.Println("Stop monitoring logs")
			return
		}
	}
}

func main() {
	db, err := NewLoggingDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Cleanup()

	// Set up cancellation context and waitgroup
	ctx, cancelFunc := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	wg.Add(2)
	go processLogs(ctx, wg, "/Users/claudiu.olteanu/Documents/httpmonitor/test.log", db)
	go monitorLogs(ctx, wg, db)

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	<-termChan // Blocks here until interrupted

	// Handle shutdown
	fmt.Println("*********************************\nShutdown signal received\n*********************************")
	cancelFunc() // Signal cancellation to context.Context
	wg.Wait()    // Block here until are workers are done
}
