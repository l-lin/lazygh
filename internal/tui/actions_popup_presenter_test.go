package tui

import "testing"

func TestActionsPopupPresenter_GivenReactionPickerErrorAndSearchFilter_WhenResolvingChromeAndPromptState_ThenItUsesTheSnapshot(t *testing.T) {
	subject := actionsPopupPresenter{
		reactionPickerVisible: true,
		errorMessage:          "rate limited",
		searchQuery:           "smile",
		filteredActionCount:   1,
		totalActionCount:      6,
		searchText:            "joy",
		searchCursor:          2,
	}

	if actual := subject.title(); actual != reactionPickerTitle+" · rate limited" {
		t.Fatalf("expected title %q, actual %q", reactionPickerTitle+" · rate limited", actual)
	}
	if actual := subject.footer(); actual != "1 of 6 reactions" {
		t.Fatalf("expected footer %q, actual %q", "1 of 6 reactions", actual)
	}
	if actual := subject.promptText(); actual != "joy" {
		t.Fatalf("expected prompt text %q, actual %q", "joy", actual)
	}
	if actual := subject.promptCursor(); actual != 2 {
		t.Fatalf("expected prompt cursor %d, actual %d", 2, actual)
	}
}

func TestActionsPopupPresenter_GivenAssigneePickerQueryState_WhenResolvingEmptyMessage_ThenItUsesAssigneeSpecificCopy(t *testing.T) {
	blankQuery := actionsPopupPresenter{assigneePickerVisible: true}
	if actual := blankQuery.emptyMessage(); actual != assigneePickerSearchFooterHint {
		t.Fatalf("expected empty-message hint %q, actual %q", assigneePickerSearchFooterHint, actual)
	}

	filteredQuery := actionsPopupPresenter{assigneePickerVisible: true, searchQuery: "ali"}
	if actual := filteredQuery.emptyMessage(); actual != "No matching assignees." {
		t.Fatalf("expected filtered empty message %q, actual %q", "No matching assignees.", actual)
	}
}
