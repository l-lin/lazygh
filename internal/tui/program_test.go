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

func TestKeybindingSpecs_GivenProgram_WhenListingDetailBindings_ThenDetailViewUsesBracketsForItsOwnTabsAndEscapeVariantsToClose(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: '[', handler: subject.previousDetailTab})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: ']', handler: subject.nextDetailTab})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: gocui.KeyEsc, handler: subject.closeDetail})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: gocui.KeyCtrlLsqBracket, handler: subject.closeDetail})
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

func TestKeybindingSpecs_GivenProgram_WhenListingDetailNavigationBindings_ThenDetailViewSupportsWordAndLineVisualMotions(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'j', handler: subject.moveSelectionDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'k', handler: subject.moveSelectionUp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'e', handler: subject.moveDetailCursorToWordEnd})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'b', handler: subject.moveDetailCursorToPreviousWord})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'V', handler: subject.enterDetailLineVisualMode})
}

func TestKeybindingSpecs_GivenProgram_WhenListingHelpBindings_ThenQuestionMarkTogglesThePopupFromAnyMainPaneAndEscapeVariantsCloseIt(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: '?', handler: subject.toggleHelp})
	}
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: gocui.KeyEsc, handler: subject.closeHelp})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewHelpName, key: gocui.KeyCtrlLsqBracket, handler: subject.closeHelp})
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
