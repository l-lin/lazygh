package tui

import "testing"

func TestKeybindingSpecs_GivenProgram_WhenListingOpenLinkBindings_ThenXIsAvailableOnlyInTheDetailView(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingKeyExists(t, actual, viewDetailName, 'x')
	then_bindingDoesNotExist(t, actual, viewPullRequestsName, 'x')
	then_bindingDoesNotExist(t, actual, viewUserName, 'x')
}
