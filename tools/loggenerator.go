package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func generateLog() string {
	hosts := []string{"127.0.0.1", "172.16.0.1"}
	lognames := []string{"root@test.com", "user@test.com", "-"}
	users := []string{"alex", "george", "chris", "-"}
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	urls := []string{"/user", "/user/subscriptions", "/analytics", "/analytics/reports"}
	protocols := []string{"HTTP/1.2", "HTTP/2.0"}
	statusCodes := []int{200, 201, 202, 300, 301, 400, 401, 403, 404, 405, 500, 503}
	size := rand.Intn(5000)

	return fmt.Sprintf("%s %s %s [%s] \"%s %s %s\" %d %d",
		hosts[rand.Intn(len(hosts))],
		lognames[rand.Intn(len(lognames))],
		users[rand.Intn(len(users))],
		time.Now().Format("02/Jan/2006:15:04:05 -0700"),
		methods[rand.Intn(len(methods))],
		urls[rand.Intn(len(urls))],
		protocols[rand.Intn(len(protocols))],
		statusCodes[rand.Intn(len(statusCodes))],
		size,
	)
}

func generateLogs(ctx context.Context) {
	for {
		select {
		case <-time.After(100 * time.Millisecond):
			fmt.Println(generateLog())
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	go generateLogs(ctx)
	<-termChan // Blocks here until interrupted
	cancelFunc()
}
