package tui

import (
	"reflect"
	"testing"

	"github.com/jesseduffield/gocui"
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

func TestKeybindingSpecs_GivenProgram_WhenListingDetailBindings_ThenDetailViewUsesBracketsEnterAndEscapeVariantsForItsLocalActions(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: '[', handler: subject.previousDetailTab})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: ']', handler: subject.nextDetailTab})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: gocui.KeyEnter, handler: subject.toggleInlineConversationVisibility})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: gocui.KeyEsc, handler: subject.closeDetail})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: gocui.KeyCtrlLsqBracket, handler: subject.closeDetail})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'q', handler: subject.closeDetail})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: '[', handler: subject.previousPullRequestTab})
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

func TestKeybindingSpecs_GivenProgram_WhenListingVerticalNavigationBindings_ThenMainPanesSupportJKAndArrowKeys(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 'j', handler: subject.moveSelectionDown})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyArrowDown, handler: subject.moveSelectionDown})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 'k', handler: subject.moveSelectionUp})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyArrowUp, handler: subject.moveSelectionUp})
	}
}

func TestKeybindingSpecs_GivenProgram_WhenListingEdgeNavigationBindings_ThenSideViewsAndTheActionsPopupSupportGGAndG(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 'g', handler: subject.moveSideSelectionToTop})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 'G', handler: subject.moveSideSelectionToBottom})
	}
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: 'g', handler: subject.moveActionsPopupSelectionToTop})
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

func TestKeybindingSpecs_GivenProgram_WhenListingDetailNavigationBindings_ThenDetailViewSupportsWordAndLineVisualMotions(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'e', handler: subject.moveDetailCursorToWordEnd})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'b', handler: subject.moveDetailCursorToPreviousWord})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'V', handler: subject.enterDetailLineVisualMode})
}

func TestKeybindingSpecs_GivenProgram_WhenListingInlineConversationBindings_ThenDetailViewSupportsZAAsAPrefixToggle(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'z', handler: subject.armInlineConversationTogglePrefix})
}

func TestKeybindingSpecs_GivenProgram_WhenListingHelpBindings_ThenQuestionMarkTogglesThePopupFromAnyMainPaneAndEscapeVariantsCloseIt(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: '?', handler: subject.toggleHelp})
	}
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: gocui.KeyEsc, handler: subject.closeHelp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: gocui.KeyCtrlLsqBracket, handler: subject.closeHelp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: 'q', handler: subject.closeHelp})
}

func TestKeybindingSpecs_GivenProgram_WhenListingDismissBindings_ThenQMirrorsEscapeOutsideTextInputs(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 'q', handler: subject.exitReviewMode})
	}
	then_bindingExists(t, actual, keybindingSpec{viewName: viewActionsPopupName, key: 'q', handler: subject.closeActionsPopup})
	for _, viewName := range []string{viewSearchName, viewActionsPopupSearchName, viewModalEditorName} {
		then_bindingDoesNotExist(t, actual, viewName, 'q')
	}
}

func then_bindingExists(t *testing.T, specs []keybindingSpec, expected keybindingSpec) {
	t.Helper()

	for _, actual := range specs {
		if actual.viewName == expected.viewName && reflect.DeepEqual(actual.key, expected.key) && sameHandler(actual.handler, expected.handler) {
			return
		}
	}

	t.Fatalf("expected binding %+v, actual %+v", expected, specs)
}

func sameHandler(actual func(*gocui.Gui, *gocui.View) error, expected func(*gocui.Gui, *gocui.View) error) bool {
	return reflect.ValueOf(actual).Pointer() == reflect.ValueOf(expected).Pointer()
}
