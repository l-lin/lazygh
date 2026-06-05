package tui

import "testing"

func TestHelpPresenter_GivenReviewFilesFocusAndConfiguredReviewMotionOverrides_WhenListingLocalHelpEntries_ThenItUsesTheSnapshotResolver(t *testing.T) {
	subject := helpPresenter{
		actionContext: ActionContext{
			Mode:       ScreenModeReview,
			ActiveView: ViewState{Focus: FocusPullRequestsView},
		},
		keyResolver: newKeybindingLabelResolver([]resolvedKeybindingAction{
			given_resolvedKeybindingAction(keymapScopeReview, "previous_file", "g["),
			given_resolvedKeybindingAction(keymapScopeReview, "next_file", "g]"),
			given_resolvedKeybindingAction(keymapScopeReview, "previous_comment", "gh"),
			given_resolvedKeybindingAction(keymapScopeReview, "next_comment", "gl"),
			given_resolvedKeybindingAction(keymapScopeReview, "previous_unresolved_comment", "gH"),
			given_resolvedKeybindingAction(keymapScopeReview, "next_unresolved_comment", "gL"),
		}),
	}

	actual := subject.localHelpEntries()

	then_helpEntriesContainKey(t, actual, "Previous/next file", "g[/g]")
	then_helpEntriesContainKey(t, actual, "Previous/next comment", "gh/gl")
	then_helpEntriesContainKey(t, actual, "Previous/next unresolved comment", "gH/gL")
}

func TestHelpPresenter_GivenPullRequestChangesDetailContextWithInlineReplyAvailable_WhenListingLocalHelpEntries_ThenItShowsInlineCommentEntries(t *testing.T) {
	subject := helpPresenter{
		actionContext: ActionContext{
			Mode:            ScreenModeBrowser,
			ActiveView:      ViewState{Focus: FocusDetailView},
			MainView:        MainViewResolver{ContentKind: MainContentKindPullRequestDetail},
			ActiveDetailTab: ChangesDetailTab,
		},
		keyResolver:                 newKeybindingLabelResolver(nil),
		inlineCommentReplyAvailable: true,
	}

	actual := subject.localHelpEntries()

	then_helpEntriesContainKey(t, actual, "Add inline comment", "c")
	then_helpEntriesContainKey(t, actual, "Reply to inline comment", "r")
}

func given_resolvedKeybindingAction(scope string, action string, label string) resolvedKeybindingAction {
	return resolvedKeybindingAction{
		action: keybindingAction{id: keybindingActionID{scope: scope, action: action}},
		bindings: []configuredKeySequence{{
			label: label,
		}},
		overridden: true,
	}
}

func then_helpEntriesContainKey(t *testing.T, entries []helpEntry, description string, expected string) {
	t.Helper()

	for _, entry := range entries {
		if entry.Description != description {
			continue
		}
		if entry.Key != expected {
			t.Fatalf("expected key %q for %q, actual %q", expected, description, entry.Key)
		}
		return
	}

	t.Fatalf("expected help entry %q in %+v", description, entries)
}
