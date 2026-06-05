package tui

import (
	"errors"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/story"
)

const (
	reviewStoryActionTitle           = "Start review as story"
	storyReviewConfigureAgentMessage = "configure `story_review.agent_command` in ~/.config/lazygh/config.toml before using story review"
	storyReviewUnavailableMessage    = "story review is unavailable"
)

type reviewStoryGenerator interface {
	Generate(config story.Config, request story.Request) (story.Review, error)
}

type preparedStoryReview struct {
	summary         githubdomain.PullRequest
	detail          githubdomain.PullRequestDetail
	detailOK        bool
	diffData        reviewDiffData
	storyData       reviewStoryData
	pendingReviewID string
}

type commandReviewStoryGenerator struct {
	generator story.Generator
}

func (generator commandReviewStoryGenerator) Generate(config story.Config, request story.Request) (story.Review, error) {
	return generator.generator.Generate(config, request)
}

func (program *Program) reviewStoryAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(newActionsPopupStatusLineError(program.model.Focus(), errActionsPopupActionUnavailable))
	if actualErr := program.validateStoryReviewAvailability(); actualErr != nil {
		requested = actionsPopupErrorRequested(newActionsPopupStatusLineError(program.model.Focus(), actualErr))
	} else if summary, ok := program.currentPullRequestSummary(); ok {
		requested = MsgReviewStoryRequested{Summary: summary}
	}
	return actionsPopupAction{
		id:        "review-pr-as-story",
		title:     reviewStoryActionTitle,
		icon:      actionsPopupReviewStoryIcon,
		requested: requested,
	}
}

func (program *Program) validateStoryReviewAvailability() error {
	if !program.hasDetailQueries() || !program.hasReviewMutations() || program.storyGenerator == nil {
		return errors.New(storyReviewUnavailableMessage)
	}
	if !program.runtimeConfig.storyReviewConfig.Configured() {
		return errors.New(storyReviewConfigureAgentMessage)
	}
	return nil
}

func (program *Program) requestStoryReview(summary githubdomain.PullRequest, forceRefresh bool) []Cmd {
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if pullRequestKeyFromIdentity(repository, summary.Number) == "" {
		return nil
	}

	if !forceRefresh {
		if cached, ok := program.storyReviewForSummary(summary); ok && !storyReviewNeedsRefresh(summary, cached, ok) {
			program.clearFeedbackMessage()
			program.startStoryReviewSession(summary, cached.pendingReviewID, cached.story)
			return nil
		}
	} else {
		program.invalidatePullRequestStoryReview(repository, summary.Number)
	}

	program.startStoryReviewLoading()
	return []Cmd{storyReviewPrepareCmd{request: pullRequestStoryReviewPrepareRequest{summary: summary}}}
}

func (program *Program) applyPreparedStoryReview(prepared preparedStoryReview) {
	repository := pullRequestRepositoryName(prepared.summary.Repository)
	program.applyPullRequestDiffCacheResult(repository, prepared.summary.Number, pullRequestDiffResult{data: prepared.diffData}, pullRequestDiffCacheApplyOptions{})
	if prepared.detailOK {
		program.applyPullRequestDetailCacheResult(repository, prepared.summary.Number, pullRequestDetailResult{detail: clonePullRequestDetail(prepared.detail)}, pullRequestDetailCacheApplyOptions{invalidateDocuments: true})
	}
	program.cacheStoryReview(prepared.summary, prepared.pendingReviewID, prepared.storyData)
	program.startStoryReviewSession(prepared.summary, prepared.pendingReviewID, prepared.storyData)
}

func buildStoryReviewMetadata(summary githubdomain.PullRequest, detail githubdomain.PullRequestDetail, detailOK bool, rawDiff githubdomain.PullRequestDiff) story.Metadata {
	metadata := story.Metadata{
		Number:       summary.Number,
		Title:        strings.TrimSpace(summary.Title),
		Body:         strings.TrimSpace(summary.Body),
		URL:          strings.TrimSpace(summary.URL),
		Additions:    0,
		Deletions:    0,
		ChangedFiles: len(rawDiff.Files),
	}
	for _, file := range rawDiff.Files {
		metadata.Additions += file.Additions
		metadata.Deletions += file.Deletions
	}
	if detailOK {
		metadata.Title = preferredStoryMetadataValue(strings.TrimSpace(detail.Title), metadata.Title)
		metadata.Body = preferredStoryMetadataValue(strings.TrimSpace(detail.Body), metadata.Body)
		metadata.URL = preferredStoryMetadataValue(strings.TrimSpace(detail.URL), metadata.URL)
		metadata.Base = strings.TrimSpace(detail.BaseRefName)
		metadata.Head = strings.TrimSpace(detail.HeadRefName)
		if detail.Author != nil {
			metadata.Author = strings.TrimSpace(detail.Author.Login)
		}
		if detail.ChangedFiles > 0 {
			metadata.ChangedFiles = detail.ChangedFiles
		}
		if detail.Additions > 0 {
			metadata.Additions = detail.Additions
		}
		if detail.Deletions > 0 {
			metadata.Deletions = detail.Deletions
		}
	}
	return metadata
}

func buildStoryReviewDiffItems(files []githubdomain.PullRequestDiffFile) []story.DiffItem {
	items := make([]story.DiffItem, 0, len(files))
	for _, file := range files {
		trimmedPath := strings.TrimSpace(file.Path)
		if trimmedPath == "" {
			continue
		}
		items = append(items, story.DiffItem{File: trimmedPath})
	}
	return items
}

func preferredStoryMetadataValue(primary string, fallback string) string {
	trimmedPrimary := strings.TrimSpace(primary)
	if trimmedPrimary != "" {
		return trimmedPrimary
	}
	return strings.TrimSpace(fallback)
}
