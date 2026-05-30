package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestDetailStateModel_GivenWordWrapToggle_WhenUpdating_ThenItReturnsUpdatedCopyWithoutMutatingTheOriginal(t *testing.T) {
	subject := detailStateModel{wrapWidth: 72}

	actual := subject.withWordWrapToggled()

	if !subject.wordWrapEnabled() {
		t.Fatal("expected the original detail state to keep word wrap enabled")
	}
	if actual.wordWrapEnabled() {
		t.Fatal("expected the toggled detail state to disable word wrap")
	}
	if actual.wrapWidth != 72 {
		t.Fatalf("expected the toggled wrap width %d, actual %d", 72, actual.wrapWidth)
	}
	if subject.wrapWidth != 72 {
		t.Fatalf("expected the original wrap width %d, actual %d", 72, subject.wrapWidth)
	}
}

func TestUpdate_GivenMsgToggleDetailWordWrapRequested_WhenApplying_ThenItTogglesWordWrapInvalidatesRenderCachesAndClosesTheActionsPopup(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.OpenDetail()
	model.OpenActionsPopup(1)
	subject := given_pullRequestCommentProgram(model, &fakePullRequestDetailLoader{})
	subject.pullRequestDetailDocumentCache[pullRequestDetailDocumentCacheKey{pullRequestKey: "acme/widgets#42", tab: DescriptionDetailTab, width: 80}] = detailDocument{width: 80}
	subject.pullRequestConversationDocumentCache[pullRequestDetailDocumentCacheKey{pullRequestKey: "acme/widgets#42", tab: CommentsDetailTab, width: 80}] = browserConversationDocument{text: "cached"}
	subject.reviewDiffRenderCache[reviewDiffRenderCacheKey{filePath: "internal/tui/render.go", width: 80}] = reviewDiffRenderCacheEntry{document: detailDocument{width: 80}}

	Update(subject, MsgToggleDetailWordWrapRequested{})

	if subject.detailState.wordWrapEnabled() {
		t.Fatal("expected word wrap to be disabled after toggling it once")
	}
	if len(subject.pullRequestDetailDocumentCache) != 0 || len(subject.pullRequestConversationDocumentCache) != 0 {
		t.Fatalf("expected pull request detail render caches to be invalidated, actual detail=%d conversation=%d", len(subject.pullRequestDetailDocumentCache), len(subject.pullRequestConversationDocumentCache))
	}
	if len(subject.reviewDiffRenderCache) != 0 {
		t.Fatalf("expected the review diff render cache to be invalidated, actual %d entries", len(subject.reviewDiffRenderCache))
	}
	if subject.model.ActionsPopupVisible() {
		t.Fatal("expected the accepted popup action to close the actions popup")
	}
	if actual := subject.feedbackMessage; actual != detailWordWrapDisabledMessage {
		t.Fatalf("expected feedback message %q, actual %q", detailWordWrapDisabledMessage, actual)
	}
}

func TestCurrentActionsPopupActions_GivenDetailViewFocus_WhenResolving_ThenItIncludesTheWordWrapAction(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)

	actual := subject.currentActionsPopupActions()

	if !given_hasActionTitle(actual, disableDetailWordWrapActionTitle) {
		t.Fatalf("expected actions to contain %q, actual %v", disableDetailWordWrapActionTitle, given_actionTitles(actual))
	}
}

func TestCurrentActionsPopupActions_GivenPullRequestsView_WhenResolving_ThenItOmitsTheWordWrapAction(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())

	actual := subject.currentActionsPopupActions()

	if given_hasActionTitle(actual, enableDetailWordWrapActionTitle) || given_hasActionTitle(actual, disableDetailWordWrapActionTitle) {
		t.Fatalf("expected the word-wrap action to stay off non-detail views, actual %v", given_actionTitles(actual))
	}
}

func TestKeybindingSpecs_GivenProgram_WhenListingWordWrapBindings_ThenCtrlWIsAvailableOnlyInTheDetailView(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: gocui.KeyCtrlW, handler: subject.toggleDetailWordWrap})
	then_bindingDoesNotExist(t, actual, viewPullRequestsName, gocui.KeyCtrlW)
	then_bindingDoesNotExist(t, actual, viewUserName, gocui.KeyCtrlW)
}

func TestHelpPopup_GivenDetailFocus_WhenTogglingHelp_ThenItShowsTheWordWrapBinding(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	then_helpEntryUsesKey(t, helpView.Buffer(), detailWordWrapHelpLabel, "Ctrl+W")
}

func TestLayout_GivenDescriptionTabWithALongMarkdownParagraphAndDisabledWordWrap_WhenBuildingViewZeroDocument_ThenItKeepsTheParagraphOnOneVisibleLine(t *testing.T) {
	markdown := strings.Repeat("wordwraptoggle ", 24)
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {Title: "First PR", Number: 42, Body: markdown, BaseRefName: "main", HeadRefName: "feature/wrap-toggle", State: "OPEN"},
		},
	})
	subject.model.OpenDetail()
	Update(subject, MsgToggleDetailWordWrapRequested{})
	gui := given_headlessGuiWithSize(t, 40, 20)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualDescriptionLineCount := 0
	for _, line := range detailView.BufferLines() {
		if strings.Contains(line, "wordwraptoggle") {
			actualDescriptionLineCount++
		}
	}
	if actualDescriptionLineCount != 1 {
		t.Fatalf("expected the description to stay on one visible line with word wrap disabled, actual %d in %q", actualDescriptionLineCount, detailView.Buffer())
	}
}

func TestLayout_GivenChangesTabWithALongDiffLineAndEnabledWordWrap_WhenBuildingViewZeroDocument_ThenItWrapsTheDiffAcrossVisibleLines(t *testing.T) {
	longLine := strings.Repeat("browserdiffwrap ", 12)
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": {
				UnifiedDiff: strings.Join([]string{
					"diff --git a/internal/tui/render.go b/internal/tui/render.go",
					"index 1111111..2222222 100644",
					"--- a/internal/tui/render.go",
					"+++ b/internal/tui/render.go",
					"@@ -1,1 +1,1 @@",
					"+" + longLine,
				}, "\n"),
				Files: []githubcli.PullRequestDiffFile{{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 1}},
			},
		},
	})
	gui := given_headlessGuiWithSize(t, 60, 30)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	lineIndex, actualLine := given_detailDocumentLineContaining(t, document, "browserdiffwrap browserdiffwrap browserdiffwrap")

	if actual := reviewDiffDocumentRowCountForLine(document, lineIndex); actual < 2 {
		t.Fatalf("expected the browser changes diff to wrap across multiple rendered rows, actual %d for %q", actual, actualLine)
	}
}

func TestReviewMode_GivenAnInlineThreadWithALongMarkdownCommentAndDisabledWordWrap_WhenBuildingViewZeroDocument_ThenItKeepsTheThreadBodyOnOneVisibleLine(t *testing.T) {
	threadBody := strings.Repeat("nowrapthreadbody ", 18)
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": {
				UnifiedDiff: strings.Join([]string{
					"diff --git a/internal/tui/render.go b/internal/tui/render.go",
					"index 1111111..2222222 100644",
					"--- a/internal/tui/render.go",
					"+++ b/internal/tui/render.go",
					"@@ -1,1 +1,1 @@",
					"+new line",
				}, "\n"),
				Files: []githubcli.PullRequestDiffFile{{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 1}},
				Threads: []githubcli.PullRequestReviewThread{{
					ID:       "thread-1",
					Path:     "internal/tui/render.go",
					Line:     1,
					DiffSide: "RIGHT",
					Comments: []githubcli.PullRequestComment{{
						Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
						Body:      threadBody,
						CreatedAt: "2026-05-05T10:00:00Z",
					}},
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 60, 30)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	subject.model.FocusDetailView()
	Update(subject, MsgToggleDetailWordWrapRequested{})
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualWrappedLineCount := 0
	for _, line := range detailView.BufferLines() {
		if strings.Contains(line, "nowrapthreadbody") {
			actualWrappedLineCount++
		}
	}
	if actualWrappedLineCount != 1 {
		t.Fatalf("expected the review thread body to stay on one visible line with word wrap disabled, actual %d in %q", actualWrappedLineCount, detailView.Buffer())
	}
}
