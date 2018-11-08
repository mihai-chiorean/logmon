package monitor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mihaichiorean/monidog/parser"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type SeekReader interface {
	io.Reader
	io.Seeker
	Stat() (os.FileInfo, error)
}

// LogScanner exposes the vehaviour we want from a log scanner. Since it can be implemented in
// multiple ways, this could come handy when replacing implementations
type LogScanner interface {
	Subscribe() <-chan parser.Log
	Close() error
}

// Watch will start watching a file, scan and parse new logs
func Watch(f SeekReader, p parser.LogParser, every time.Duration) (LogScanner, error) {
	log := zap.NewExample().Sugar()
	defer log.Sync()
	if p == nil {
		return nil, fmt.Errorf("parser is required to handle the file, nil provided")
	}
	ls := logScanner{
		SugaredLogger: log,
		interval:      every,
		subscribing:   make(chan chan parser.Log),
		closing:       make(chan chan error),
		parser:        p,
	}
	go ls.loop(f)
	return &ls, nil
}

type logScanner struct {
	*zap.SugaredLogger
	f           *os.File
	interval    time.Duration
	parser      parser.LogParser
	subscribing chan chan parser.Log
	closing     chan chan error
}

// Subscribe creates a new channel for the client caller and passes that to the worker
// routine, returning the channel to the client for reading
func (ls *logScanner) Subscribe() <-chan parser.Log {
	ch := make(chan parser.Log, 10)
	ls.subscribing <- ch
	return ch
}

func (ls *logScanner) Close() error {
	errc := make(chan error)
	ls.closing <- errc
	err := <-errc
	return err
}

func (ls *logScanner) parseLog(line string) (parser.Log, error) {
	log, err := ls.parser.Parse(line)
	if err != nil {
		return nil, errors.Wrapf(err, "failed parsing log. [%s]", line)
	}
	return log, nil
}

func (ls *logScanner) readLines(f io.Reader) ([]parser.Log, error) {
	newLogs := []parser.Log{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// (IMPROVEMENT) could probably implement a scanner that returns a log struct instead of string
		t := scanner.Text()
		l, err := ls.parseLog(t)
		if err != nil {
			ls.With(
				zap.Error(err),
				zap.String("line", t),
			).Debug("Failed to parse log line")
			break
		}
		newLogs = append(newLogs, l)
	}
	if err := scanner.Err(); err != nil {
		ls.With(zap.Error(err)).Warn("file scanning failed")
		return newLogs, err
	}
	return newLogs, nil
}

func (ls *logScanner) hasChanged(f SeekReader, stats os.FileInfo) (int64, bool, error) {

	pos, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return pos, false, err
	}

	s, err := f.Stat()
	if err != nil {
		return pos, false, err
	}
	if s.Size() == stats.Size() {
		return pos, false, nil
	}
	// file size has grown, move reading position
	if s.Size() > stats.Size() {
		stats = s
		return pos, true, nil
	}
	stats = s
	// if the file is now smaller, start from the beginning
	return int64(0), true, nil
}

// loop will begin watching a designated file in read only mode
// and return a cancel/stop function or error if it was unable to start watching
func (ls *logScanner) loop(f SeekReader) {
	// define some state
	subscribers := []chan parser.Log{}
	stats, err := f.Stat()
	if err != nil {
		ls.With(zap.Error(err)).Fatal("Failed to read stats for file")
		//fmt.Fprintf(os.Stderr, "file stats failed: %s", err.Error())
		panic(err)
		return
	}

	// channel used to trigger sending logs to subscribers
	updates := make(chan parser.Log)

	// waiting for new content
	var tick time.Time
	// already parsed logs
	var queue []parser.Log
	for {
		var delay time.Duration
		var u chan parser.Log
		var head parser.Log

		// channel used to trigger an new check of the file stats
		if now := time.Now(); tick.After(now) {
			delay = tick.Sub(now)
		}
		check := time.After(delay)

		if len(queue) > 0 {
			u = updates
			head = queue[0]
			go func() {
				u <- head
			}()
		}

		select {
		// subscribe task
		case subc := <-ls.subscribing:
			subscribers = append(subscribers, subc)
		// check for file changes task
		case <-check:
			// set the next tick when to check the file for changes
			tick = time.Now().Add(ls.interval)
			pos, changed, err := ls.hasChanged(f, stats)
			if err != nil {
				ls.Fatalf("cannot read log file stats. %s", err.Error())
			}
			if changed == false {
				break
			}

			if changed && pos > 0 {
				f.Seek(pos, io.SeekStart)
			}

			if changed && pos == 0 {
				f.Seek(0, 0)
			}

			// read fresh content
			logLines, err := ls.readLines(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
			}
			if len(logLines) > 0 {
				queue = append(queue, logLines...)
			}
		// send updates to subscribers
		case l := <-u:
			for _, s := range subscribers {
				s <- l
			}
			queue = queue[1:]
		// close() task
		case errc := <-ls.closing:
			errc <- err
			close(updates)
			for _, s := range subscribers {
				close(s)
			}
			return
		}
	}
}
