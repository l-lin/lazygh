package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestKeybindingSpecs_GivenProgram_WhenListingHorizontalCursorBindings_ThenDetailAndBuildPopupSupportArrowKeys(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: gocui.KeyArrowLeft, handler: subject.moveDetailCursorLeft})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: gocui.KeyArrowRight, handler: subject.moveDetailCursorRight})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: gocui.KeyArrowLeft, handler: subject.movePullRequestBuildRunPopupCursorLeft})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: gocui.KeyArrowRight, handler: subject.movePullRequestBuildRunPopupCursorRight})
}
