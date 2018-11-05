package monitor

import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"

	"github.com/mihaichiorean/monidog/parser"
	"github.com/pkg/errors"
)

// CancelFunc is the type used to describe a function that will stop the access log scanner
type CancelFunc func()

// LogScanner exposes the vehaviour we want from a log scanner. Since it can be implemented in
// multiple ways, this could come handy when replacing implementations
type LogScanner interface {
	Channel() chan parser.Log
	Start(path string) (CancelFunc, error)
}

type logScanner struct {
	listeners []chan parser.Log
	interval  time.Duration
	parser    parser.LogParser
}

// NewLogScanner takes a LogParser and an interval to build a LogScanner
func NewLogScanner(p parser.LogParser, interval time.Duration) LogScanner {
	l := logScanner{
		listeners: make([]chan parser.Log, 0, 10),
		interval:  interval,
		parser:    p,
	}
	return &l
}

// Channel is a factory that creates a new buffered channel to send logs on. Since this is the sender,
// the channel is meant to be closed by the log scanner.
func (ls *logScanner) Channel() chan parser.Log {
	ch := make(chan parser.Log, 10)
	ls.listeners = append(ls.listeners, ch)
	return ch
}

// stop will cancel log updates and close the listeners
func (ls *logScanner) stop() {
	for _, c := range ls.listeners {
		close(c)
	}
}

func (ls *logScanner) parseLog(line string) (parser.Log, error) {
	log, err := ls.parser.Parse(line)
	if err != nil {
		return nil, errors.Wrapf(err, "failed parsing log. [%s]", line)
	}
	return log, nil
}

func (ls *logScanner) notifyListeners(lg parser.Log) {
	for _, l := range ls.listeners {
		l <- lg
	}
}

// Start will begin watching a designated file in read only mode
// and return a cancel/stop function or error if it was unable to start watching
func (ls *logScanner) Start(path string) (CancelFunc, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0755)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read from path %s", path)
	}

	done := make(chan struct{})
	cancel := func() {
		done <- struct{}{}
	}

	// read lines of logs
	read := func(f *os.File, wg *sync.WaitGroup) {
		wg.Add(1)
		scanner := bufio.NewScanner(f)
		go func() {
			for scanner.Scan() {
				// (IMPROVEMENT) could probably implement a scanner that returns a log struct instead of string
				t := scanner.Text()
				l, err := ls.parseLog(t)
				if err != nil {
					// TODO log error
					break
				}
				ls.notifyListeners(l)
			}
			if err := scanner.Err(); err != nil {
				//fmt.Fprintln(os.Stderr, "reading standard input:", err)
				//ls.notifyListeners("some error")
				// TODO log error
			}
			wg.Done()
		}()
	}
	go func() {
		var wg sync.WaitGroup
		stats, err := f.Stat()
		if err != nil {
			//fmt.Fprintf(os.Stderr, "file stats failed: %s", err.Error())
			panic(err)
			// TODO send error to listeners
			// TODO close listeners
			return
		}
		// waiting for new content
		t := time.NewTicker(ls.interval)
		for {
			select {
			case <-t.C:
				pos, err := f.Seek(0, io.SeekCurrent)
				if err != nil {
					// TODO be more graceful
					panic(err)
				}
				s, err := f.Stat()
				if err != nil {
					// TODO be more graceful
					panic(err)
				}
				if s.Size() == stats.Size() {
					time.Sleep(ls.interval)
					break
				}

				// file size has grown, move reading position
				if s.Size() > stats.Size() {
					f.Seek(pos, io.SeekStart)
					stats = s
				}

				// if the file is now smaller, start from the beginning
				if s.Size() < stats.Size() {
					f.Seek(0, 0)
					stats = s
				}

				// read fresh content
				read(f, &wg)
				wg.Wait()

			case <-done:
				// Abort the walk if done is closed.
				t.Stop()
				//ls.notifyListeners("exit called")
				// TODO log something
				return
			default:
			}
		}
	}()
	return cancel, nil
}
