package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) reviewSessionReadModel() reviewSessionReadModel {
	if program == nil {
		return reviewSessionReadModel{}
	}

	state := program.navigationState.reviewSession
	model := reviewSessionReadModel{
		active:              state.active,
		mode:                state.mode,
		summary:             state.summary,
		pendingReviewID:     strings.TrimSpace(state.pendingReviewID),
		selectedFileTreeRow: state.selectedFileTreeRow,
		collapsedTreeRowIDs: copyStringBoolMap(state.collapsedTreeRowIDs),
		collapsedThreadIDs:  copyStringBoolMap(state.collapsedThreadIDs),
		story:               state.story,
		detailWrapWidth:     program.detailState.wrapWidth,
		markdownRenderer:    program.markdownRenderer,
		connectedUserLogin:  program.currentConnectedUserLogin(),
		loadingSpinner:      program.loadingSpinnerFrame(),
	}
	if !model.active {
		return model
	}

	if result, ok := program.pullRequestDetailForSummary(model.summary); ok {
		model.descriptionResult = result
		model.descriptionResultKnown = true
		if result.err == nil {
			model.descriptionOverview = program.renderCurrentPullRequestOverview(model.summary, result.detail, model.detailWrapWidth)
		}
	}
	if result, ok := program.pullRequestDiffForSummary(model.summary); ok {
		model.diffResult = result
		model.diffResultKnown = true
	}

	stateSnapshot := program.screenState()
	mainView := stateSnapshot.MainViewResolver()
	model.mainContentKind = mainView.ContentKind
	if stateSnapshot.Mode == ScreenModeStoryReview && mainView.SourceView.Focus == FocusPullRequestsView {
		if _, ok := model.selectedStoryChapter(); ok {
			model.mainContentKind = MainContentKindStoryChapter
		}
	}
	return model
}

func copyStringBoolMap(source map[string]bool) map[string]bool {
	if len(source) == 0 {
		return nil
	}

	copied := make(map[string]bool, len(source))
	for key, value := range source {
		copied[key] = value
	}
	return copied
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
	return program.reviewSessionReadModel().showsDescription()
}

func (program *Program) reviewSessionShowsStoryChapter() bool {
	return program.reviewSessionReadModel().showsStoryChapter()
}

func (program *Program) reviewSessionDetailIdentity() string {
	return program.reviewSessionReadModel().detailIdentity()
}

func (program *Program) reviewSessionFiles() []Item {
	readModel := program.reviewSessionReadModel()
	if !readModel.active {
		return nil
	}
	program.clampReviewSessionSelection()
	return program.reviewSessionReadModel().files()
}

func (program *Program) reviewSessionSelectedVisibleLine() int {
	readModel := program.reviewSessionReadModel()
	if !readModel.diffResultKnown || readModel.diffResult.err != nil {
		return 0
	}
	program.clampReviewSessionSelection()
	return program.reviewSessionReadModel().selectedVisibleLine()
}

func (program *Program) selectedReviewSessionDiffFile() (reviewDiffFile, bool) {
	return program.reviewSessionReadModel().selectedDiffFile()
}

func (program *Program) clampReviewSessionSelection() {
	readModel := program.reviewSessionReadModel()
	selectableRows, ok := readModel.selectableRows()
	if !ok {
		program.navigationState.reviewSession = program.navigationState.reviewSession.clampedSelection(nil, nil)
		return
	}

	fileRows, fileRowsOK := readModel.fileRows()
	if !fileRowsOK {
		fileRows = nil
	}
	program.navigationState.reviewSession = program.navigationState.reviewSession.clampedSelection(selectableRows, fileRows)
}

func (program *Program) adjustReviewSessionSelection(change int) {
	selectableRows, ok := program.reviewSessionReadModel().selectableRows()
	if !ok {
		program.navigationState.reviewSession = program.navigationState.reviewSession.adjustedSelection(nil, change)
		return
	}

	program.navigationState.reviewSession = program.navigationState.reviewSession.adjustedSelection(selectableRows, change)
}

func (program *Program) moveReviewSessionSelectionToTop() {
	selectableRows, ok := program.reviewSessionReadModel().selectableRows()
	if !ok {
		program.navigationState.reviewSession = program.navigationState.reviewSession.selectionAtTop(nil)
		return
	}

	program.navigationState.reviewSession = program.navigationState.reviewSession.selectionAtTop(selectableRows)
}

func (program *Program) moveReviewSessionSelectionToBottom() {
	selectableRows, ok := program.reviewSessionReadModel().selectableRows()
	if !ok {
		program.navigationState.reviewSession = program.navigationState.reviewSession.selectionAtBottom(nil)
		return
	}

	program.navigationState.reviewSession = program.navigationState.reviewSession.selectionAtBottom(selectableRows)
}

func (program *Program) reviewSessionSelectableRows() ([]int, bool) {
	return program.reviewSessionReadModel().selectableRows()
}

func (program *Program) reviewSessionRawTree() (reviewDiffTree, []reviewDiffFile, bool) {
	return program.reviewSessionReadModel().rawTree()
}

func (program *Program) reviewSessionCurrentTree() (reviewDiffTree, []reviewDiffFile, bool) {
	return program.reviewSessionReadModel().currentTree()
}

func (program *Program) selectedReviewSessionStoryChapter() (reviewStoryChapter, bool) {
	return program.reviewSessionReadModel().selectedStoryChapter()
}

func (program *Program) reviewSessionFileRows() ([]int, bool) {
	return program.reviewSessionReadModel().fileRows()
}
