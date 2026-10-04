package githubcli

import (
	"strconv"
	"sync/atomic"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

var commandActivitySequence atomic.Uint64

type commandActivityObserver interface {
	ObserveActivity(githubdomain.CommandActivity)
}

type activityObservingRunner struct {
	Runner
	observe func(githubdomain.CommandActivity)
}

// NewActivityObservingRunner reports command lifecycle events. The callback must be concurrency-safe.
func NewActivityObservingRunner(runner Runner, observe func(githubdomain.CommandActivity)) Runner {
	if runner == nil {
		runner = execRunner{}
	}
	if observe == nil {
		return runner
	}
	return activityObservingRunner{Runner: runner, observe: observe}
}

func (runner activityObservingRunner) ObserveActivity(activity githubdomain.CommandActivity) {
	runner.observe(activity)
}

func nextCommandActivityID() string {
	return "gh-" + strconv.FormatUint(commandActivitySequence.Add(1), 10)
}
