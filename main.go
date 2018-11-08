package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mihaichiorean/monidog/monitor"
	"github.com/mihaichiorean/monidog/parser"
	"github.com/mihaichiorean/monidog/reporter"
	"github.com/pkg/errors"
)

var logfilePath = flag.String("logfile", "/var/log/access.log", "log file to tail")

func main() {
	flag.Parse()

	// handle kill signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	p := parser.AccessLogParser{}
	filePath := *logfilePath
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		log.Fatal("Unable to open log file", errors.Wrapf(err, "failed to read from path %s", filePath))
	}

	// start a log scanner and check for updates every second
	scanner, err := monitor.Watch(f, p, 1000*time.Millisecond)
	if err != nil {
		log.Fatal(err.Error())
	}
	r := reporter.NewReporter(10 * time.Second)
	closeReporter := r.Start(scanner.Subscribe())
	//alert := alerts.NewAlert("2min", 1*time.Minute, 10)
	//if err := alert.Start(scanner.Subscribe()); err != nil {
	//	panic(err.Error())
	//}

	for {
		select {
		case <-sigs:
			closeReporter()
			scanner.Close()
			fmt.Println("See ya!")
			os.Exit(0)
		}
	}
}
