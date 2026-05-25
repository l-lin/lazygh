package tui

import "testing"

func TestFooterPresenter_GivenActiveSearchOnTheSamePane_WhenResolvingPaneFooterState_ThenItHidesTheAppliedSearchSummary(t *testing.T) {
	model := given_model()
	model.StartSearchForTarget(FocusUserView, MyPullRequestsTab)
	model.UpdateSearchDraft("dummy")

	subject := footerPresenter{
		model: model,
		paneSearchSummaries: map[Focus]string{
			FocusUserView: "/dummy (2 matches)",
		},
	}

	actual := subject.paneFooterStateFor(FocusUserView)
	if actual.Visible() {
		t.Fatalf("expected the footer summary to stay hidden while search is active, actual %q", actual.Text())
	}
}

func TestFooterPresenter_GivenFocusedPullRequestsPaneAndResolvedBindings_WhenResolvingStatusLineKeyHints_ThenItUsesTheSnapshotResolver(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.FocusPullRequestsView()

	subject := footerPresenter{
		model:                   model,
		screenState:             model.ScreenState(),
		keyResolver:             newKeybindingLabelResolver(given_footerResolvedActions()),
		actionsPopupActionCount: 1,
	}

	actual := subject.statusLineKeyHintsText()
	expected := "!: help, s/Ctrl+S: search, p: action"
	if actual != expected {
		t.Fatalf("expected status line key hints %q, actual %q", expected, actual)
	}
}

func given_footerResolvedActions() []resolvedKeybindingAction {
	return []resolvedKeybindingAction{
		given_footerResolvedAction(keybindingActionID{scope: keymapScopeMain, action: "toggle_help"}, "!"),
		given_footerResolvedAction(keybindingActionID{scope: keymapScopeMain, action: "open_search"}, "s", "ctrl+s"),
		given_footerResolvedAction(keybindingActionID{scope: keymapScopePullRequests, action: "open_actions_popup"}, "p"),
	}
}

func given_footerResolvedAction(actionID keybindingActionID, labels ...string) resolvedKeybindingAction {
	bindings := make([]configuredKeySequence, 0, len(labels))
	for _, label := range labels {
		bindings = append(bindings, configuredKeySequence{label: label})
	}

	return resolvedKeybindingAction{
		action:     keybindingAction{id: actionID},
		bindings:   bindings,
		overridden: true,
	}
}
