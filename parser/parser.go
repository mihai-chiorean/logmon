package parser

import (
	"reflect"

	"github.com/Songmu/axslogparser"
)

// Log is a generic log interface that is used by monitor and reporters
type Log interface{}

// LogParser is an interface that describes the behaviour expected to be exposed
// by a parser used in the system
type LogParser interface {
	Parse(string) (Log, error)
	LogType() string
}

// AccessLogParser is an implementation of the LogParser that uses axslogparser
// to process access log lines
type AccessLogParser struct{}

func (a AccessLogParser) Parse(line string) (Log, error) {
	return axslogparser.Parse(line)
}

func (a AccessLogParser) LogType() string {
	l := axslogparser.Log{}
	return reflect.TypeOf(l).Name()
}
