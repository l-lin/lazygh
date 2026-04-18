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
