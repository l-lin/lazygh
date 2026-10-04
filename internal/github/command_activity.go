package github

import "time"

type CommandOutcome uint8

const (
	CommandRunning CommandOutcome = iota
	CommandSucceeded
	CommandFailed
)

// CommandActivity contains display-safe session telemetry, never payloads or error text.
type CommandActivity struct {
	ID         string
	Command    string
	StartedAt  time.Time
	FinishedAt time.Time
	Outcome    CommandOutcome
	QueueWait  time.Duration
	Owner      string
}
