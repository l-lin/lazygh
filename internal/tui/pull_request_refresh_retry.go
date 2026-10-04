package tui

import "time"

type pullRequestRefreshRetry struct {
	failureCount uint8
	retryAfter   time.Time
}

func (retry pullRequestRefreshRetry) afterFailure(now time.Time) pullRequestRefreshRetry {
	retry.failureCount = min(retry.failureCount+1, 5)
	retry.retryAfter = now.Add(min(30*time.Second<<(retry.failureCount-1), 5*time.Minute))
	return retry
}

func (retry pullRequestRefreshRetry) ready(now time.Time) bool {
	return !now.Before(retry.retryAfter)
}
