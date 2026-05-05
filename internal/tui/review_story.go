package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/story"
)

const (
	actionsPopupReviewStoryIcon      = ""
	reviewStoryActionTitle           = "Review PR as story"
	storyReviewConfigureAgentMessage = "configure `story_review.agent_command` in ~/.config/lazygh/config.toml before using story review"
	storyReviewUnavailableMessage    = "story review is unavailable"
	storyReviewGeneratingFeedback    = "Generating story review chapters..."
)

type reviewStoryGenerator interface {
	Generate(config story.Config, request story.Request) (story.Review, error)
}

type commandReviewStoryGenerator struct {
	generator story.Generator
}

func (generator commandReviewStoryGenerator) Generate(config story.Config, request story.Request) (story.Review, error) {
	return generator.generator.Generate(config, request)
}

func (program *Program) reviewStoryAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "review-pr-as-story",
		title:    reviewStoryActionTitle,
		icon:     actionsPopupReviewStoryIcon,
		keywords: []string{"review", "story", "chapter", "ai", "narrative"},
		execute:  program.executeReviewStoryAction,
	}
}

func (program *Program) executeReviewStoryAction(gui *gocui.Gui) actionsPopupActionResult {
	if program.githubLoader == nil || program.storyGenerator == nil {
		return program.storyReviewStatusLineErrorResult(errors.New(storyReviewUnavailableMessage))
	}
	if !program.storyReviewConfig.Configured() {
		return program.storyReviewStatusLineErrorResult(errors.New(storyReviewConfigureAgentMessage))
	}

	summary, ok := program.model.SelectedPullRequestSummary()
	if !ok {
		return program.storyReviewStatusLineErrorResult(errActionsPopupActionUnavailable)
	}

	program.setFeedback(program.model.Focus(), storyReviewGeneratingFeedback)
	program.asyncRunner.Go(func() {
		program.loadStoryReview(gui, summary)
	})
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) loadStoryReview(gui *gocui.Gui, summary githubcli.PullRequest) {
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
			program.setFeedback(program.model.Focus(), "missing pull request identity")
			return program.refreshViewsIfGUI(gui)
		})
		return
	}

	detail, detailOK := program.storyReviewDetail(summary)
	rawDiff, actualErr := program.githubLoader.GetPullRequestDiff(repository, summary.Number)
	if actualErr != nil {
		program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
			program.setFeedback(program.model.Focus(), strings.TrimSpace(actualErr.Error()))
			return program.refreshViewsIfGUI(gui)
		})
		return
	}

	generatedStory, actualErr := program.storyGenerator.Generate(program.storyReviewConfig, story.Request{
		Metadata:  buildStoryReviewMetadata(summary, detail, detailOK, rawDiff),
		DiffItems: buildStoryReviewDiffItems(rawDiff.Files),
		DiffText:  rawDiff.UnifiedDiff,
	})
	if actualErr != nil {
		program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
			program.setFeedback(program.model.Focus(), strings.TrimSpace(actualErr.Error()))
			return program.refreshViewsIfGUI(gui)
		})
		return
	}

	pendingReviewID, actualErr := program.githubLoader.StartPendingPullRequestReview(repository, summary.Number)
	if actualErr != nil {
		program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
			program.setFeedback(program.model.Focus(), strings.TrimSpace(actualErr.Error()))
			return program.refreshViewsIfGUI(gui)
		})
		return
	}

	diffData := buildReviewDiffData(rawDiff)
	storyData := buildReviewStoryData(generatedStory, diffData.Files)
	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		program.feedbackMessage = ""
		key := pullRequestDetailKey(summary.Repository, summary.Number)
		program.pullRequestDiffCache[key] = pullRequestDiffResult{data: diffData}
		if detailOK {
			program.pullRequestDetailCache[key] = pullRequestDetailResult{detail: detail}
		}
		program.startStoryReviewSession(summary, pendingReviewID, storyData)
		if gui == nil {
			return nil
		}
		return program.layout(gui)
	})
}

func (program *Program) storyReviewDetail(summary githubcli.PullRequest) (githubcli.PullRequestDetail, bool) {
	if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
		return result.detail, true
	}
	if program.githubLoader == nil {
		return githubcli.PullRequestDetail{}, false
	}
	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return githubcli.PullRequestDetail{}, false
	}
	detail, actualErr := program.githubLoader.GetPullRequestDetail(repository, summary.Number)
	if actualErr != nil {
		return githubcli.PullRequestDetail{}, false
	}
	return detail, true
}

func buildStoryReviewMetadata(summary githubcli.PullRequest, detail githubcli.PullRequestDetail, detailOK bool, rawDiff githubcli.PullRequestDiff) story.Metadata {
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

func buildStoryReviewDiffItems(files []githubcli.PullRequestDiffFile) []story.DiffItem {
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
