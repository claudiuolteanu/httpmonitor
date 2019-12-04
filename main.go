package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"httpmonitor/monitor"
)

func main() {
	alerts := []*monitor.Alert{monitor.NewAlert("Traffic from last 2 minutes", 5*time.Second, 2*time.Minute, 3.0, monitor.RequestMethodLabel, ".*")}
	m, err := monitor.NewMonitor("/Users/claudiu.olteanu/Documents/httpmonitor/test.log", alerts)
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
