package tui

import "sync"

const (
	pullRequestWorkerLimit           = 4
	pullRequestBackgroundWorkerLimit = 2
)

type pullRequestAsyncTask struct {
	key string
	run func()
}

type pullRequestAsyncRunner struct {
	asyncRunner
	mu                sync.Mutex
	activeKey         string
	pending           []pullRequestAsyncTask
	running           int
	backgroundRunning int
}

func newPullRequestAsyncRunner(backend asyncRunner) *pullRequestAsyncRunner {
	return &pullRequestAsyncRunner{asyncRunner: backend}
}

func (runner *pullRequestAsyncRunner) goPullRequest(key string, run func()) {
	if run == nil {
		return
	}
	runner.mu.Lock()
	runner.pending = append(runner.pending, pullRequestAsyncTask{key: key, run: run})
	runs := runner.takeReadyLocked()
	runner.mu.Unlock()
	runner.dispatch(runs)
}

func (runner *pullRequestAsyncRunner) setActivePullRequest(key string) {
	runner.mu.Lock()
	runner.activeKey = key
	runs := runner.takeReadyLocked()
	runner.mu.Unlock()
	runner.dispatch(runs)
}

func (runner *pullRequestAsyncRunner) takeReadyLocked() []func() {
	var runs []func()
	for runner.running < pullRequestWorkerLimit && len(runner.pending) > 0 {
		selectedIndex := -1
		for index, task := range runner.pending {
			if task.key != "" && task.key == runner.activeKey {
				selectedIndex = index
				break
			}
		}
		background := selectedIndex == -1
		if background {
			// Reserve two slots so a slow background backlog cannot block visible description and diff loads.
			if runner.backgroundRunning >= pullRequestBackgroundWorkerLimit {
				break
			}
			selectedIndex = 0
		}
		task := runner.pending[selectedIndex]
		copy(runner.pending[selectedIndex:], runner.pending[selectedIndex+1:])
		runner.pending[len(runner.pending)-1] = pullRequestAsyncTask{}
		runner.pending = runner.pending[:len(runner.pending)-1]
		runner.running++
		if background {
			runner.backgroundRunning++
		}
		runs = append(runs, func() {
			defer runner.completed(background)
			task.run()
		})
	}
	return runs
}

func (runner *pullRequestAsyncRunner) completed(background bool) {
	runner.mu.Lock()
	runner.running--
	if background {
		runner.backgroundRunning--
	}
	runs := runner.takeReadyLocked()
	runner.mu.Unlock()
	runner.dispatch(runs)
}

func (runner *pullRequestAsyncRunner) dispatch(runs []func()) {
	for _, run := range runs {
		runner.asyncRunner.Go(run)
	}
}

func (program *Program) activePullRequestLoadKey() string {
	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return ""
	}
	return pullRequestDetailKey(summary.Repository, summary.Number)
}

func (program *Program) syncPullRequestAsyncPriority() {
	if runner, ok := program.asyncRunner.(*pullRequestAsyncRunner); ok {
		runner.setActivePullRequest(program.activePullRequestLoadKey())
	}
}

func (program *Program) runPullRequestAsync(key string, run func()) {
	if runner, ok := program.asyncRunner.(*pullRequestAsyncRunner); ok {
		program.syncPullRequestAsyncPriority()
		runner.goPullRequest(key, run)
		return
	}
	program.runAsync(run)
}
