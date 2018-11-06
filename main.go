package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/mihaichiorean/monidog/monitor"
	"github.com/mihaichiorean/monidog/parser"
	"github.com/pkg/profile"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		fmt.Println("profiling cpu")
		defer profile.Start(profile.CPUProfile).Stop()
	}

	// ... rest of the program ...

	if *memprofile != "" {
		fmt.Println("profiling mem")
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	p := parser.AccessLogParser{}
	scanner, err := monitor.Watch("test/access.log", p, 500*time.Millisecond)
	if err != nil {
		panic(err)
	}
	//r := model.NewReporter(10*time.Second, terminal.IsTerminal(int(os.Stdout.Fd())))
	//scanner.AddListener(r)

	//alert := alerts.NewAlert("2min", 1*time.Minute, 10)
	//if err := alert.Start(scanner.Subscribe()); err != nil {
	//	panic(err.Error())
	//}
	//defer alert.Stop()
	ch := scanner.Subscribe()
	counts := 0
	for {
		select {
		case <-ch:
			counts++
			fmt.Println(counts)
		case <-sigs:
			//alert.Stop()
			scanner.Close()
			fmt.Println("See ya!")
			os.Exit(0)
		}
	}
}
