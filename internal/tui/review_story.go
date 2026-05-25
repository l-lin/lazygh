package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"

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
	return actionsPopupAction{
		id:      "review-pr-as-story",
		title:   reviewStoryActionTitle,
		icon:    actionsPopupReviewStoryIcon,
		execute: actionsPopupExecuteErr(program.executeReviewStoryAction),
	}
}

func (program *Program) executeReviewStoryAction(gui *gocui.Gui) error {
	if actualErr := program.validateStoryReviewAvailability(); actualErr != nil {
		return newActionsPopupStatusLineError(program.model.Focus(), actualErr)
	}

	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return newActionsPopupStatusLineError(program.model.Focus(), errActionsPopupActionUnavailable)
	}
	return program.dispatch(gui, MsgReviewStoryRequested{Summary: summary})
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

func (program *Program) applyPreparedStoryReview(prepared preparedStoryReview) {
	key := pullRequestDetailKey(prepared.summary.Repository, prepared.summary.Number)
	program.pullRequestDiffCache[key] = pullRequestDiffResult{data: prepared.diffData}
	if prepared.detailOK {
		program.pullRequestDetailCache[key] = pullRequestDetailResult{detail: clonePullRequestDetail(prepared.detail)}
		program.invalidatePullRequestDetailDocumentCache()
	}
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
