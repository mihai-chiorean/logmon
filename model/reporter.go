package model

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	axs "github.com/Songmu/axslogparser"
	"github.com/mihaichiorean/monidog/parser"
	"github.com/pkg/errors"
)

// Reporter is a struct that gathers stats for a time period of logs
type Reporter struct {
	reportWindow time.Duration
	bucketMS     time.Duration
	buckets      []Bucket
	in           chan parser.Log
}

// NewReporter is the factory function for a new reporter.
// intervalSize is the intervals at which we want it to report
// historySize is how much do we want to go back in time and cache
func NewReporter(window time.Duration, isTerminal bool) *Reporter {
	// opinionated option to choose a 10 bucket granularity/accuracy level. Should be customizable
	bucketSize := (window / 10)
	r := Reporter{
		reportWindow: window,
		bucketMS:     bucketSize,
		buckets:      make([]Bucket, 0, 10),
		in:           make(chan parser.Log),
	}
	return &r
}

// Channel implements the monitor.Listener. it returns the channel on which the reporter is listening on
func (r *Reporter) Channel() chan parser.Log {
	return r.in
}

// clear removes all old counts from the bucketlist
func (r *Reporter) clear(_ time.Time) {
	ts := time.Now()
	cutoff := ts.Add(-(r.bucketMS * 9))

	for i, b := range r.buckets {
		if b.ts.Unix() >= cutoff.Unix() {
			if i > 0 && r.buckets[i-1].ts.Unix() < cutoff.Unix() {
				r.buckets = r.buckets[i:]
				break
			}
		}
	}
}

// incSection increments a section's counts
func (r *Reporter) incSection(s string, ts time.Time) int {
	cutoff := time.Now().Add(-(r.reportWindow))
	if ts.Before(cutoff) {
		// this log is too old. discard
		return 0
	}
	r.clear(ts)
	bucketTS := ts.Truncate(r.bucketMS)
	l := len(r.buckets)
	for i := l - 1; i >= 0; i-- {
		b := r.buckets[i]
		if b.ts == bucketTS {
			return b.Inc(s)
		}
	}

	// new bucket
	bucket := Bucket{
		ts: bucketTS,
		counters: map[string]int{
			s: 1,
		},
	}
	r.buckets = append(r.buckets, bucket)
	return 1
}

// parseSection reeturns a section given a path string
func parseSection(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", errors.Wrapf(err, "unable to parse uri: %s", s)
	}

	// remove trailing slash
	str := u.Path
	if len(u.Path) > 1 {
		str = strings.TrimRight(u.Path, "/")
	}

	// section is whatever is before the final slash
	i := strings.LastIndex(str, "/")
	if i > 1 {
		str = str[:i]
	}
	return str, nil
}

// seectionStats returns a map of sections and the number of hits they got in the previous window
func (r *Reporter) sectionStats() map[string]int {
	totals := map[string]int{}
	for _, b := range r.buckets {
		c := b.Counters()
		for k, v := range c {
			if _, ok := totals[k]; !ok {
				totals[k] = 0
			}
			totals[k] += v
		}
	}
	return totals
}

// hotSection returns the section with the most hits in the previous interval of logs
func (r *Reporter) hotSection() (string, int) {
	totals := r.sectionStats()
	max := 0
	section := ""
	for k, v := range totals {
		if v > max {
			max = v
			section = k
		}
	}

	return section, max
}

func (r *Reporter) add(l *axs.Log) error {
	sec, err := parseSection(l.RequestURI)
	if err != nil {
		return errors.Wrap(err, "Add to reporter failed")
	}
	r.incSection(sec, l.Time)
	return nil
}

// Start triggers the async flow of printing stats. returns a function used to stop the reporter from printing
func (r *Reporter) Start() func() {
	done := make(chan struct{})
	cancel := func() {
		done <- struct{}{}
	}
	go func() {
		t := time.NewTicker(r.reportWindow)
		for {
			select {
			case log := <-r.in:
				l := reflect.ValueOf(log).Interface().(*axs.Log)
				r.add(l)
			case <-t.C:
				r.PrintSectionStats()
			case <-done:
				return
			default:
			}
		}
	}()
	return cancel
}

// PrintSectionStats shows the section with the most hits
func (r *Reporter) PrintSectionStats() {
	fmt.Println("--------------------------------------")
	t := r.sectionStats()
	i := 0
	for s, v := range t {
		// clearing line by printing empty space over it
		// print section stats on thee ith line
		fmt.Println(s, v)
		i++
	}
	sec, count := r.hotSection()
	fmt.Println("highest hits section: ", sec, count)
}
