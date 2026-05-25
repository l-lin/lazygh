package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) optimisticallyUpdatePullRequestTitle(repository string, number int, title string) {
	program.optimisticallyUpdatePullRequestFields(repository, number, func(summary *githubdomain.PullRequest, detail *githubdomain.PullRequestDetail) {
		trimmedTitle := strings.TrimSpace(title)
		summary.Title = trimmedTitle
		detail.Title = trimmedTitle
	})
}

func (program *Program) optimisticallyUpdatePullRequestDescription(repository string, number int, body string) {
	program.optimisticallyUpdatePullRequestFields(repository, number, func(summary *githubdomain.PullRequest, detail *githubdomain.PullRequestDetail) {
		trimmedBody := strings.TrimSpace(body)
		summary.Body = trimmedBody
		detail.Body = trimmedBody
		detail.BodyHTML = ""
	})
}

func (program *Program) optimisticallyUpdatePullRequestFields(repository string, number int, mutate func(*githubdomain.PullRequest, *githubdomain.PullRequestDetail)) {
	if program == nil || mutate == nil {
		return
	}

	key := pullRequestMutationCacheKey(repository, number)
	if key == "" {
		return
	}

	identity := githubdomain.PullRequest{Repository: githubdomain.Repository{NameWithOwner: strings.TrimSpace(repository)}, Number: number}
	program.mutateLoadedPullRequestSummaries(identity, func(summary *githubdomain.PullRequest) {
		if summary == nil {
			return
		}
		seedDetail := program.optimisticPullRequestDetailSeed(identity)
		mutate(summary, &seedDetail)
	})

	result, ok := program.pullRequestDetailCache[key]
	if !ok || result.err != nil {
		result = pullRequestDetailResult{detail: program.optimisticPullRequestDetailSeed(identity)}
	}
	mutate(&identity, &result.detail)
	result.sourceUpdatedAt = ""
	result.needsRefresh = true
	program.pullRequestDetailCache[key] = result
	program.invalidatePullRequestDetailDocumentCache()
	program.invalidatePersistentPullRequest(repository, number)
}

func (program *Program) optimisticPullRequestDetailSeed(identity githubdomain.PullRequest) githubdomain.PullRequestDetail {
	seed := githubdomain.PullRequestDetail{Number: identity.Number}
	if program == nil {
		return seed
	}

	key := pullRequestDetailKey(identity.Repository, identity.Number)
	if result, ok := program.pullRequestDetailCache[key]; ok && result.err == nil {
		return result.detail
	}

	if summary, ok := program.currentPullRequestSummary(); ok && samePullRequestIdentity(summary, identity) {
		return githubdomain.PullRequestDetail{
			Title:   strings.TrimSpace(summary.Title),
			Number:  summary.Number,
			URL:     strings.TrimSpace(summary.URL),
			Body:    strings.TrimSpace(summary.Body),
			State:   strings.TrimSpace(summary.State),
			IsDraft: summary.IsDraft,
		}
	}

	if program.navigationState.openedPullRequestSummary != nil && samePullRequestIdentity(*program.navigationState.openedPullRequestSummary, identity) {
		summary := *program.navigationState.openedPullRequestSummary
		return githubdomain.PullRequestDetail{
			Title:   strings.TrimSpace(summary.Title),
			Number:  summary.Number,
			URL:     strings.TrimSpace(summary.URL),
			Body:    strings.TrimSpace(summary.Body),
			State:   strings.TrimSpace(summary.State),
			IsDraft: summary.IsDraft,
		}
	}

	if samePullRequestIdentity(program.navigationState.reviewSession.summary, identity) {
		summary := program.navigationState.reviewSession.summary
		return githubdomain.PullRequestDetail{
			Title:   strings.TrimSpace(summary.Title),
			Number:  summary.Number,
			URL:     strings.TrimSpace(summary.URL),
			Body:    strings.TrimSpace(summary.Body),
			State:   strings.TrimSpace(summary.State),
			IsDraft: summary.IsDraft,
		}
	}

	return seed
}
