package model

import "time"

// Bucket is a struct that holds counters and stats for a time period
type Bucket struct {
	Ts       time.Time
	counters map[string]int
}

func NewBucket(ts time.Time) *Bucket {
	bucket := Bucket{
		Ts:       ts,
		counters: map[string]int{},
	}
	return &bucket
}

// Inc increments a counter in this bucket
func (b *Bucket) Inc(s string) int {
	v, ok := b.counters[s]
	if !ok {
		b.counters[s] = 0
		v = 0
	}
	v += 1
	b.counters[s] = v
	return v
}

// Counters returns a copy of the internal counters map
func (b *Bucket) Counters() map[string]int {
	m := make(map[string]int)
	for k, v := range b.counters {
		m[k] = v
	}
	return m
}
