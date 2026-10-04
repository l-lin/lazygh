package tui

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

var refreshActivitySequence atomic.Uint64

// AI: measure refresh ownership at the job boundary; per-command attribution would require propagating context through provider ports.
func runObservedCommandAsync(runAsync func(func()), report func(Msg), now func() time.Time, command, owner string, run func() error) {
	queuedAt := now()
	id := "refresh-" + strconv.FormatUint(refreshActivitySequence.Add(1), 10)
	runWorkflowCommandAsync(runAsync, func() {
		startedAt := now()
		activity := githubdomain.CommandActivity{
			ID: id, Command: command, StartedAt: startedAt,
			QueueWait: startedAt.Sub(queuedAt), Owner: owner, Outcome: githubdomain.CommandRunning,
		}
		report(MsgCommandActivity{Activity: activity})
		err := run()
		activity.FinishedAt = now()
		activity.Outcome = githubdomain.CommandSucceeded
		if err != nil {
			activity.Outcome = githubdomain.CommandFailed
		}
		report(MsgCommandActivity{Activity: activity})
	})
}

func (runtime workflowShellRuntime) runObservedPullRequestAsync(key, command, owner string, run func() error) {
	runObservedCommandAsync(func(task func()) {
		runtime.runPullRequestAsync(key, task)
	}, runtime.dispatchAsyncMessage, time.Now, command, owner, run)
}

// Snapshot ownership on the UI thread, before selection or manual-refresh state changes.
func (program *Program) refreshCommandOwner(batchID uint64, manual bool, key string) string {
	if batchID != 0 {
		return fmt.Sprintf("scheduled batch=%d", batchID)
	}
	if manual {
		return "manual"
	}
	if key == program.activePullRequestLoadKey() {
		return "interactive"
	}
	return "background"
}
