package tui

import (
	"strings"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func (program *Program) optimisticallyUpdatePullRequestTitle(repository string, number int, title string) {
	program.optimisticallyUpdatePullRequestFields(repository, number, func(summary *githubcli.PullRequest, detail *githubcli.PullRequestDetail) {
		trimmedTitle := strings.TrimSpace(title)
		summary.Title = trimmedTitle
		detail.Title = trimmedTitle
	})
}

func (program *Program) optimisticallyUpdatePullRequestDescription(repository string, number int, body string) {
	program.optimisticallyUpdatePullRequestFields(repository, number, func(summary *githubcli.PullRequest, detail *githubcli.PullRequestDetail) {
		trimmedBody := strings.TrimSpace(body)
		summary.Body = trimmedBody
		detail.Body = trimmedBody
		detail.BodyHTML = ""
	})
}

func (program *Program) optimisticallyUpdatePullRequestFields(repository string, number int, mutate func(*githubcli.PullRequest, *githubcli.PullRequestDetail)) {
	if program == nil || mutate == nil {
		return
	}

	key := pullRequestMutationCacheKey(repository, number)
	if key == "" {
		return
	}

	identity := githubcli.PullRequest{Repository: githubcli.Repository{NameWithOwner: strings.TrimSpace(repository)}, Number: number}
	program.mutateLoadedPullRequestSummaries(identity, func(summary *githubcli.PullRequest) {
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

func (program *Program) optimisticPullRequestDetailSeed(identity githubcli.PullRequest) githubcli.PullRequestDetail {
	seed := githubcli.PullRequestDetail{Number: identity.Number}
	if program == nil {
		return seed
	}

	key := pullRequestDetailKey(identity.Repository, identity.Number)
	if result, ok := program.pullRequestDetailCache[key]; ok && result.err == nil {
		return result.detail
	}

	if summary, ok := program.currentPullRequestSummary(); ok && samePullRequestIdentity(summary, identity) {
		return githubcli.PullRequestDetail{
			Title:   strings.TrimSpace(summary.Title),
			Number:  summary.Number,
			URL:     strings.TrimSpace(summary.URL),
			Body:    strings.TrimSpace(summary.Body),
			State:   strings.TrimSpace(summary.State),
			IsDraft: summary.IsDraft,
		}
	}

	if program.openedPullRequestSummary != nil && samePullRequestIdentity(*program.openedPullRequestSummary, identity) {
		summary := *program.openedPullRequestSummary
		return githubcli.PullRequestDetail{
			Title:   strings.TrimSpace(summary.Title),
			Number:  summary.Number,
			URL:     strings.TrimSpace(summary.URL),
			Body:    strings.TrimSpace(summary.Body),
			State:   strings.TrimSpace(summary.State),
			IsDraft: summary.IsDraft,
		}
	}

	if samePullRequestIdentity(program.reviewSession.summary, identity) {
		summary := program.reviewSession.summary
		return githubcli.PullRequestDetail{
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
