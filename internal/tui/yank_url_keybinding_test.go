package tui

import "testing"

func TestKeybindingSpecs_GivenProgram_WhenListingYankBindings_ThenYIsAvailableInTheMainViews(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: 'y', handler: subject.copyPullRequestURL})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'y', handler: subject.copyPullRequestURL})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'y', handler: subject.copyPullRequestURL})
}
