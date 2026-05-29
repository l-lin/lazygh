package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestKeybindingSpecs_GivenProgram_WhenListingPullRequestBrowserBindings_ThenAltBIsAvailableInThePullRequestViewsAndBuildPopup(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: 'b', mod: gocui.ModAlt, handler: subject.openPullRequestInBrowserShortcut})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'b', mod: gocui.ModAlt, handler: subject.openPullRequestInBrowserShortcut})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewDetailName, key: 'b', mod: gocui.ModAlt, handler: subject.openPullRequestInBrowserShortcut})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: 'b', mod: gocui.ModAlt, handler: subject.openPullRequestInBrowserShortcut})
}
