package githubcli

import (
	"fmt"
	"strings"
	"sync"
)

type pullRequestRefresh struct {
	users       int
	threadsOnce sync.Once
	threads     []PullRequestReviewThread
	threadsErr  error
}

// AcquirePullRequestRefresh reserves shared results before work enters the read queue.
// Results expire after the last queued or running read releases its reservation.
func (client *PullRequestDetailService) AcquirePullRequestRefresh(repository string, number int) func() {
	key := fmt.Sprintf("%s#%d", strings.TrimSpace(repository), number)
	client.refreshMutex.Lock()
	if client.refreshes == nil {
		client.refreshes = make(map[string]*pullRequestRefresh)
	}
	scope := client.refreshes[key]
	if scope == nil {
		scope = &pullRequestRefresh{}
		client.refreshes[key] = scope
	}
	scope.users++
	client.refreshMutex.Unlock()
	var releaseOnce sync.Once
	return func() {
		releaseOnce.Do(func() {
			client.refreshMutex.Lock()
			defer client.refreshMutex.Unlock()
			scope.users--
			if scope.users == 0 {
				delete(client.refreshes, key)
			}
		})
	}
}

func (client *PullRequestDetailService) pullRequestRefresh(repository string, number int) *pullRequestRefresh {
	client.refreshMutex.Lock()
	defer client.refreshMutex.Unlock()
	return client.refreshes[fmt.Sprintf("%s#%d", strings.TrimSpace(repository), number)]
}

func (scope *pullRequestRefresh) loadThreads(load func() ([]PullRequestReviewThread, error)) ([]PullRequestReviewThread, error) {
	scope.threadsOnce.Do(func() { scope.threads, scope.threadsErr = load() })
	if scope.threadsErr != nil {
		return nil, scope.threadsErr
	}
	if scope.threads == nil {
		return nil, nil
	}
	threads := make([]PullRequestReviewThread, len(scope.threads))
	for index, thread := range scope.threads {
		threads[index] = thread.normalized()
	}
	return threads, nil
}

func (adapter *PullRequestDetailAdapter) AcquirePullRequestRefresh(repository string, number int) func() {
	return adapter.service.AcquirePullRequestRefresh(repository, number)
}
