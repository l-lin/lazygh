package tui

import (
	"testing"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"github.com/jesseduffield/gocui"
)

func TestKeybindingSpecs_GivenMainPaneSearchOverride_WhenListingBindings_ThenItAppliesToEveryMainPane(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"main": {
			"open_search": {"s"},
		},
	})

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewDetailName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 's', handler: subject.openSearch})
		then_bindingDoesNotExist(t, actual, viewName, '/')
	}
}

func TestKeybindingSpecs_GivenPullRequestsOpenDetailOverride_WhenListingBindings_ThenItReplacesOnlyThatScopedAction(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"pull_requests": {
			"open_detail": {"o"},
		},
	})

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: 'o', handler: subject.openDetail})
	then_bindingDoesNotExist(t, actual, viewPullRequestsName, gocui.KeyEnter)
	then_bindingExists(t, actual, keybindingSpec{viewName: viewUserName, key: gocui.KeyEnter, handler: subject.openDetail})
}

func TestKeybindingSpecs_GivenConflictingPullRequestsOverride_WhenListingBindings_ThenItIgnoresTheBadEntry(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_model(), appconfig.KeymapOverrides{
		"pull_requests": {
			"open_detail": {"y"},
		},
	})

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: gocui.KeyEnter, handler: subject.openDetail})
	actualHandler := given_handlerForBinding(t, actual, viewPullRequestsName, 'y')
	if !sameHandler(actualHandler, subject.copyPullRequestURL) {
		t.Fatalf("expected %q to keep the copy URL handler", "y")
	}
}

func given_programWithKeymapOverrides(model *Model, overrides appconfig.KeymapOverrides) *Program {
	subject := NewProgramWithModel(model)
	subject.ApplyKeymapOverrides(overrides)
	return subject
}
