package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestKeybindingSpecs_GivenProgram_WhenListingYankBindings_ThenYIsAvailableInTheMainViews(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: 'y', mod: gocui.ModAlt, handler: subject.copyPullRequestURL})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'y', mod: gocui.ModAlt, handler: subject.copyPullRequestURL})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'y', mod: gocui.ModAlt, handler: subject.copyPullRequestURL})
}
