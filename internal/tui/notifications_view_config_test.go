package tui

import (
	"testing"

	appconfig "github.com/l-lin/lazygh/internal/config"
)

func TestNotificationLoadPlan_GivenNotificationsViewDisabled_WhenPlanning_ThenItSkipsHydrationAndRefresh(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.notificationLoadPlan()

	if len(actual.messages) != 0 {
		t.Fatalf("expected no notification load messages, actual %v", actual.messages)
	}
	if len(actual.commands) != 0 {
		t.Fatalf("expected no notification load commands, actual %v", actual.commands)
	}
}

func TestDisplayConfig_GivenNotificationsFocus_WhenDisablingNotificationsView_ThenItRecoversToPullRequests(t *testing.T) {
	model := given_model()
	model.FocusNotificationsView()
	subject := NewProgramWithModel(model)

	subject.ApplyDisplayConfig(appconfig.DisplayConfig{})

	if actual := subject.model.Focus(); actual != FocusPullRequestsView {
		t.Fatalf("expected focus to recover to %v, actual %v", FocusPullRequestsView, actual)
	}
	if _, ok := subject.screenState().ViewByNumber(sidePanelNotificationsViewNumber); ok {
		t.Fatal("expected disabled notifications view to be absent from screen state")
	}
}

func TestKeybindingSpecs_GivenNotificationsViewDisabled_WhenListingBindings_ThenItOmitsNotificationsShortcuts(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingDoesNotExist(t, actual, viewUserName, '3')
	then_bindingDoesNotExist(t, actual, viewPullRequestsName, '3')
	then_bindingDoesNotExist(t, actual, viewNotificationsName, 'r')
}

func TestKeybindingSpecs_GivenNotificationsViewEnabled_WhenListingBindings_ThenItKeepsNotificationsShortcuts(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.ApplyDisplayConfig(appconfig.DisplayConfig{NotificationsView: true})

	actual := subject.keybindingSpecs()

	then_bindingExists(t, actual, keybindingSpec{viewName: viewPullRequestsName, key: '3', handler: subject.focusNotificationsView})
	then_bindingExists(t, actual, keybindingSpec{viewName: viewNotificationsName, key: 'r', handler: subject.markNotificationRead})
}
