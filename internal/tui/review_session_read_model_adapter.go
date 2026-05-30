package tui

import (
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) reviewSessionReadModel() reviewSessionReadModel {
	if program == nil {
		return reviewSessionReadModel{}
	}
	if program.refreshReadCache.enabled && program.refreshReadCache.reviewSessionReadModelSet {
		return program.refreshReadCache.reviewSessionReadModel
	}

	model := program.buildReviewSessionReadModel()
	program.cacheReviewSessionReadModel(model)
	return model
}

func (program *Program) buildReviewSessionReadModel() reviewSessionReadModel {
	if program == nil {
		return reviewSessionReadModel{}
	}

	state := program.navigationState.reviewSession
	model := reviewSessionReadModel{
		active:                        state.active,
		mode:                          state.mode,
		summary:                       state.summary,
		pendingReviewID:               strings.TrimSpace(state.pendingReviewID),
		selectedFileTreeRow:           state.selectedFileTreeRow,
		collapsedTreeRowIDs:           state.collapsedTreeRowIDs,
		collapsedThreadIDs:            state.collapsedThreadIDs,
		story:                         state.story,
		detailWrapWidth:               program.detailState.wrapWidth,
		wordWrapEnabled:               program.detailWordWrapEnabled(),
		markdownRenderer:              program.markdownRenderer,
		connectedUserLogin:            program.currentConnectedUserLogin(),
		loadingSpinner:                program.loadingSpinnerFrame(),
		browserCollapsedSectionStates: program.browserCollapsedSectionStates,
	}
	if !model.active {
		return model
	}

	model.mainContentKind = program.reviewSessionMainContentKind()
	if result, ok := program.pullRequestDetailForSummary(model.summary); ok {
		model.descriptionResult = result
		model.descriptionResultKnown = true
	}
	if result, ok := program.pullRequestDiffForSummary(model.summary); ok {
		model.diffResult = result
		model.diffResultKnown = true
	}
	return model
}

func (program *Program) reviewSessionMetadataContent() string {
	return program.reviewSessionReadModel().metadataContent()
}

func (program *Program) reviewSessionDetailContent() string {
	return program.reviewSessionReadModel().detailContent()
}

func (program *Program) reviewSessionDescriptionSummaryAndDetail() (githubdomain.PullRequest, githubdomain.PullRequestDetail, bool) {
	readModel := program.reviewSessionReadModel()
	return readModel.descriptionSummaryAndDetail()
}

func (program *Program) reviewSessionShowsDescription() bool {
	return program.reviewSessionMainContentKind() == MainContentKindReviewDescription
}

func (program *Program) reviewSessionShowsStoryChapter() bool {
	return program.reviewSessionMainContentKind() == MainContentKindStoryChapter
}

func (program *Program) reviewSessionDetailIdentity() string {
	state := program.navigationState.reviewSession
	if !state.active {
		return ""
	}

	repositoryName := pullRequestRepositoryName(state.summary.Repository)
	pendingReviewID := strings.TrimSpace(state.pendingReviewID)
	switch program.reviewSessionMainContentKind() {
	case MainContentKindReviewDescription:
		return fmt.Sprintf("review:%s:%d:%s:description", repositoryName, state.summary.Number, pendingReviewID)
	case MainContentKindStoryChapter:
		if chapter, ok := program.selectedReviewSessionStoryChapter(); ok {
			return fmt.Sprintf("review:%s:%d:%s:chapter:%s", repositoryName, state.summary.Number, pendingReviewID, chapter.ID)
		}
	}

	selectedFilePath := fmt.Sprintf("row:%d", state.selectedFileTreeRow)
	if selectedFile, ok := program.selectedReviewSessionDiffFile(); ok {
		selectedFilePath = selectedFile.Path
	}
	return fmt.Sprintf("review:%s:%d:%s:file:%s", repositoryName, state.summary.Number, pendingReviewID, selectedFilePath)
}

func (program *Program) reviewSessionFiles() []Item {
	readModel := program.reviewSessionReadModel()
	if !readModel.active {
		return nil
	}
	return readModel.files()
}

func (program *Program) reviewSessionSelectedVisibleLine() int {
	return program.reviewSessionReadModel().selectedVisibleLine()
}

func (program *Program) selectedReviewSessionDiffFile() (reviewDiffFile, bool) {
	data, ok := program.reviewSessionDiffData()
	if !ok {
		return reviewDiffFile{}, false
	}
	return reviewSessionSelectedDiffFile(program.navigationState.reviewSession, data)
}

func (program *Program) reviewSessionSelectableRows() ([]int, bool) {
	return program.reviewSessionReadModel().selectableRows()
}

func (program *Program) reviewSessionRawTree() (reviewDiffTree, []reviewDiffFile, bool) {
	data, ok := program.reviewSessionDiffData()
	if !ok {
		return reviewDiffTree{}, nil, false
	}
	tree, files := reviewSessionRawTree(program.navigationState.reviewSession, data)
	return tree, files, true
}

func (program *Program) reviewSessionCurrentTree() (reviewDiffTree, []reviewDiffFile, bool) {
	return program.reviewSessionReadModel().currentTree()
}

func (program *Program) selectedReviewSessionStoryChapter() (reviewStoryChapter, bool) {
	data, ok := program.reviewSessionDiffData()
	if !ok {
		return reviewStoryChapter{}, false
	}
	return reviewSessionSelectedStoryChapter(program.navigationState.reviewSession, data)
}

func (program *Program) reviewSessionFileRows() ([]int, bool) {
	return program.reviewSessionReadModel().fileRows()
}
