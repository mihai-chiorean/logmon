package alerts

import (
	"fmt"
	"reflect"
	"time"

	axs "github.com/Songmu/axslogparser"
	"github.com/mihaichiorean/monidog/parser"
)

// Alert -
type Alert struct {
	name     string
	window   time.Duration
	bucketMS time.Duration
	limit    int
	buckets  map[time.Time]int
	total    int
	active   bool
	cancel   func()
}

// NewAlert constructs a new alert with given name, window and alert threshold
func NewAlert(name string, window time.Duration, trigger int) *Alert {
	a := Alert{
		name:     name,
		window:   window,
		bucketMS: window / 100,
		limit:    trigger,
		buckets:  map[time.Time]int{},
		total:    0,
		active:   false,
	}
	return &a
}

func (a *Alert) Start(in <-chan parser.Log) error {
	if a.cancel != nil {
		return fmt.Errorf("%s alert already started", a.name)
	}
	done := make(chan struct{})
	cancel := func() {
		done <- struct{}{}
		close(done)
	}
	a.cancel = cancel
	// cleanup old log counters every bucketMS l
	t := time.NewTicker(a.bucketMS)
	defer t.Stop()
	for {
		select {
		case log, ok := <-in:
			if !ok {
				// channel closed
				a.Stop()
				continue
			}
			l := reflect.ValueOf(log).Interface().(*axs.Log)
			a.Inc(l.Time)
			a.CheckAndAlert()
		case <-t.C:
			a.clear()
			a.CheckAndAlert()
		case <-done:
			return nil
		}
	}
	return nil
}

// Stop will cancel an alert
func (a *Alert) Stop() error {
	if a.cancel == nil {
		return fmt.Errorf("cannot stop %s alert. not started yet")
	}
	a.cancel()
	a.cancel = nil
	return nil
}

func (a *Alert) clear() {
	cutoff := time.Now().Add(-(a.window))

	old := []time.Time{}
	// cleanup
	for k, _ := range a.buckets {
		if k.Unix() < cutoff.Unix() {
			old = append(old, k)
		}
	}

	for _, v := range old {
		count, _ := a.buckets[v]
		delete(a.buckets, v)
		a.total -= count
	}
	if a.active == true && a.total < a.limit {
		a.active = false
	}
}

func (a *Alert) Inc(ts time.Time) {
	a.clear()
	cutoff := time.Now().Add(-(a.window))
	if ts.Before(cutoff) {
		return
	}
	k := ts.Truncate(a.bucketMS)
	if _, ok := a.buckets[k]; !ok {
		a.buckets[k] = 0
	}
	a.buckets[k] += 1
	a.total += 1
}

func (a *Alert) CheckAndAlert() {
	if a.total >= a.limit && a.active == false {
		a.active = true
		fmt.Printf("!!!! %s:  alert triggered !!!!\n", a.name)
	}
}
