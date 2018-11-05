package model

// LogMessage is a struct that describes a line in the log file
type LogMessage struct {
	IsError    bool
	RawMessage string
	Error      error
}
