package tui

import (
	"reflect"
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestKeybindingSpecs_GivenProgram_WhenListingDetailBindings_ThenDetailViewSupportsBracketFallbackForControlLeftBracket(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: '[', handler: subject.closeDetail})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: gocui.KeyEsc, handler: subject.closeDetail})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: '[', handler: subject.previousPullRequestTab})
}

func TestKeybindingSpecs_GivenProgram_WhenListingPagingBindings_ThenControlDAndControlUAreAvailableInAllViews(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyCtrlD, handler: subject.pageDown})
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: gocui.KeyCtrlU, handler: subject.pageUp})
	}
}

func TestKeybindingSpecs_GivenProgram_WhenListingDetailNavigationBindings_ThenDetailViewSupportsJAndKScrolling(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'j', handler: subject.moveSelectionDown})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'k', handler: subject.moveSelectionUp})
}

func TestKeybindingSpecs_GivenProgram_WhenListingHelpBindings_ThenQuestionMarkTogglesThePopupAndEscapeVariantsCloseIt(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: "", key: '?', handler: subject.toggleHelp})
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
