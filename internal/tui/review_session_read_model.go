package tui

import (
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type reviewSessionReadModel struct {
	active                        bool
	mode                          reviewSessionMode
	mainContentKind               MainContentKind
	summary                       githubdomain.PullRequest
	pendingReviewID               string
	selectedFileTreeRow           int
	collapsedTreeRowIDs           map[string]bool
	collapsedThreadIDs            map[string]bool
	story                         reviewStoryData
	detailWrapWidth               int
	wordWrapEnabled               bool
	markdownRenderer              MarkdownRenderer
	connectedUserLogin            string
	loadingSpinner                string
	descriptionResult             pullRequestDetailResult
	descriptionResultKnown        bool
	browserCollapsedSectionStates map[string]bool
	diffResult                    pullRequestDiffResult
	diffResultKnown               bool
}

func (model reviewSessionReadModel) isActive() bool {
	return model.active
}

func (model reviewSessionReadModel) metadataContent() string {
	reference := pullRequestReference(model.summary, githubdomain.PullRequestDetail{})
	return valueOrDash(strings.TrimSpace(reference))
}

func (model reviewSessionReadModel) detailContent() string {
	if !model.active {
		return ""
	}
	if model.showsDescription() {
		return model.descriptionContent()
	}
	if model.showsStoryChapter() {
		return model.storyChapterContent()
	}
	if !model.diffResultKnown {
		return model.loadingDetail()
	}
	if model.diffResult.err != nil {
		return model.diffErrorDetail(model.diffResult.err)
	}
	selectedFile, ok := model.selectedDiffFile()
	if !ok {
		return model.noDiffDetail()
	}
	return renderReviewDiffFileWithCollapsedThreadsForViewerAndWordWrap(selectedFile, model.markdownRenderer, model.detailWrapWidth, model.wordWrapEnabled, model.collapsedThreadIDs, model.connectedUserLogin)
}

func (model reviewSessionReadModel) descriptionContent() string {
	summary := model.summary
	if model.descriptionResultKnown {
		if model.descriptionResult.err != nil {
			return renderPullRequestDetailError(summary, model.descriptionResult.err)
		}

		header := renderPullRequestBrowserHeader(summary, model.descriptionResult.detail)
		overview := model.descriptionOverview()
		content := renderPullRequestDescriptionWithWordWrap(summary, model.descriptionResult.detail, model.markdownRenderer, model.detailWrapWidth, model.wordWrapEnabled)
		return renderPullRequestBrowserDetailContent(header, overview, content, model.detailWrapWidth)
	}

	return renderPullRequestDetailLoading(summary, model.loadingSpinner)
}

func (model reviewSessionReadModel) descriptionOverview() string {
	if !model.descriptionResultKnown || model.descriptionResult.err != nil {
		return ""
	}
	return browserDetailReadModel{
		summary:                model.summary,
		detail:                 model.descriptionResult.detail,
		width:                  model.detailWrapWidth,
		collapsedSectionStates: model.browserCollapsedSectionStates,
	}.renderOverview()
}

func (model reviewSessionReadModel) storyChapterContent() string {
	chapter, ok := model.selectedStoryChapter()
	if !ok {
		return model.noDiffDetail()
	}

	sections := []string{"# " + strings.TrimSpace(chapter.Title)}
	if strings.TrimSpace(chapter.Narrative) != "" {
		sections = append(sections, strings.TrimSpace(chapter.Narrative))
	}
	return renderMarkdownWithFallback(strings.Join(sections, "\n\n"), model.markdownRenderer, markdownRenderWidthForWordWrap(model.detailWrapWidth, model.wordWrapEnabled), "No chapter narrative is available.")
}

func (model reviewSessionReadModel) descriptionSummaryAndDetail() (githubdomain.PullRequest, githubdomain.PullRequestDetail, bool) {
	if !model.showsDescription() {
		return githubdomain.PullRequest{}, githubdomain.PullRequestDetail{}, false
	}
	if !model.descriptionResultKnown || model.descriptionResult.err != nil {
		return model.summary, githubdomain.PullRequestDetail{}, false
	}
	return model.summary, model.descriptionResult.detail, true
}

func (model reviewSessionReadModel) showsDescription() bool {
	if !model.active {
		return false
	}
	return model.mainContentKind == MainContentKindReviewDescription
}

func (model reviewSessionReadModel) showsStoryChapter() bool {
	if !model.active {
		return false
	}
	return model.mainContentKind == MainContentKindStoryChapter
}

func (model reviewSessionReadModel) loadingDetail() string {
	repository := pullRequestRepositoryName(model.summary.Repository)
	lines := []string{
		fmt.Sprintf("Pending review %s is open for %s#%d.", valueOrDash(model.pendingReviewID), repository, model.summary.Number),
		"",
		"Loading pull request diff...",
		fmt.Sprintf("Running `gh api repos/%s/pulls/%d -H 'Accept: application/vnd.github.v3.diff'`.", repository, model.summary.Number),
	}
	return strings.Join(lines, "\n")
}

func (model reviewSessionReadModel) diffErrorDetail(err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		message = "Unknown error. GitHub misplaced the diff again."
	}

	lines := []string{model.loadingDetail(), "", message}
	return strings.Join(lines, "\n")
}

func (model reviewSessionReadModel) noDiffDetail() string {
	repository := pullRequestRepositoryName(model.summary.Repository)
	lines := []string{
		fmt.Sprintf("Pending review %s is open for %s#%d.", valueOrDash(model.pendingReviewID), repository, model.summary.Number),
		"",
		"No changed files are available for this review.",
	}
	return strings.Join(lines, "\n")
}

func (model reviewSessionReadModel) detailIdentity() string {
	if !model.active {
		return ""
	}
	if model.showsDescription() {
		return fmt.Sprintf(
			"review:%s:%d:%s:description",
			pullRequestRepositoryName(model.summary.Repository),
			model.summary.Number,
			model.pendingReviewID,
		)
	}
	if chapter, ok := model.selectedStoryChapter(); ok {
		return fmt.Sprintf(
			"review:%s:%d:%s:chapter:%s",
			pullRequestRepositoryName(model.summary.Repository),
			model.summary.Number,
			model.pendingReviewID,
			chapter.ID,
		)
	}
	selectedFilePath := fmt.Sprintf("row:%d", model.selectedFileTreeRow)
	if selectedFile, ok := model.selectedDiffFile(); ok {
		selectedFilePath = selectedFile.Path
	}
	return fmt.Sprintf(
		"review:%s:%d:%s:file:%s",
		pullRequestRepositoryName(model.summary.Repository),
		model.summary.Number,
		model.pendingReviewID,
		selectedFilePath,
	)
}

func (model reviewSessionReadModel) files() []Item {
	if !model.active {
		return nil
	}

	tree, files, ok := model.currentTree()
	if !ok {
		if model.diffResultKnown && model.diffResult.err != nil {
			return []Item{{Title: "Could not load file tree", Detail: model.diffErrorDetail(model.diffResult.err)}}
		}
		return []Item{{Title: "Loading file tree...", Detail: model.loadingDetail()}}
	}
	if len(tree.Rows) == 0 {
		return []Item{{Title: "No changed files", Detail: model.noDiffDetail()}}
	}
	return reviewDiffTreeItems(tree, files)
}

func (model reviewSessionReadModel) selectedVisibleLine() int {
	selectableRows, ok := model.selectableRows()
	if !ok || len(selectableRows) == 0 {
		return 0
	}
	if model.selectedFileTreeRow < 0 {
		if model.mode != reviewSessionModeStory {
			if fileRows, fileRowsOK := model.fileRows(); fileRowsOK && len(fileRows) > 0 {
				return fileRows[0]
			}
		}
		return selectableRows[0]
	}
	return adjustVisibleSelection(model.selectedFileTreeRow, selectableRows, 0)
}

func (model reviewSessionReadModel) selectedDiffFile() (reviewDiffFile, bool) {
	row, files, ok := model.selectedTreeRow()
	if !ok {
		return reviewDiffFile{}, false
	}
	fileIndex := row.FileIndex
	if fileIndex < 0 {
		rawTree, _, rawTreeOK := model.rawTree()
		if !rawTreeOK {
			return reviewDiffFile{}, false
		}
		fileIndex, ok = reviewDiffTreeFirstDescendantFileIndex(rawTree, row.ID)
		if !ok {
			return reviewDiffFile{}, false
		}
	}
	if fileIndex < 0 || fileIndex >= len(files) {
		return reviewDiffFile{}, false
	}
	return files[fileIndex], true
}

func (model reviewSessionReadModel) selectableRows() ([]int, bool) {
	tree, _, ok := model.currentTree()
	if !ok {
		return nil, false
	}
	return reviewDiffSelectableTreeRowIndexes(tree), true
}

func (model reviewSessionReadModel) rawTree() (reviewDiffTree, []reviewDiffFile, bool) {
	if !model.diffResultKnown || model.diffResult.err != nil {
		return reviewDiffTree{}, nil, false
	}
	if model.mode == reviewSessionModeStory && len(model.story.Tree.Rows) > 0 {
		return model.story.Tree, model.diffResult.data.Files, true
	}
	return model.diffResult.data.FileTree, model.diffResult.data.Files, true
}

func (model reviewSessionReadModel) currentTree() (reviewDiffTree, []reviewDiffFile, bool) {
	tree, files, ok := model.rawTree()
	if !ok {
		return reviewDiffTree{}, nil, false
	}
	return reviewDiffTreeVisibleRows(tree, model.collapsedTreeRowIDs), files, true
}

func (model reviewSessionReadModel) selectedTreeRow() (reviewDiffTreeRow, []reviewDiffFile, bool) {
	tree, files, ok := model.currentTree()
	if !ok || len(tree.Rows) == 0 {
		return reviewDiffTreeRow{}, nil, false
	}
	rowIndex := clampIndex(model.selectedFileTreeRow, len(tree.Rows))
	return tree.Rows[rowIndex], files, true
}

func (model reviewSessionReadModel) selectedStoryChapter() (reviewStoryChapter, bool) {
	row, _, ok := model.selectedTreeRow()
	if !ok || row.Kind != reviewDiffTreeRowKindChapter {
		return reviewStoryChapter{}, false
	}
	if row.ChapterIndex < 0 || row.ChapterIndex >= len(model.story.Chapters) {
		return reviewStoryChapter{}, false
	}
	return model.story.Chapters[row.ChapterIndex], true
}

func (model reviewSessionReadModel) fileRows() ([]int, bool) {
	tree, _, ok := model.currentTree()
	if !ok {
		return nil, false
	}
	return reviewDiffSelectableRowIndexes(tree), true
}
