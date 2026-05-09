package tui

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestKeybindingSpecs_GivenProgram_WhenListingBuildRunPopupBindings_ThenItUsesDetailLikeNavigation(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'h', handler: subject.movePullRequestBuildRunPopupCursorLeft})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'j', handler: subject.movePullRequestBuildRunPopupCursorDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: gocui.KeyArrowDown, handler: subject.movePullRequestBuildRunPopupCursorDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'k', handler: subject.movePullRequestBuildRunPopupCursorUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: gocui.KeyArrowUp, handler: subject.movePullRequestBuildRunPopupCursorUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'l', handler: subject.movePullRequestBuildRunPopupCursorRight})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: '0', handler: subject.movePullRequestBuildRunPopupCursorToRowStart})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: '$', handler: subject.movePullRequestBuildRunPopupCursorToRowEnd})
	then_bindingKeyExists(t, actual, viewPullRequestBuildInfoName, 'g')
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'G', handler: subject.movePullRequestBuildRunPopupCursorToBottom})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'w', handler: subject.movePullRequestBuildRunPopupCursorToNextWord})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'e', handler: subject.movePullRequestBuildRunPopupCursorToWordEnd})
	then_bindingKeyExists(t, actual, viewPullRequestBuildInfoName, 'b')
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'W', handler: subject.movePullRequestBuildRunPopupCursorToNextBigWord})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'E', handler: subject.movePullRequestBuildRunPopupCursorToBigWordEnd})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'B', handler: subject.movePullRequestBuildRunPopupCursorToPreviousBigWord})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'v', handler: subject.enterPullRequestBuildRunPopupVisualMode})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'V', handler: subject.enterPullRequestBuildRunPopupLineVisualMode})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: '/', handler: subject.openSearch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'n', handler: subject.nextPullRequestBuildRunPopupSearchMatch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'N', handler: subject.previousPullRequestBuildRunPopupSearchMatch})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'y', handler: subject.copyPullRequestBuildRunPopupContent})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: gocui.KeyCtrlD, handler: subject.pagePullRequestBuildRunPopupDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: gocui.KeyCtrlU, handler: subject.pagePullRequestBuildRunPopupUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: gocui.KeyCtrlF, handler: subject.fullPagePullRequestBuildRunPopupDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: gocui.KeyPgdn, handler: subject.fullPagePullRequestBuildRunPopupDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: gocui.KeyCtrlB, handler: subject.fullPagePullRequestBuildRunPopupUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: gocui.KeyPgup, handler: subject.fullPagePullRequestBuildRunPopupUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: gocui.KeyEsc, handler: subject.closePullRequestBuildRunPopup})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'q', handler: subject.closePullRequestBuildRunPopup})
}

func TestBrowserMode_GivenTheCursorOnANonPendingBuild_WhenPressingEnter_ThenItCollapsesTheBuildsSectionWithoutOpeningTheBuildRunPopup(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/build-run-popup",
				State:       "OPEN",
				StatusCheckRollup: []githubcli.PullRequestStatusCheck{{
					Name:         "test",
					WorkflowName: "CI",
					Status:       "COMPLETED",
					Conclusion:   "FAILURE",
					Link:         "https://github.com/acme/widgets/actions/runs/42",
				}},
			},
		},
	}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: loader.details["acme/widgets#42"]}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "CI / test (Failed)")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	enterHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyEnter)
	actualErr = enterHandler(gui, detailView)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 0 {
		t.Fatalf("expected no queued build run load, actual %d", len(asyncRunner.runs))
	}
	then_viewDoesNotExist(t, gui, viewPullRequestBuildInfoName)
	if !strings.Contains(detailView.Buffer(), " "+pullRequestOverviewFailureIcon+" Builds") {
		t.Fatalf("expected enter to collapse the builds section, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "CI / test (Failed)") {
		t.Fatalf("expected the collapsed builds section to hide the build row, actual %q", detailView.Buffer())
	}
	then_currentViewNameIs(t, gui, viewDetailName)
}

func TestBrowserMode_GivenTheCursorOnAPendingBuild_WhenPressingEnter_ThenItStillTogglesTheBuildsSectionWithoutLoadingAnything(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/build-run-popup",
				State:       "OPEN",
				StatusCheckRollup: []githubcli.PullRequestStatusCheck{
					{Name: "test", WorkflowName: "CI", Status: "COMPLETED", Conclusion: "FAILURE", Link: "https://github.com/acme/widgets/actions/runs/42"},
					{Name: "deploy", Status: "IN_PROGRESS"},
				},
			},
		},
	}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: loader.details["acme/widgets#42"]}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "deploy (Pending)")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	enterHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyEnter)
	actualErr = enterHandler(gui, detailView)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 0 {
		t.Fatalf("expected no queued build run load, actual %d", len(asyncRunner.runs))
	}
	then_viewDoesNotExist(t, gui, viewPullRequestBuildInfoName)
	if !strings.Contains(detailView.Buffer(), " "+pullRequestOverviewFailureIcon+" Builds") {
		t.Fatalf("expected enter to collapse the builds section, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "deploy (Pending)") {
		t.Fatalf("expected the collapsed builds section to hide the pending build row, actual %q", detailView.Buffer())
	}
	then_currentViewNameIs(t, gui, viewDetailName)
}

func TestPullRequestBuildRunPopup_GivenVisible_WhenNavigating_ThenItMovesTheCursorAndPagesLikeTheDetailView(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.OpenDetail()
	subject := given_pullRequestCommentProgram(model, &fakePullRequestDetailLoader{})
	contentLines := make([]string, 0, 30)
	for lineNumber := 1; lineNumber <= 30; lineNumber++ {
		contentLines = append(contentLines, "Line "+strconv.Itoa(lineNumber))
	}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{
		checkTitle: "CI / test",
		runURL:     "https://github.com/acme/widgets/actions/runs/42",
		body:       strings.Join(contentLines, "\n"),
	})
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestBuildInfoName)
	if !gui.Cursor {
		t.Fatal("expected the cursor to be visible inside the build run popup")
	}

	downHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestBuildInfoName, 'j')
	actualErr = downHandler(gui, popupView)
	then_noError(t, actualErr)
	if actual := subject.pullRequestBuildRunPopup.viewState.cursor.line; actual != 1 {
		t.Fatalf("expected popup cursor line %d, actual %d", 1, actual)
	}

	halfPageDownHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestBuildInfoName, gocui.KeyCtrlD)
	actualErr = halfPageDownHandler(gui, popupView)
	then_noError(t, actualErr)
	if actual := subject.pullRequestBuildRunPopup.viewState.originRow; actual <= 0 {
		t.Fatalf("expected popup origin row to advance after half-page down, actual %d", actual)
	}

	fullPageDownHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestBuildInfoName, gocui.KeyPgdn)
	previousOriginRow := subject.pullRequestBuildRunPopup.viewState.originRow
	actualErr = fullPageDownHandler(gui, popupView)
	then_noError(t, actualErr)
	if actual := subject.pullRequestBuildRunPopup.viewState.originRow; actual <= previousOriginRow {
		t.Fatalf("expected popup origin row to advance after full-page down, previous %d actual %d", previousOriginRow, actual)
	}
}

func TestPullRequestBuildRunPopup_GivenVisible_WhenSearching_ThenItUsesTheStatusLinePromptAndMovesTheCursorToTheFirstMatch(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.OpenDetail()
	subject := given_pullRequestCommentProgram(model, &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{
		checkTitle: "CI / test",
		runURL:     "https://github.com/acme/widgets/actions/runs/42",
		body:       strings.Join([]string{"Alpha line", "Target match line", "Omega line"}, "\n"),
	})
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	document := subject.currentPullRequestBuildRunPopupDocument(popupView)
	expectedLineIndex, _ := given_detailDocumentLineContaining(t, document, "Target match line")

	openSearchHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestBuildInfoName, '/')
	actualErr = openSearchHandler(gui, popupView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewSearchName)
	then_statusLineKeyHintsAre(t, gui, "Enter: submit, Escape: cancel")

	searchView, actualErr := gui.View(viewSearchName)
	then_noError(t, actualErr)
	for _, character := range "Target" {
		actualHandled := subject.editSearch(searchView, 0, character, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(character))
		}
	}

	submitHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewSearchName, gocui.KeyEnter)
	actualErr = submitHandler(gui, searchView)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewPullRequestBuildInfoName)
	if actual := subject.pullRequestBuildRunPopup.searchQuery; actual != "Target" {
		t.Fatalf("expected applied popup search query %q, actual %q", "Target", actual)
	}
	if actual := subject.pullRequestBuildRunPopup.viewState.cursor.line; actual != expectedLineIndex {
		t.Fatalf("expected popup cursor line %d after search, actual %d", expectedLineIndex, actual)
	}
	then_statusLineKeyHintsAre(t, gui, "/: search, y: copy, Escape: back")
}

func TestPullRequestBuildRunPopup_GivenSubmittedSearch_WhenPressingNAndN_ThenItMovesToTheNextAndPreviousMatch(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.OpenDetail()
	subject := given_pullRequestCommentProgram(model, &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{
		checkTitle: "CI / test",
		runURL:     "https://github.com/acme/widgets/actions/runs/42",
		body:       strings.Join([]string{"Target first", "middle", "Target second"}, "\n"),
	})
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	document := subject.currentPullRequestBuildRunPopupDocument(popupView)
	firstMatchLineIndex, _ := given_detailDocumentLineContaining(t, document, "Target first")
	secondMatchLineIndex, _ := given_detailDocumentLineContaining(t, document, "Target second")

	openSearchHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestBuildInfoName, '/')
	actualErr = openSearchHandler(gui, popupView)
	then_noError(t, actualErr)
	searchView, actualErr := gui.View(viewSearchName)
	then_noError(t, actualErr)
	for _, character := range "Target" {
		actualHandled := subject.editSearch(searchView, 0, character, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(character))
		}
	}
	submitHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewSearchName, gocui.KeyEnter)
	actualErr = submitHandler(gui, searchView)
	then_noError(t, actualErr)

	if actual := subject.pullRequestBuildRunPopup.viewState.cursor.line; actual != firstMatchLineIndex {
		t.Fatalf("expected popup cursor line %d after submitted search, actual %d", firstMatchLineIndex, actual)
	}

	nextHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestBuildInfoName, 'n')
	actualErr = nextHandler(gui, popupView)
	then_noError(t, actualErr)
	if actual := subject.pullRequestBuildRunPopup.viewState.cursor.line; actual != secondMatchLineIndex {
		t.Fatalf("expected popup cursor line %d after next match, actual %d", secondMatchLineIndex, actual)
	}

	previousHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestBuildInfoName, 'N')
	actualErr = previousHandler(gui, popupView)
	then_noError(t, actualErr)
	if actual := subject.pullRequestBuildRunPopup.viewState.cursor.line; actual != firstMatchLineIndex {
		t.Fatalf("expected popup cursor line %d after previous match, actual %d", firstMatchLineIndex, actual)
	}
}

func TestPullRequestBuildRunPopup_GivenVisible_WhenYankingAVisualSelection_ThenItCopiesTheSelectedTextAndReturnsToNormalMode(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.OpenDetail()
	subject := given_pullRequestCommentProgram(model, &fakePullRequestDetailLoader{})
	clipboardWriter := &fakeClipboardWriter{}
	subject.clipboardWriter = clipboardWriter
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{
		checkTitle: "CI / test",
		runURL:     "https://github.com/acme/widgets/actions/runs/42",
		body:       "Alpha Beta",
	})
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	document := subject.currentPullRequestBuildRunPopupDocument(popupView)
	lineIndex, _ := given_detailDocumentLineContaining(t, document, "Alpha Beta")
	subject.pullRequestBuildRunPopup.viewState.cursor = detailPosition{line: lineIndex, column: 0}
	subject.pullRequestBuildRunPopup.viewState.preferredColumn = 0
	subject.syncPullRequestBuildRunPopupViewState(document, popupView.InnerHeight())
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	visualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestBuildInfoName, 'v')
	actualErr = visualHandler(gui, popupView)
	then_noError(t, actualErr)
	rightHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestBuildInfoName, 'l')
	for range 4 {
		actualErr = rightHandler(gui, popupView)
		then_noError(t, actualErr)
	}
	yankHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestBuildInfoName, 'y')
	actualErr = yankHandler(gui, popupView)
	then_noError(t, actualErr)

	if actual := clipboardWriter.writes; len(actual) != 1 || actual[0] != "Alpha" {
		t.Fatalf("expected clipboard writes %v, actual %v", []string{"Alpha"}, actual)
	}
	if subject.pullRequestBuildRunPopup.viewState.mode != detailNormalMode {
		t.Fatalf("expected popup mode %v after yanking, actual %v", detailNormalMode, subject.pullRequestBuildRunPopup.viewState.mode)
	}
	then_statusLineContains(t, gui, detailYankSuccessMessage)
}

func TestActionsPopup_GivenDetailCursorOnBuildLink_WhenOpening_ThenItShowsTheBuildRunAndJobLogsActions(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/build-run-action",
				State:       "OPEN",
				StatusCheckRollup: []githubcli.PullRequestStatusCheck{{
					Name:         "test",
					WorkflowName: "CI",
					Status:       "COMPLETED",
					Conclusion:   "FAILURE",
					Link:         "https://github.com/acme/widgets/actions/runs/42",
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "CI / test (Failed)")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	expectedBuildRunLabel := actionsPopupLabel(actionsPopupBuildRunIcon, pullRequestBuildRunActionTitle)
	if !strings.Contains(popupView.Buffer(), expectedBuildRunLabel) {
		t.Fatalf("expected popup buffer to contain %q, actual %q", expectedBuildRunLabel, popupView.Buffer())
	}
	expectedBuildLogsLabel := actionsPopupLabel(actionsPopupBuildRunLogsIcon, pullRequestBuildRunLogsActionTitle)
	if !strings.Contains(popupView.Buffer(), expectedBuildLogsLabel) {
		t.Fatalf("expected popup buffer to contain %q, actual %q", expectedBuildLogsLabel, popupView.Buffer())
	}
}

func TestReviewMode_GivenViewOneFocusedOnADescriptionBuild_WhenOpeningActionsPopup_ThenItShowsTheBuildAndLinkActions(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/build-run-action",
				State:       "OPEN",
				StatusCheckRollup: []githubcli.PullRequestStatusCheck{{
					Name:         "test",
					WorkflowName: "CI",
					Status:       "COMPLETED",
					Conclusion:   "FAILURE",
					Link:         "https://github.com/acme/widgets/actions/runs/42",
				}},
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusUserView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "CI / test (Failed)")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	for _, expected := range []string{
		actionsPopupLabel(actionsPopupBuildRunIcon, pullRequestBuildRunActionTitle),
		actionsPopupLabel(actionsPopupBuildRunLogsIcon, pullRequestBuildRunLogsActionTitle),
		actionsPopupLabel(actionsPopupOpenLinkIcon, "Open link under cursor"),
	} {
		if !strings.Contains(popupView.Buffer(), expected) {
			t.Fatalf("expected popup buffer to contain %q, actual %q", expected, popupView.Buffer())
		}
	}
}

func TestActionsPopup_GivenTheCursorOffABuildLink_WhenOpening_ThenItHidesTheBuildRunAndJobLogsActions(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/build-run-action",
				State:       "OPEN",
				StatusCheckRollup: []githubcli.PullRequestStatusCheck{{
					Name:         "test",
					WorkflowName: "CI",
					Status:       "COMPLETED",
					Conclusion:   "FAILURE",
					Link:         "https://github.com/acme/widgets/actions/runs/42",
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Body 42")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), pullRequestBuildRunActionTitle) {
		t.Fatalf("expected popup buffer to hide %q, actual %q", pullRequestBuildRunActionTitle, popupView.Buffer())
	}
	if strings.Contains(popupView.Buffer(), pullRequestBuildRunLogsActionTitle) {
		t.Fatalf("expected popup buffer to hide %q, actual %q", pullRequestBuildRunLogsActionTitle, popupView.Buffer())
	}
}

func TestActionsPopup_GivenBuildRunActionSelected_WhenExecuting_ThenItClosesThePopupShowsSpinnerAndEventuallyOpensTheBuildRunPopup(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		buildRuns: map[string]string{
			"https://github.com/acme/widgets/actions/runs/42": "Run #42\nStatus: completed",
		},
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/build-run-action",
				State:       "OPEN",
				StatusCheckRollup: []githubcli.PullRequestStatusCheck{{
					Name:         "test",
					WorkflowName: "CI",
					Status:       "COMPLETED",
					Conclusion:   "FAILURE",
					Link:         "https://github.com/acme/widgets/actions/runs/42",
				}},
			},
		},
	}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: loader.details["acme/widgets#42"]}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "CI / test (Failed)")
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("build run", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "build run"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_statusLineContains(t, gui, "Running `gh run view 42 -R acme/widgets --verbose`.")
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued build run load, actual %d", len(asyncRunner.runs))
	}

	asyncRunner.runs[0]()

	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Run #42") {
		t.Fatalf("expected build run popup buffer to contain %q, actual %q", "Run #42", popupView.Buffer())
	}
}

func TestSanitizePullRequestBuildRunLog_GivenPrefixedGitHubActionsLogLines_WhenSanitizing_ThenItKeepsOnlyTheLogTextAfterUnknownStep(t *testing.T) {
	actual := sanitizePullRequestBuildRunLog(strings.Join([]string{
		"Test / (RW) (GP) (Back) Test    UNKNOWN STEP    2026-04-24T17:36:05.0135694Z ##[endgroup]",
		"Test / (RW) (GP) (Back) Test    UNKNOWN STEP    2026-04-24T17:36:05.0137572Z Secret source: Actions",
		"plain line",
	}, "\n"))

	expected := strings.Join([]string{
		"2026-04-24T17:36:05.0135694Z ##[endgroup]",
		"2026-04-24T17:36:05.0137572Z Secret source: Actions",
		"plain line",
	}, "\n")
	if actual != expected {
		t.Fatalf("expected sanitized logs %q, actual %q", expected, actual)
	}
}

func TestSanitizePullRequestBuildRunLog_GivenUnknownStepLinesWithoutTimestamps_WhenSanitizing_ThenItStillDropsTheRepeatedPrefixAndUnknownStepLabel(t *testing.T) {
	actual := sanitizePullRequestBuildRunLog(strings.Join([]string{
		"Consumer contract tests    UNKNOWN STEP    ***",
		"Consumer contract tests    UNKNOWN STEP    ",
		"2026-05-02T00:40:33.7759098Z   VAULT_SECRET_CI_GITHUB_APP_GENERIC_ID: ***",
		"2026-05-02T00:40:33.7766286Z   VAULT_SECRET_CI_GITHUB_APP_GENERIC_PEM: ***",
		"Consumer contract tests    UNKNOWN STEP    ***",
		"Consumer contract tests    UNKNOWN STEP    ***",
	}, "\n"))

	expected := strings.Join([]string{
		"***",
		"",
		"2026-05-02T00:40:33.7759098Z   VAULT_SECRET_CI_GITHUB_APP_GENERIC_ID: ***",
		"2026-05-02T00:40:33.7766286Z   VAULT_SECRET_CI_GITHUB_APP_GENERIC_PEM: ***",
		"***",
		"***",
	}, "\n")
	if actual != expected {
		t.Fatalf("expected sanitized logs %q, actual %q", expected, actual)
	}
}

func TestActionsPopup_GivenBuildRunPopupVisible_WhenOpening_ThenItDoesNotShowTheViewJobLogsAction(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{
		checkTitle: "CI / test",
		runURL:     "https://github.com/acme/widgets/actions/runs/42",
		repository: "acme/widgets",
		body:       "Run #42\nStatus: completed",
		jobs:       []githubcli.PullRequestBuildRunJob{{DatabaseID: 1234, Name: "Test job"}},
	})
	then_noError(t, actualErr)

	given_pullRequestBuildRunPopupCursorOnLineContaining(t, gui, subject, "Job: Test job (#1234)")
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
}

func TestActionsPopup_GivenViewJobLogsActionSelectedFromTheBuildOverview_WhenExecuting_ThenItClosesThePopupShowsSpinnerAndOpensASanitizedLargeLogsPopup(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		buildRunJobs: map[string][]githubcli.PullRequestBuildRunJob{
			"https://github.com/acme/widgets/actions/runs/42": {
				{DatabaseID: 999, Name: "lint", URL: "https://github.com/acme/widgets/actions/runs/42/job/999"},
				{DatabaseID: 1234, Name: "test", URL: "https://github.com/acme/widgets/actions/runs/42/job/1234"},
			},
		},
		buildLogs: map[int]string{
			1234: strings.Join([]string{
				"Test / (RW) (GP) (Back) Test    UNKNOWN STEP    2026-04-24T17:36:05.0135694Z ##[endgroup]",
				"Test / (RW) (GP) (Back) Test    UNKNOWN STEP    2026-04-24T17:36:05.0137572Z Secret source: Actions",
			}, "\n"),
		},
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/build-run-action",
				State:       "OPEN",
				StatusCheckRollup: []githubcli.PullRequestStatusCheck{{
					Name:         "test",
					WorkflowName: "CI",
					Status:       "COMPLETED",
					Conclusion:   "FAILURE",
					Link:         "https://github.com/acme/widgets/actions/runs/42",
				}},
			},
		},
	}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: loader.details["acme/widgets#42"]}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "CI / test (Failed)")
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("job logs", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "job logs"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_statusLineContains(t, gui, string(loadingSpinnerFrames[0]))
	then_statusLineContains(t, gui, "Running `gh run view 42 -R acme/widgets --json jobs`.")
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued job logs load, actual %d", len(asyncRunner.runs))
	}

	asyncRunner.runs[0]()

	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Title, "Build logs · test") {
		t.Fatalf("expected popup title to contain %q, actual %q", "Build logs · test", popupView.Title)
	}
	if !strings.Contains(popupView.Buffer(), "2026-04-24T17:36:05.0135694Z ##[endgroup]") {
		t.Fatalf("expected sanitized logs popup to contain %q, actual %q", "2026-04-24T17:36:05.0135694Z ##[endgroup]", popupView.Buffer())
	}
	if strings.Contains(popupView.Buffer(), "Test / (RW) (GP) (Back) Test    UNKNOWN STEP") {
		t.Fatalf("expected sanitized logs popup to hide the raw prefix, actual %q", popupView.Buffer())
	}
	if !reflect.DeepEqual(loader.buildRunJobCalls, []string{"acme/widgets"}) {
		t.Fatalf("expected build run job calls %v, actual %v", []string{"acme/widgets"}, loader.buildRunJobCalls)
	}
	if !reflect.DeepEqual(loader.buildLogCalls, []int{1234}) {
		t.Fatalf("expected build log calls %v, actual %v", []int{1234}, loader.buildLogCalls)
	}
	then_viewOccupiesAtLeastPercentOfScreen(t, gui, viewPullRequestBuildInfoName, 90, 90)
	then_currentViewNameIs(t, gui, viewPullRequestBuildInfoName)
}

func given_pullRequestBuildRunPopupCursorOnLineContaining(t *testing.T, gui *gocui.Gui, subject *Program, segment string) {
	t.Helper()

	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	document := subject.currentPullRequestBuildRunPopupDocument(popupView)
	lineIndex, _ := given_detailDocumentLineContaining(t, document, segment)
	subject.pullRequestBuildRunPopup.viewState.cursor = detailPosition{line: lineIndex, column: 0}
	subject.pullRequestBuildRunPopup.viewState.preferredColumn = 0
	subject.syncPullRequestBuildRunPopupViewState(document, popupView.InnerHeight())
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
}

func then_viewOccupiesAtLeastPercentOfScreen(t *testing.T, gui *gocui.Gui, viewName string, minimumWidthPercent int, minimumHeightPercent int) {
	t.Helper()

	maxX, maxY := gui.Size()
	x0, y0, x1, y1, actualErr := gui.ViewPosition(viewName)
	then_noError(t, actualErr)
	actualWidth := x1 - x0 + 1
	actualHeight := y1 - y0 + 1
	minimumWidth := (maxX * minimumWidthPercent) / 100
	minimumHeight := (maxY * minimumHeightPercent) / 100
	if actualWidth < minimumWidth {
		t.Fatalf("expected view %q width at least %d (%d%% of %d), actual %d", viewName, minimumWidth, minimumWidthPercent, maxX, actualWidth)
	}
	if actualHeight < minimumHeight {
		t.Fatalf("expected view %q height at least %d (%d%% of %d), actual %d", viewName, minimumHeight, minimumHeightPercent, maxY, actualHeight)
	}
}
