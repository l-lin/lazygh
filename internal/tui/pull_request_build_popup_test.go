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
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'g', handler: subject.movePullRequestBuildRunPopupCursorToTop})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'G', handler: subject.movePullRequestBuildRunPopupCursorToBottom})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'w', handler: subject.movePullRequestBuildRunPopupCursorToNextWord})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'e', handler: subject.movePullRequestBuildRunPopupCursorToWordEnd})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'b', handler: subject.movePullRequestBuildRunPopupCursorToPreviousWord})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'v', handler: subject.enterPullRequestBuildRunPopupVisualMode})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'V', handler: subject.enterPullRequestBuildRunPopupLineVisualMode})
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

func TestBrowserMode_GivenTheCursorOnANonPendingBuild_WhenPressingEnter_ThenItShowsSpinnerBeforeOpeningTheBuildRunPopup(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		buildRuns: map[string]string{
			"https://github.com/acme/widgets/actions/runs/42": "Run #42\nStatus: completed\nConclusion: failure",
		},
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

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued build run load, actual %d", len(asyncRunner.runs))
	}
	then_viewDoesNotExist(t, gui, viewPullRequestBuildInfoName)
	then_statusLineContains(t, gui, string(loadingSpinnerFrames[0]))
	then_statusLineContains(t, gui, "Running `gh run view 42 -R acme/widgets --verbose`.")
	then_currentViewNameIs(t, gui, viewDetailName)

	asyncRunner.runs[0]()

	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Title, "Build run · CI / test") {
		t.Fatalf("expected popup title to contain %q, actual %q", "Build run · CI / test", popupView.Title)
	}
	for _, expected := range []string{"Run: https://github.com/acme/widgets/actions/runs/42", "Run #42", "Status: completed", "Conclusion: failure"} {
		if !strings.Contains(popupView.Buffer(), expected) {
			t.Fatalf("expected popup buffer to contain %q, actual %q", expected, popupView.Buffer())
		}
	}
	if popupView.InnerHeight() < 10 {
		t.Fatalf("expected a taller build run popup, actual inner height %d", popupView.InnerHeight())
	}
	if !gui.Cursor {
		t.Fatal("expected the cursor to be visible inside the build run popup")
	}
	if !reflect.DeepEqual(loader.buildRunCalls, []string{"acme/widgets"}) {
		t.Fatalf("expected build run calls %v, actual %v", []string{"acme/widgets"}, loader.buildRunCalls)
	}
	then_currentViewNameIs(t, gui, viewPullRequestBuildInfoName)
}

func TestBrowserMode_GivenTheCursorOnAPendingBuild_WhenPressingEnter_ThenItDoesNotStartLoadingOrOpenTheBuildRunPopup(t *testing.T) {
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

func TestActionsPopup_GivenDetailCursorOnBuildLink_WhenOpening_ThenItShowsTheBuildRunAction(t *testing.T) {
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
	if !strings.Contains(popupView.Buffer(), actionsPopupLabel(actionsPopupBuildRunIcon, pullRequestBuildRunActionTitle)) {
		t.Fatalf("expected popup buffer to contain %q, actual %q", actionsPopupLabel(actionsPopupBuildRunIcon, pullRequestBuildRunActionTitle), popupView.Buffer())
	}
}

func TestActionsPopup_GivenTheCursorOffABuildLink_WhenOpening_ThenItHidesTheBuildRunAction(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/build-run-action",
				State:       "OPEN",
				StatusCheckRollup: []githubcli.PullRequestStatusCheck{
					{Name: "test", WorkflowName: "CI", Status: "COMPLETED", Conclusion: "FAILURE", Link: "https://github.com/acme/widgets/actions/runs/42"},
					{Name: "deploy", Status: "IN_PROGRESS"},
				},
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
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "deploy (Pending)")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), pullRequestBuildRunActionTitle) {
		t.Fatalf("expected popup buffer to hide %q, actual %q", pullRequestBuildRunActionTitle, popupView.Buffer())
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
