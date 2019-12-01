package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"httpmonitor/monitor"
)

func main() {
	m, err := monitor.NewMonitor("/Users/claudiu.olteanu/Documents/httpmonitor/test.log")
	if err != nil {
		log.Fatal(err)
	}

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	m.Start()
	<-termChan // Blocks here until interrupted
	m.Stop()
}
