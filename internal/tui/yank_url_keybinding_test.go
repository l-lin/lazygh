package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestKeybindingSpecs_GivenProgram_WhenListingYankBindings_ThenAltShiftYIsAvailableInTheMainViews(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: 'Y', mod: gocui.ModAlt, handler: subject.copyPullRequestURL})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'Y', mod: gocui.ModAlt, handler: subject.copyPullRequestURL})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'Y', mod: gocui.ModAlt, handler: subject.copyPullRequestURL})
}
