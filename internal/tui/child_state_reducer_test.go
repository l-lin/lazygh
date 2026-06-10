package tui

import (
	"reflect"
	"strings"
	"testing"
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestDetailStateModel_GivenStaleIdentity_WhenSyncing_ThenItReturnsUpdatedCopyWithoutMutatingTheOriginal(t *testing.T) {
	document := newDetailDocument("alpha\nbeta", 40)
	subject := detailStateModel{
		lastIdentity: "stale",
		viewState: detailViewState{
			cursor:          detailPosition{line: 99, column: 7},
			preferredColumn: 7,
		},
	}

	actual := subject.synced("fresh", document, 3, "")

	if actual.lastIdentity != "fresh" {
		t.Fatalf("expected synced identity %q, actual %q", "fresh", actual.lastIdentity)
	}
	if actual.viewState.cursor.line != 0 || actual.viewState.cursor.column != 0 {
		t.Fatalf("expected the synced cursor to reset to the top of the document, actual %+v", actual.viewState.cursor)
	}
	if actual.viewState.preferredColumn != 0 {
		t.Fatalf("expected the synced preferred column 0, actual %d", actual.viewState.preferredColumn)
	}
	if subject.lastIdentity != "stale" {
		t.Fatalf("expected the original identity to stay %q, actual %q", "stale", subject.lastIdentity)
	}
	if subject.viewState.cursor.line != 99 || subject.viewState.preferredColumn != 7 {
		t.Fatalf("expected the original detail view state to stay unchanged, actual %+v", subject.viewState)
	}
}

func TestReviewSessionState_GivenFoldableTree_WhenCollapsingAllRows_ThenItReturnsUpdatedCopyWithoutMutatingTheOriginal(t *testing.T) {
	tree := reviewDiffTree{Rows: []reviewDiffTreeRow{
		{ID: "dir", Foldable: true},
		{ID: "child", Foldable: true},
		{ID: "plain", Foldable: false},
	}}
	subject := reviewSessionState{collapsedTreeRowIDs: map[string]bool{"dir": false}}

	actual, changed := subject.withAllTreeRowsCollapsed(tree, true)
	if !changed {
		t.Fatal("expected the fold reducer to report a change")
	}
	if !actual.collapsedTreeRowIDs["dir"] || !actual.collapsedTreeRowIDs["child"] {
		t.Fatalf("expected all foldable rows to be collapsed, actual %v", actual.collapsedTreeRowIDs)
	}
	if subject.collapsedTreeRowIDs["dir"] {
		t.Fatalf("expected the original fold map to stay unchanged, actual %v", subject.collapsedTreeRowIDs)
	}
	if _, ok := subject.collapsedTreeRowIDs["child"]; ok {
		t.Fatalf("expected the original fold map to stay free of new entries, actual %v", subject.collapsedTreeRowIDs)
	}
}

func TestReviewSessionState_GivenLifecycleAndSummaryTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	start := reviewSessionStartDescriptor{
		mode:                         reviewSessionModeStory,
		sourceFocus:                  FocusDetailView,
		sourceDetailTab:              CommentsDetailTab,
		sourcePaneLayoutSize:         PaneLayoutFullscreen,
		sourceFullscreenPane:         FocusDetailView,
		sourceDetailFullscreenReturn: PaneLayoutDefault,
		summary:                      githubdomain.PullRequest{Number: 42, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}},
		pendingReviewID:              " PRR_pending ",
		story:                        reviewStoryData{Summary: "Story"},
	}
	subject := reviewSessionState{active: true, pendingReviewID: "stale", summary: githubdomain.PullRequest{Number: 7}}

	started := subject.started(start)
	resummarized := started.withSummary(githubdomain.PullRequest{Number: 99, Repository: githubdomain.Repository{NameWithOwner: "acme/rocket"}})
	cleared := started.cleared()

	if !started.active {
		t.Fatal("expected the started review session to be active")
	}
	if actual := started.mode; actual != reviewSessionModeStory {
		t.Fatalf("expected started mode %v, actual %v", reviewSessionModeStory, actual)
	}
	if actual := started.pendingReviewID; actual != "PRR_pending" {
		t.Fatalf("expected started pending review id %q, actual %q", "PRR_pending", actual)
	}
	if actual := started.selectedFileTreeRow; actual != -1 {
		t.Fatalf("expected started file-tree row %d, actual %d", -1, actual)
	}
	if started.collapsedTreeRowIDs == nil || started.collapsedThreadIDs == nil {
		t.Fatal("expected the started review session to initialize collapsed-id maps")
	}
	if actual := resummarized.summary.Number; actual != 99 {
		t.Fatalf("expected replaced review summary number %d, actual %d", 99, actual)
	}
	if actual := resummarized.summary.Repository.NameWithOwner; actual != "acme/rocket" {
		t.Fatalf("expected replaced review summary repository %q, actual %q", "acme/rocket", actual)
	}
	if !reflect.DeepEqual(cleared, reviewSessionState{}) {
		t.Fatalf("expected cleared review session %+v, actual %+v", reviewSessionState{}, cleared)
	}
	if actual := subject.pendingReviewID; actual != "stale" {
		t.Fatalf("expected the original pending review id %q, actual %q", "stale", actual)
	}
	if actual := subject.summary.Number; actual != 7 {
		t.Fatalf("expected the original review summary number %d, actual %d", 7, actual)
	}
	if !subject.active {
		t.Fatal("expected the original review session to stay active")
	}
}

func TestDetailStateModel_GivenActiveTabTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := detailStateModel{activeTab: CommentsDetailTab, wrapWidth: 72}

	selected := subject.withActiveTab(CommitsDetailTab)
	projected := subject.withProjectedScreenStateApplication(projectedScreenStateApplication{hasDetailTab: true, activeDetailTabIndex: 3}, browserDetailTabs)
	advancedForward := subject.withAdvancedActiveTab(1, browserDetailTabs)
	advancedBackward := subject.withAdvancedActiveTab(-1, browserDetailTabs)
	unchanged := subject.withProjectedScreenStateApplication(projectedScreenStateApplication{}, browserDetailTabs)

	if actual := selected.activeTab; actual != CommitsDetailTab {
		t.Fatalf("expected selected active tab %v, actual %v", CommitsDetailTab, actual)
	}
	if actual := projected.activeTab; actual != ChangesDetailTab {
		t.Fatalf("expected projected active tab %v, actual %v", ChangesDetailTab, actual)
	}
	if actual := advancedForward.activeTab; actual != CommitsDetailTab {
		t.Fatalf("expected forward advanced active tab %v, actual %v", CommitsDetailTab, actual)
	}
	if actual := advancedBackward.activeTab; actual != DescriptionDetailTab {
		t.Fatalf("expected backward advanced active tab %v, actual %v", DescriptionDetailTab, actual)
	}
	if actual := unchanged.activeTab; actual != CommentsDetailTab {
		t.Fatalf("expected unchanged active tab %v, actual %v", CommentsDetailTab, actual)
	}
	if actual := subject.activeTab; actual != CommentsDetailTab {
		t.Fatalf("expected the original active tab %v, actual %v", CommentsDetailTab, actual)
	}
	if actual := subject.wrapWidth; actual != 72 {
		t.Fatalf("expected the original wrap width %d, actual %d", 72, actual)
	}
}

func TestDetailStateModel_GivenMotionAndSearchSyncTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	document := newDetailDocument(strings.Join([]string{"Alpha", "Target one", "Omega", "Target two"}, "\n"), 40)
	subject := detailStateModel{viewState: detailViewState{
		cursor:                detailPosition{line: 1, column: 0},
		currentSearchMatch:    -1,
		searchCacheDocumentID: 99,
		searchCacheQuery:      "stale",
	}}
	scrollUpSubject := detailStateModel{viewState: detailViewState{cursor: detailPosition{line: 2, column: 0}, originRow: 1}}
	recenterSubject := detailStateModel{viewState: detailViewState{cursor: detailPosition{line: 3, column: 0}}}

	searchSynced := subject.withSearchSynced(document, "Target")
	focused := subject.withFocusedLineAndSearchSynced(document, 2, 3, "Target")
	scrolledDown := subject.withViewportOperation(document, 2, detailViewportOperationScrollDown)
	scrolledUp := scrollUpSubject.withViewportOperation(document, 2, detailViewportOperationScrollUp)
	recentered := recenterSubject.withViewportOperation(document, 2, detailViewportOperationRecenter)
	placedTop := subject.withViewportOperation(document, 2, detailViewportOperationPlaceTop)
	placedBottom := subject.withViewportOperation(document, 2, detailViewportOperationPlaceBottom)
	replaced := subject.withViewState(detailViewState{cursor: detailPosition{line: 2, column: 0}, originRow: 1, currentSearchMatch: 7})

	if actual := len(searchSynced.viewState.searchMatches); actual != 2 {
		t.Fatalf("expected synced search match count %d, actual %d", 2, actual)
	}
	if actual := searchSynced.viewState.currentSearchMatch; actual != 0 {
		t.Fatalf("expected synced current search match %d, actual %d", 0, actual)
	}
	if actual := focused.viewState.cursor.line; actual != 3 {
		t.Fatalf("expected focused cursor line %d, actual %d", 3, actual)
	}
	if actual := focused.viewState.currentSearchMatch; actual != 1 {
		t.Fatalf("expected focused current search match %d, actual %d", 1, actual)
	}
	if actual := scrolledDown.viewState.cursor.line; actual != 2 {
		t.Fatalf("expected scrolled-down cursor line %d, actual %d", 2, actual)
	}
	if actual := scrolledDown.viewState.originRow; actual != 1 {
		t.Fatalf("expected scrolled-down origin row %d, actual %d", 1, actual)
	}
	if !scrolledDown.viewState.manualViewportScroll {
		t.Fatal("expected scrolling down to keep manual viewport scrolling enabled")
	}
	if actual := scrolledUp.viewState.cursor.line; actual != 1 {
		t.Fatalf("expected scrolled-up cursor line %d, actual %d", 1, actual)
	}
	if actual := scrolledUp.viewState.originRow; actual != 0 {
		t.Fatalf("expected scrolled-up origin row %d, actual %d", 0, actual)
	}
	expectedRecenterOrigin := centeredViewportOrigin(3, 2, document.rowCount())
	if actual := recentered.viewState.originRow; actual != expectedRecenterOrigin {
		t.Fatalf("expected recentered origin row %d, actual %d", expectedRecenterOrigin, actual)
	}
	if actual := recentered.viewState.preserveViewportSyncCount; actual != 3 {
		t.Fatalf("expected recentered preserve-sync count %d, actual %d", 3, actual)
	}
	expectedTopOrigin := placedViewportOrigin(1, 2, document.rowCount(), viewportPlacementTop)
	if actual := placedTop.viewState.originRow; actual != expectedTopOrigin {
		t.Fatalf("expected top-placed origin row %d, actual %d", expectedTopOrigin, actual)
	}
	if actual := placedTop.viewState.preserveViewportSyncCount; actual != 3 {
		t.Fatalf("expected top-placed preserve-sync count %d, actual %d", 3, actual)
	}
	expectedBottomOrigin := placedViewportOrigin(1, 2, document.rowCount(), viewportPlacementBottom)
	if actual := placedBottom.viewState.originRow; actual != expectedBottomOrigin {
		t.Fatalf("expected bottom-placed origin row %d, actual %d", expectedBottomOrigin, actual)
	}
	if actual := placedBottom.viewState.preserveViewportSyncCount; actual != 3 {
		t.Fatalf("expected bottom-placed preserve-sync count %d, actual %d", 3, actual)
	}
	if actual := replaced.viewState.cursor.line; actual != 2 {
		t.Fatalf("expected replaced cursor line %d, actual %d", 2, actual)
	}
	if actual := replaced.viewState.originRow; actual != 1 {
		t.Fatalf("expected replaced origin row %d, actual %d", 1, actual)
	}
	if actual := replaced.viewState.currentSearchMatch; actual != 7 {
		t.Fatalf("expected replaced current search match %d, actual %d", 7, actual)
	}
	if actual := subject.viewState.cursor.line; actual != 1 {
		t.Fatalf("expected the original cursor line %d, actual %d", 1, actual)
	}
	if actual := subject.viewState.currentSearchMatch; actual != -1 {
		t.Fatalf("expected the original current search match %d, actual %d", -1, actual)
	}
	if actual := subject.viewState.searchCacheQuery; actual != "stale" {
		t.Fatalf("expected the original search cache query %q, actual %q", "stale", actual)
	}
	if len(subject.viewState.searchMatches) != 0 {
		t.Fatalf("expected the original search matches to stay empty, actual %+v", subject.viewState.searchMatches)
	}
	if subject.viewState.manualViewportScroll {
		t.Fatal("expected the original manual viewport scroll flag to stay false")
	}
}

func TestDetailStateModel_GivenPendingPrefixAndVisualModeTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := detailStateModel{viewState: detailViewState{
		cursor:                 detailPosition{line: 3, column: 4},
		visualAnchor:           detailPosition{line: 1, column: 2},
		mode:                   detailVisualMode,
		pendingKeySequence:     keySequenceState{pendingTarget: keySequenceTarget{viewName: viewDetailName}},
		pendingCharacterMotion: detailPendingCharacterMotion{active: true, direction: detailCharacterMotionDirectionForward, mode: detailCharacterMotionMatch},
		pendingYank:            true,
	}}

	cleared := subject.withPendingPrefixCleared()
	exited := subject.withVisualModeExited()

	if cleared.viewState.pendingKeySequence != (keySequenceState{}) {
		t.Fatalf("expected the cleared pending key sequence to reset, actual %+v", cleared.viewState.pendingKeySequence)
	}
	if cleared.viewState.pendingCharacterMotion != (detailPendingCharacterMotion{}) {
		t.Fatalf("expected the cleared pending character motion to reset, actual %+v", cleared.viewState.pendingCharacterMotion)
	}
	if cleared.viewState.pendingYank {
		t.Fatal("expected the cleared pending yank flag to reset")
	}
	if actual := cleared.viewState.mode; actual != detailVisualMode {
		t.Fatalf("expected clear-only mode %v, actual %v", detailVisualMode, actual)
	}
	if actual := exited.viewState.mode; actual != detailNormalMode {
		t.Fatalf("expected exited mode %v, actual %v", detailNormalMode, actual)
	}
	if actual := exited.viewState.visualAnchor; actual != subject.viewState.cursor {
		t.Fatalf("expected exited visual anchor %+v, actual %+v", subject.viewState.cursor, actual)
	}
	if exited.viewState.pendingYank {
		t.Fatal("expected exiting visual mode to clear the pending yank flag")
	}
	if subject.viewState.pendingKeySequence == (keySequenceState{}) {
		t.Fatalf("expected the original pending key sequence to stay armed, actual %+v", subject.viewState.pendingKeySequence)
	}
	if subject.viewState.pendingCharacterMotion == (detailPendingCharacterMotion{}) {
		t.Fatalf("expected the original pending character motion to stay armed, actual %+v", subject.viewState.pendingCharacterMotion)
	}
	if !subject.viewState.pendingYank {
		t.Fatal("expected the original pending yank flag to stay true")
	}
	if actual := subject.viewState.mode; actual != detailVisualMode {
		t.Fatalf("expected the original mode %v, actual %v", detailVisualMode, actual)
	}
	if actual := subject.viewState.visualAnchor; actual != (detailPosition{line: 1, column: 2}) {
		t.Fatalf("expected the original visual anchor %+v, actual %+v", detailPosition{line: 1, column: 2}, actual)
	}
}

func TestDetailStateModel_GivenPendingKeySequenceTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	originalTarget := keySequenceTargetFor(viewDetailName, keymapScopeSelection, "move_selection_to_top")
	replacementTarget := keySequenceTargetFor(viewDetailName, keymapScopeSelection, "recenter_selection")
	subject := detailStateModel{viewState: detailViewState{pendingKeySequence: keySequenceState{pendingTarget: originalTarget}}}

	armed := subject.withPendingKeySequenceArmed(replacementTarget)
	cleared := subject.withPendingKeySequenceCleared()

	if actual := armed.pendingKeySequenceTarget(); actual != replacementTarget {
		t.Fatalf("expected armed pending key sequence target %+v, actual %+v", replacementTarget, actual)
	}
	if actual := cleared.pendingKeySequenceTarget(); actual != (keySequenceTarget{}) {
		t.Fatalf("expected cleared pending key sequence target %+v, actual %+v", keySequenceTarget{}, actual)
	}
	if actual := subject.pendingKeySequenceTarget(); actual != originalTarget {
		t.Fatalf("expected the original pending key sequence target %+v, actual %+v", originalTarget, actual)
	}
}

func TestDetailStateModel_GivenYankHighlightLifecycleTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	now := time.Date(2026, time.May, 29, 14, 0, 0, 0, time.UTC)
	selection := detailSelectionRange{start: detailPosition{line: 0, column: 0}, end: detailPosition{line: 0, column: 4}}
	subject := detailStateModel{viewState: detailViewState{yankHighlight: detailYankHighlightState{active: true, start: detailPosition{line: 1, column: 0}, end: detailPosition{line: 1, column: 2}, expiresAt: now.Add(-time.Second)}}}

	activated := subject.withYankHighlightActivated(selection, now.Add(time.Second))
	cleared, changed := subject.withExpiredYankHighlightCleared(now)

	if !activated.hasYankHighlight() {
		t.Fatal("expected activated detail state to report an active yank highlight")
	}
	if actual := activated.viewState.yankHighlight.start; actual != selection.start {
		t.Fatalf("expected activated highlight start %+v, actual %+v", selection.start, actual)
	}
	if actual := activated.viewState.yankHighlight.end; actual != selection.end {
		t.Fatalf("expected activated highlight end %+v, actual %+v", selection.end, actual)
	}
	if actual := activated.viewState.yankHighlight.expiresAt; !actual.Equal(now.Add(time.Second)) {
		t.Fatalf("expected activated expiry %v, actual %v", now.Add(time.Second), actual)
	}
	if !changed {
		t.Fatal("expected expired yank highlight cleanup to report a change")
	}
	if cleared.hasYankHighlight() {
		t.Fatalf("expected cleared detail state to drop the yank highlight, actual %+v", cleared.viewState.yankHighlight)
	}
	if !subject.hasYankHighlight() {
		t.Fatal("expected the original detail state to keep its yank highlight")
	}
	if actual := subject.viewState.yankHighlight.start; actual != (detailPosition{line: 1, column: 0}) {
		t.Fatalf("expected the original highlight start %+v, actual %+v", detailPosition{line: 1, column: 0}, actual)
	}
}
