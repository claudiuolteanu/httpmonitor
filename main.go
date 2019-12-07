package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"httpmonitor/monitor"
)

func main() {
	filename := flag.String("filename", "/tmp/access.log", "path to HTTP access log")
	threshold := flag.Float64("threshold", 10.0, "number of requests per second that needs to be exceeded to generate an alert")
	flag.Parse()

	alerts := []*monitor.Alert{
		monitor.NewAlert(
			fmt.Sprintf("Traffic from last 2 minutes with %f threshold", *threshold),
			5*time.Second,
			2*time.Minute,
			*threshold,
			monitor.RequestMethodLabel,
			monitor.AllEntriesPattern,
		),
	}

	m, err := monitor.NewMonitor(*filename, alerts)
	if err != nil {
		log.Fatal(err)
	}

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	m.Run()
	<-termChan // Blocks here until interrupted
	m.Stop()
}
