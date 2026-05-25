package tui

import (
	"reflect"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestBindingsForViews_GivenMultipleViewsAndDefinitions_WhenExpanding_ThenItCreatesOneBindingPerCombinationInOrder(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := bindingsForViews(
		[]string{viewUserName, viewDetailName},
		keybindingDefinition{key: 'x', handler: subject.quit},
		keybindingDefinition{key: gocui.KeyEnter, handler: subject.openDetail},
	)

	expected := []keybindingSpec{
		{viewName: viewUserName, key: 'x', handler: subject.quit},
		{viewName: viewUserName, key: gocui.KeyEnter, handler: subject.openDetail},
		{viewName: viewDetailName, key: 'x', handler: subject.quit},
		{viewName: viewDetailName, key: gocui.KeyEnter, handler: subject.openDetail},
	}
	if len(actual) != len(expected) {
		t.Fatalf("expected %d bindings, actual %d", len(expected), len(actual))
	}
	for index, expectedBinding := range expected {
		actualBinding := actual[index]
		if actualBinding.viewName != expectedBinding.viewName || !reflect.DeepEqual(actualBinding.key, expectedBinding.key) || !sameHandler(actualBinding.handler, expectedBinding.handler) {
			t.Fatalf("expected binding %+v at index %d, actual %+v", expectedBinding, index, actualBinding)
		}
	}
}

func TestKeybindingSpecs_GivenProgram_WhenListingDetailBindings_ThenDetailViewUsesBracketsEnterFoldEscapeAndQForItsLocalActions(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingKeyExists(t, actual, viewDetailName, '[')
	then_bindingKeyExists(t, actual, viewDetailName, ']')
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: gocui.KeyEnter, handler: subject.toggleInlineConversationVisibility})
	then_bindingKeyExists(t, actual, viewDetailName, 'z')
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: gocui.KeyEsc, handler: subject.closeDetail})
	then_bindingDoesNotExist(t, actual, viewDetailName, gocui.KeyCtrlLsqBracket)
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'q', handler: subject.closeDetail})
	then_bindingKeyExists(t, actual, viewPullRequestsName, '[')
}

func TestNewProgram_GivenDefaultSeedData_WhenCreatingTheAppProgram_ThenItStartsOnPullRequestsView(t *testing.T) {
	subject := NewProgram()

	if subject.model.Focus() != FocusPullRequestsView {
		t.Fatalf("expected focus %v, actual %v", FocusPullRequestsView, subject.model.Focus())
	}
}

func TestKeybindingSpecs_GivenProgram_WhenListingPagingBindings_ThenControlDAndControlUAreAvailableInAllViews(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyCtrlD, handler: subject.pageDown})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyCtrlU, handler: subject.pageUp})
	}
}

func TestKeybindingSpecs_GivenProgram_WhenListingVerticalNavigationBindings_ThenMainPanesSupportJKArrowKeysAndShiftJKForViewZero(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 'j', handler: subject.moveSelectionDown})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyArrowDown, handler: subject.moveSelectionDown})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 'k', handler: subject.moveSelectionUp})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyArrowUp, handler: subject.moveSelectionUp})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 'J', handler: subject.moveDetailViewDown})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 'K', handler: subject.moveDetailViewUp})
	}
}

func TestViewZeroMotion_GivenPullRequestsFocus_WhenPressingShiftJAndShiftK_ThenItMovesTheDetailCursorWithoutChangingTheSelection(t *testing.T) {
	model := NewModel(SeedData{PullRequestTabs: []PullRequestTabSeed{{Label: "My PRs", PullRequests: []Item{{Title: "pr-1", Detail: "line 1\nline 2\nline 3"}}}}})
	model.FocusPullRequestsView()
	subject := NewProgramWithModel(model)

	downHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'J')
	actualErr := downHandler(nil, nil)
	then_noError(t, actualErr)
	if subject.model.SelectedPullRequestIndex(MyPullRequestsTab) != 0 {
		t.Fatalf("expected selected pull request index %d, actual %d", 0, subject.model.SelectedPullRequestIndex(MyPullRequestsTab))
	}
	if subject.detailState.viewState.cursor == (detailPosition{}) {
		t.Fatalf("expected the detail cursor to move, actual %+v", subject.detailState.viewState.cursor)
	}

	upHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'K')
	actualErr = upHandler(nil, nil)
	then_noError(t, actualErr)
	if subject.detailState.viewState.cursor != (detailPosition{}) {
		t.Fatalf("expected the detail cursor to return to the top, actual %+v", subject.detailState.viewState.cursor)
	}
}

func TestViewZeroMotion_GivenReviewMode_WhenPressingShiftJAndShiftK_ThenItMovesTheDiffCursorWithoutChangingTheSelectedFile(t *testing.T) {
	summary := githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: buildReviewDiffData(given_reviewSessionPullRequestDiff())}
	subject.startReviewSession(summary, "PRR_shift_jk")
	subject.clampReviewSessionSelection()
	expectedSelectedFileTreeRow := subject.navigationState.reviewSession.selectedFileTreeRow

	downHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'J')
	actualErr := downHandler(nil, nil)
	then_noError(t, actualErr)
	if subject.navigationState.reviewSession.selectedFileTreeRow != expectedSelectedFileTreeRow {
		t.Fatalf("expected selected file tree row %d, actual %d", expectedSelectedFileTreeRow, subject.navigationState.reviewSession.selectedFileTreeRow)
	}
	if subject.detailState.viewState.cursor == (detailPosition{}) {
		t.Fatalf("expected the diff cursor to move, actual %+v", subject.detailState.viewState.cursor)
	}

	upHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'K')
	actualErr = upHandler(nil, nil)
	then_noError(t, actualErr)
	if subject.detailState.viewState.cursor != (detailPosition{}) {
		t.Fatalf("expected the diff cursor to return to the top, actual %+v", subject.detailState.viewState.cursor)
	}
}

func TestKeybindingSpecs_GivenProgram_WhenListingEdgeNavigationBindings_ThenSideViewsAndTheActionsPopupSupportGGAndG(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName} {
		then_bindingKeyExists(t, actual, viewName, 'g')
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 'G', handler: subject.moveSideSelectionToBottom})
	}
	then_bindingKeyExists(t, actual, viewActionsPopupName, 'g')
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: 'G', handler: subject.moveActionsPopupSelectionToBottom})
}

func TestSideViewNavigation_GivenUserFocus_WhenPressingGGOrG_ThenItMovesToTheFirstOrLastVisibleRow(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "user-1", Detail: "User detail 1"}, {Title: "user-2", Detail: "User detail 2"}, {Title: "user-3", Detail: "User detail 3"}}})
	subject := NewProgramWithModel(model)

	bottomHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewUserName, 'G')
	actualErr := bottomHandler(nil, nil)
	then_noError(t, actualErr)
	if subject.model.SelectedUserIndex() != 2 {
		t.Fatalf("expected selected user index %d, actual %d", 2, subject.model.SelectedUserIndex())
	}

	topHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewUserName, 'g')
	actualErr = topHandler(nil, nil)
	then_noError(t, actualErr)
	if subject.model.SelectedUserIndex() != 2 {
		t.Fatalf("expected the first g to arm the motion without moving selection, actual %d", subject.model.SelectedUserIndex())
	}

	actualErr = topHandler(nil, nil)
	then_noError(t, actualErr)
	if subject.model.SelectedUserIndex() != 0 {
		t.Fatalf("expected selected user index %d, actual %d", 0, subject.model.SelectedUserIndex())
	}
}

func TestSideViewNavigation_GivenPullRequestsFocus_WhenPressingGGOrG_ThenItMovesToTheFirstOrLastVisibleRow(t *testing.T) {
	model := NewModel(SeedData{PullRequestTabs: []PullRequestTabSeed{{Label: "My PRs", PullRequests: []Item{{Title: "pr-1", Detail: "PR detail 1"}, {Title: "pr-2", Detail: "PR detail 2"}, {Title: "pr-3", Detail: "PR detail 3"}}}}})
	model.FocusPullRequestsView()
	subject := NewProgramWithModel(model)

	bottomHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'G')
	actualErr := bottomHandler(nil, nil)
	then_noError(t, actualErr)
	if subject.model.SelectedPullRequestIndex(MyPullRequestsTab) != 2 {
		t.Fatalf("expected selected pull request index %d, actual %d", 2, subject.model.SelectedPullRequestIndex(MyPullRequestsTab))
	}

	topHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'g')
	actualErr = topHandler(nil, nil)
	then_noError(t, actualErr)
	if subject.model.SelectedPullRequestIndex(MyPullRequestsTab) != 2 {
		t.Fatalf("expected the first g to arm the motion without moving selection, actual %d", subject.model.SelectedPullRequestIndex(MyPullRequestsTab))
	}

	actualErr = topHandler(nil, nil)
	then_noError(t, actualErr)
	if subject.model.SelectedPullRequestIndex(MyPullRequestsTab) != 0 {
		t.Fatalf("expected selected pull request index %d, actual %d", 0, subject.model.SelectedPullRequestIndex(MyPullRequestsTab))
	}
}

func TestKeybindingSpecs_GivenProgram_WhenListingDetailNavigationBindings_ThenDetailViewSupportsWordWORDAndLineVisualMotions(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'w', handler: subject.moveDetailCursorToNextWord})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'e', handler: subject.moveDetailCursorToWordEnd})
	then_bindingKeyExists(t, actual, viewDetailName, 'b')
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'W', handler: subject.moveDetailCursorToNextBigWord})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'E', handler: subject.moveDetailCursorToBigWordEnd})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'B', handler: subject.moveDetailCursorToPreviousBigWord})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'y', handler: subject.startDetailYank})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'V', handler: subject.enterDetailLineVisualMode})
}

func TestKeybindingSpecs_GivenProgram_WhenListingInlineConversationBindings_ThenDetailViewSupportsZAAsAPrefixToggle(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingKeyExists(t, actual, viewDetailName, 'z')
}

func TestKeybindingSpecs_GivenProgram_WhenListingHelpBindings_ThenQuestionMarkTogglesThePopupFromAnyMainPaneAndEscapeAndQCloseIt(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: '?', handler: subject.toggleHelp})
	}
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: gocui.KeyEsc, handler: subject.closeHelp})
	then_bindingDoesNotExist(t, actual, viewHelpName, gocui.KeyCtrlLsqBracket)
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: 'q', handler: subject.closeHelp})
}

func TestKeybindingSpecs_GivenProgram_WhenListingDismissBindings_ThenQMirrorsEscapeOutsideTextInputsWithoutControlBracketAliases(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 'q', handler: subject.exitReviewMode})
		then_bindingDoesNotExist(t, actual, viewName, gocui.KeyCtrlLsqBracket)
	}
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: 'q', handler: subject.closeActionsPopup})
	then_bindingDoesNotExist(t, actual, viewActionsPopupName, gocui.KeyCtrlLsqBracket)
	for _, viewName := range []string{viewSearchName, viewActionsPopupSearchName, viewModalEditorName} {
		then_bindingDoesNotExist(t, actual, viewName, 'q')
	}
}

func then_bindingExists(t *testing.T, specs []keybindingSpec, expected keybindingSpec) {
	t.Helper()

	for _, actual := range specs {
		if actual.viewName == expected.viewName && reflect.DeepEqual(actual.key, expected.key) && actual.mod == expected.mod && sameHandler(actual.handler, expected.handler) {
			return
		}
	}

	t.Fatalf("expected binding %+v, actual %+v", expected, specs)
}

func then_bindingKeyExists(t *testing.T, specs []keybindingSpec, expectedView string, expectedKey any) {
	t.Helper()

	for _, actual := range specs {
		if actual.viewName == expectedView && reflect.DeepEqual(actual.key, expectedKey) && actual.mod == gocui.ModNone {
			return
		}
	}

	t.Fatalf("expected a binding for view %q and key %v, actual %+v", expectedView, expectedKey, specs)
}

func sameHandler(actual func(*gocui.Gui, *gocui.View) error, expected func(*gocui.Gui, *gocui.View) error) bool {
	return reflect.ValueOf(actual).Pointer() == reflect.ValueOf(expected).Pointer()
}
