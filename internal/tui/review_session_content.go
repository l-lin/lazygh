package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

func (program *Program) reviewSessionMetadataContent() string {
	summary := program.reviewSession.summary
	detail := githubcli.PullRequestDetail{Title: summary.Title, Number: summary.Number, Body: summary.Body, State: summary.State, UpdatedAt: summary.UpdatedAt}
	bodyLines := []string{fmt.Sprintf("Pending review: %s", valueOrDash(program.reviewSession.pendingReviewID))}

	if result, ok := program.pullRequestDetailForSummary(summary); ok {
		if result.err != nil {
			bodyLines = append(bodyLines, "", "Could not load richer pull request metadata.", strings.TrimSpace(result.err.Error()))
		} else {
			detail = result.detail
		}
	} else {
		bodyLines = append(bodyLines, "", "Loading richer pull request metadata...")
	}

	if result, ok := program.reviewSessionDiffResult(); ok {
		if result.err != nil {
			bodyLines = append(bodyLines, "", "Could not load changed files.", strings.TrimSpace(result.err.Error()))
		} else {
			bodyLines = append(bodyLines, renderReviewSessionMetadataStats(result.data.Stats))
		}
	} else {
		bodyLines = append(bodyLines, "", "Loading changed files...")
	}

	return renderPullRequestDetailContent(renderPullRequestDetailHeader(summary, detail), strings.Join(bodyLines, "\n"))
}

func renderReviewSessionMetadataStats(stats reviewDiffStats) string {
	return strings.Join([]string{
		fmt.Sprintf("Changed files: %d", stats.ChangedFiles),
		styleText(fmt.Sprintf("+%d", stats.Additions), foregroundColorEscape(theme.DiffAdditionHex)),
		styleText(fmt.Sprintf("-%d", stats.Deletions), foregroundColorEscape(theme.DiffDeletionHex)),
	}, "  ")
}

func (program *Program) reviewSessionDetailContent() string {
	if !program.reviewSession.active {
		return ""
	}
	if program.reviewSessionShowsDescription() {
		return program.reviewSessionDescriptionContent()
	}
	if program.reviewSessionShowsStoryChapter() {
		return program.reviewSessionStoryChapterContent()
	}

	result, ok := program.reviewSessionDiffResult()
	if !ok {
		return program.reviewSessionLoadingDetail()
	}
	if result.err != nil {
		return program.reviewSessionDiffErrorDetail(result.err)
	}
	selectedFile, ok := program.selectedReviewSessionDiffFile()
	if !ok {
		return program.reviewSessionNoDiffDetail()
	}
	return renderReviewDiffFileWithCollapsedThreadsForViewer(selectedFile, program.markdownRenderer, program.detailWrapWidth, program.reviewSession.collapsedThreadIDs, program.currentConnectedUserLogin())
}

func (program *Program) reviewSessionDescriptionContent() string {
	summary := program.reviewSession.summary
	detail := githubcli.PullRequestDetail{Title: summary.Title, Number: summary.Number, Body: summary.Body, State: summary.State, UpdatedAt: summary.UpdatedAt}
	if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
		detail = result.detail
	}

	return renderPullRequestDetailContentWithSeparator(
		renderPullRequestDetailHeader(summary, detail),
		renderPullRequestDescription(summary, detail, program.markdownRenderer, program.detailWrapWidth),
		program.detailWrapWidth,
	)
}

func (program *Program) reviewSessionStoryChapterContent() string {
	chapter, ok := program.selectedReviewSessionStoryChapter()
	if !ok {
		return program.reviewSessionNoDiffDetail()
	}

	sections := []string{"# " + strings.TrimSpace(chapter.Title)}
	if strings.TrimSpace(chapter.Narrative) != "" {
		sections = append(sections, strings.TrimSpace(chapter.Narrative))
	}
	if len(chapter.Files) > 0 {
		sections = append(sections, "## Files")
		for _, file := range chapter.Files {
			sections = append(sections, "- "+strings.TrimSpace(file))
		}
	}
	return renderMarkdownWithFallback(strings.Join(sections, "\n\n"), program.markdownRenderer, program.detailWrapWidth, "No chapter narrative is available.")
}

func (program *Program) reviewSessionShowsDescription() bool {
	if !program.reviewSession.active {
		return false
	}
	return program.model.currentSideFocus() == FocusUserView
}

func (program *Program) reviewSessionShowsStoryChapter() bool {
	if !program.reviewSession.active || program.reviewSession.mode != reviewSessionModeStory {
		return false
	}
	_, ok := program.selectedReviewSessionStoryChapter()
	return ok
}

func (program *Program) reviewSessionLoadingDetail() string {
	summary := program.reviewSession.summary
	repository := pullRequestRepositoryName(summary.Repository)
	lines := []string{
		fmt.Sprintf("Pending review %s is open for %s#%d.", valueOrDash(program.reviewSession.pendingReviewID), repository, summary.Number),
		"",
		"Loading pull request diff...",
		fmt.Sprintf("Running `gh api repos/%s/pulls/%d -H 'Accept: application/vnd.github.v3.diff'`.", repository, summary.Number),
	}
	return strings.Join(lines, "\n")
}

func (program *Program) reviewSessionDiffErrorDetail(err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		message = "Unknown error. GitHub misplaced the diff again."
	}

	lines := []string{program.reviewSessionLoadingDetail(), "", message}
	return strings.Join(lines, "\n")
}

func (program *Program) reviewSessionNoDiffDetail() string {
	summary := program.reviewSession.summary
	repository := pullRequestRepositoryName(summary.Repository)
	lines := []string{
		fmt.Sprintf("Pending review %s is open for %s#%d.", valueOrDash(program.reviewSession.pendingReviewID), repository, summary.Number),
		"",
		"No changed files are available for this review.",
	}
	return strings.Join(lines, "\n")
}

func (program *Program) reviewSessionChangedFileCount() int {
	if !program.reviewSession.active {
		return 0
	}
	if result, ok := program.reviewSessionDiffResult(); ok && result.err == nil {
		return result.data.Stats.ChangedFiles
	}
	if result, ok := program.pullRequestDetailForSummary(program.reviewSession.summary); ok && result.err == nil {
		return result.detail.ChangedFiles
	}
	return 0
}

func (program *Program) reviewSessionDetailIdentity() string {
	if !program.reviewSession.active {
		return ""
	}
	if program.reviewSessionShowsDescription() {
		return fmt.Sprintf(
			"review:%s:%d:%s:description",
			pullRequestRepositoryName(program.reviewSession.summary.Repository),
			program.reviewSession.summary.Number,
			program.reviewSession.pendingReviewID,
		)
	}
	if chapter, ok := program.selectedReviewSessionStoryChapter(); ok {
		return fmt.Sprintf(
			"review:%s:%d:%s:chapter:%s",
			pullRequestRepositoryName(program.reviewSession.summary.Repository),
			program.reviewSession.summary.Number,
			program.reviewSession.pendingReviewID,
			chapter.ID,
		)
	}
	selectedFilePath := fmt.Sprintf("row:%d", program.reviewSession.selectedFileTreeRow)
	if selectedFile, ok := program.selectedReviewSessionDiffFile(); ok {
		selectedFilePath = selectedFile.Path
	}
	return fmt.Sprintf(
		"review:%s:%d:%s:file:%s",
		pullRequestRepositoryName(program.reviewSession.summary.Repository),
		program.reviewSession.summary.Number,
		program.reviewSession.pendingReviewID,
		selectedFilePath,
	)
}

func (program *Program) reviewSessionDiffResult() (pullRequestDiffResult, bool) {
	if !program.reviewSession.active {
		return pullRequestDiffResult{}, false
	}
	return program.pullRequestDiffForSummary(program.reviewSession.summary)
}

func renderReadOnlyTextView(view *gocui.View, text string) {
	if view == nil {
		return
	}

	view.Clear()
	view.SetOrigin(0, 0)
	view.SetCursor(0, 0)
	fmt.Fprint(view, strings.TrimSpace(text))
}
