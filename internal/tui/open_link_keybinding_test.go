package tui

import "testing"

func TestKeybindingSpecs_GivenProgram_WhenListingOpenLinkBindings_ThenXIsAvailableOnlyInTheDetailView(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'x', handler: subject.openLinkUnderCursor})
	then_bindingDoesNotExist(t, actual, viewPullRequestsName, 'x')
	then_bindingDoesNotExist(t, actual, viewUserName, 'x')
}
