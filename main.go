package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mihaichiorean/monidog/alerts"
	"github.com/mihaichiorean/monidog/monitor"
	"github.com/mihaichiorean/monidog/parser"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	p := parser.AccessLogParser{}
	scanner := monitor.NewLogScanner(p, 500*time.Millisecond)
	scannerCancel, err := scanner.Start("access.log")
	if err != nil {
		panic(err)
	}
	defer scannerCancel()
	//r := model.NewReporter(10*time.Second, terminal.IsTerminal(int(os.Stdout.Fd())))
	//scanner.AddListener(r)

	alert := alerts.NewAlert("2min", 1*time.Minute, 10)
	if err := alert.Start(scanner.Channel()); err != nil {
		panic(err.Error())
	}
	defer alert.Stop()
	for {
		select {
		case <-sigs:
			fmt.Println("See ya!")
			scannerCancel()
			alert.Stop()
			os.Exit(0)
		}
	}
}
