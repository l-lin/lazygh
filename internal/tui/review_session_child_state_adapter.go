package tui

func (program *Program) setReviewSessionSelectedFileTreeRow(row int) {
	if program == nil {
		return
	}
	program.navigationState.reviewSession = program.navigationState.reviewSession.withSelectedFileTreeRow(row)
}

func (program *Program) setReviewSessionThreadCollapsed(threadID string, collapsed bool) {
	if program == nil {
		return
	}
	program.navigationState.reviewSession = program.navigationState.reviewSession.withThreadCollapsed(threadID, collapsed)
}

func (program *Program) setAllReviewSessionThreadsCollapsed(threads []reviewDiffThread, collapsed bool) bool {
	if program == nil {
		return false
	}
	updatedReviewSession, changed := program.navigationState.reviewSession.withAllThreadsCollapsed(threads, collapsed)
	if !changed {
		return false
	}
	program.navigationState.reviewSession = updatedReviewSession
	return true
}
