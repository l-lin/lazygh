package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func (program *Program) reviewSessionMetadataContent() string {
	reference := pullRequestReference(program.reviewSession.summary, githubcli.PullRequestDetail{})
	return valueOrDash(strings.TrimSpace(reference))
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
	if result, ok := program.pullRequestDetailForSummary(summary); ok {
		if result.err != nil {
			return renderPullRequestDetailError(summary, result.err)
		}

		header := renderPullRequestBrowserHeader(summary, result.detail)
		overview := program.renderCurrentPullRequestOverview(summary, result.detail, program.detailWrapWidth)
		content := renderPullRequestDescription(summary, result.detail, program.markdownRenderer, program.detailWrapWidth)
		return renderPullRequestBrowserDetailContent(header, overview, content, program.detailWrapWidth)
	}

	return renderPullRequestDetailLoading(summary, program.loadingSpinnerFrame())
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
	return renderMarkdownWithFallback(strings.Join(sections, "\n\n"), program.markdownRenderer, program.detailWrapWidth, "No chapter narrative is available.")
}

func (program *Program) reviewSessionDescriptionSummaryAndDetail() (githubcli.PullRequest, githubcli.PullRequestDetail, bool) {
	if !program.reviewSessionShowsDescription() {
		return githubcli.PullRequest{}, githubcli.PullRequestDetail{}, false
	}

	summary := program.reviewSession.summary
	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return summary, githubcli.PullRequestDetail{}, false
	}
	return summary, result.detail, true
}

func (program *Program) reviewSessionShowsDescription() bool {
	if !program.reviewSession.active {
		return false
	}
	return program.mainViewResolver().ContentKind == MainContentKindReviewDescription
}

func (program *Program) reviewSessionShowsStoryChapter() bool {
	if !program.reviewSession.active {
		return false
	}
	return program.mainViewResolver().ContentKind == MainContentKindStoryChapter
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
