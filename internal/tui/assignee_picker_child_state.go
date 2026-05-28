package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (state assigneePickerState) withSelectionToggled(candidate githubdomain.PullRequestAuthor) assigneePickerState {
	normalizedCandidate := normalizedAssigneePickerCandidate(candidate)
	if normalizedCandidate.Login == "" {
		return state
	}

	state.selectedLogins = copyAssigneePickerSelections(state.selectedLogins)
	if state.selectedLogins[normalizedCandidate.Login] {
		delete(state.selectedLogins, normalizedCandidate.Login)
	} else {
		state.selectedLogins[normalizedCandidate.Login] = true
	}
	return state.rememberingCandidates([]githubdomain.PullRequestAuthor{normalizedCandidate})
}

func (state assigneePickerState) withSearchReset(query string) (assigneePickerState, int) {
	state.searchRequestID++
	state.searchQuery = strings.TrimSpace(query)
	state.searchResults = nil
	state.searchLoading = false
	state.searchCommand = ""
	return state, state.searchRequestID
}

func (state assigneePickerState) withSearchLoadingStarted(query string) assigneePickerState {
	state.searchLoading = true
	state.searchCommand = formatAssigneeSearchCommand(state.target.repository, query)
	return state
}

func (state assigneePickerState) withSearchLoaded(query string, results []githubdomain.PullRequestAuthor) assigneePickerState {
	state.searchLoading = false
	state.searchCommand = ""
	state.searchQuery = strings.TrimSpace(query)
	state = state.rememberingCandidates(results)
	state.searchResults = append([]githubdomain.PullRequestAuthor(nil), results...)
	return state
}

func (state assigneePickerState) rememberingCandidates(candidates []githubdomain.PullRequestAuthor) assigneePickerState {
	if state.knownCandidates == nil {
		state.knownCandidates = map[string]githubdomain.PullRequestAuthor{}
	} else {
		state.knownCandidates = copyAssigneePickerCandidates(state.knownCandidates)
	}

	for _, candidate := range candidates {
		normalizedCandidate := normalizedAssigneePickerCandidate(candidate)
		if normalizedCandidate.Login == "" {
			continue
		}
		existing := state.knownCandidates[normalizedCandidate.Login]
		if normalizedCandidate.Name == "" {
			normalizedCandidate.Name = existing.Name
		}
		if normalizedCandidate.Name == "" && normalizedCandidate.Login == state.viewerLogin {
			normalizedCandidate.Name = state.viewerName
		}
		state.knownCandidates[normalizedCandidate.Login] = normalizedCandidate
		if normalizedCandidate.Login == state.viewerLogin && normalizedCandidate.Name != "" {
			state.viewerName = normalizedCandidate.Name
		}
	}
	return state
}

func copyAssigneePickerSelections(source map[string]bool) map[string]bool {
	copied := make(map[string]bool, len(source))
	for login, selected := range source {
		copied[login] = selected
	}
	return copied
}

func copyAssigneePickerCandidates(source map[string]githubdomain.PullRequestAuthor) map[string]githubdomain.PullRequestAuthor {
	copied := make(map[string]githubdomain.PullRequestAuthor, len(source))
	for login, candidate := range source {
		copied[login] = candidate
	}
	return copied
}
