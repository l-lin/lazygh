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
		execute: program.executeReviewStoryAction,
	}
}

func (program *Program) executeReviewStoryAction(gui *gocui.Gui) actionsPopupActionResult {
	if actualErr := program.validateStoryReviewAvailability(); actualErr != nil {
		return program.storyReviewStatusLineErrorResult(actualErr)
	}

	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return program.storyReviewStatusLineErrorResult(errActionsPopupActionUnavailable)
	}

	program.feedbackMessage = ""
	program.storyReviewLoading = true
	program.asyncRunner.Go(func() {
		program.loadStoryReview(gui, summary)
	})
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) loadStoryReview(gui *gocui.Gui, summary githubdomain.PullRequest) {
	prepared, actualErr := program.prepareStoryReview(summary)
	if actualErr != nil {
		program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
			program.storyReviewLoading = false
			program.setFeedback(program.model.Focus(), strings.TrimSpace(actualErr.Error()))
			return program.refreshViewsIfGUI(gui)
		})
		return
	}

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		program.storyReviewLoading = false
		program.feedbackMessage = ""
		program.applyPreparedStoryReview(prepared)
		if gui == nil {
			return nil
		}
		return program.layout(gui)
	})
}

func (program *Program) validateStoryReviewAvailability() error {
	if !program.hasDetailQueries() || !program.hasReviewMutations() || program.storyGenerator == nil {
		return errors.New(storyReviewUnavailableMessage)
	}
	if !program.storyReviewConfig.Configured() {
		return errors.New(storyReviewConfigureAgentMessage)
	}
	return nil
}

func (program *Program) prepareStoryReview(summary githubdomain.PullRequest) (preparedStoryReview, error) {
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return preparedStoryReview{}, errors.New("missing pull request identity")
	}

	detail, detailOK := program.storyReviewDetail(summary)
	rawDiff, actualErr := program.detailQueries.GetPullRequestDiff(repository, summary.Number)
	if actualErr != nil {
		return preparedStoryReview{}, actualErr
	}

	generatedStory, actualErr := program.storyGenerator.Generate(program.storyReviewConfig, story.Request{
		Metadata:  buildStoryReviewMetadata(summary, detail, detailOK, rawDiff),
		DiffItems: buildStoryReviewDiffItems(rawDiff.Files),
		DiffText:  rawDiff.UnifiedDiff,
	})
	if actualErr != nil {
		return preparedStoryReview{}, actualErr
	}

	pendingReviewID, actualErr := program.reviewMutations.StartPendingPullRequestReview(repository, summary.Number)
	if actualErr != nil {
		return preparedStoryReview{}, actualErr
	}

	diffData := buildReviewDiffData(rawDiff)
	return preparedStoryReview{
		summary:         summary,
		detail:          detail,
		detailOK:        detailOK,
		diffData:        diffData,
		storyData:       buildReviewStoryData(generatedStory, diffData.Files),
		pendingReviewID: pendingReviewID,
	}, nil
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

func (program *Program) storyReviewDetail(summary githubdomain.PullRequest) (githubdomain.PullRequestDetail, bool) {
	if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
		return result.detail, true
	}
	if !program.hasDetailQueries() {
		return githubdomain.PullRequestDetail{}, false
	}
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return githubdomain.PullRequestDetail{}, false
	}
	detail, actualErr := program.detailQueries.GetPullRequestDetail(repository, summary.Number)
	if actualErr != nil {
		return githubdomain.PullRequestDetail{}, false
	}
	return detail, true
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

func (program *Program) storyReviewStatusLineErrorResult(actualErr error) actionsPopupActionResult {
	return actionsPopupActionResult{
		err:             actualErr,
		feedbackMessage: strings.TrimSpace(actualErr.Error()),
		feedbackTarget:  program.model.Focus(),
	}
}
