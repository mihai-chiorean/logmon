package parser

import (
	"time"

	"github.com/Songmu/axslogparser"
	"github.com/pkg/errors"
)

// Log is a generic log interface that is used by monitor and reporters
type Log interface {
	Timestamp() time.Time
	Resource() string
}

// LogParser is an interface that describes the behaviour expected to be exposed
// by a parser used in the system
type LogParser interface {
	Parse(string) (Log, error)
}

type accessLog struct {
	*axslogparser.Log
}

func (l *accessLog) Timestamp() time.Time {
	return l.Time
}

func (l *accessLog) Resource() string {
	return l.RequestURI
}

// AccessLogParser is an implementation of the LogParser that uses axslogparser
// to process access log lines
type AccessLogParser struct{}

// NewAccessLogParser is the factory function for an access log parser
func NewAccessLogParser() *AccessLogParser {
	l := AccessLogParser{}
	return &l
}

func (a AccessLogParser) Parse(line string) (Log, error) {
	l, err := axslogparser.Parse(line)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse log line")
	}
	return &accessLog{
		l,
	}, nil
}
