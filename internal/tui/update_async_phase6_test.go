package tui

import (
	"errors"
	"os"
	"strings"
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestAsyncResultHandlers_GivenRemainingFeatureFiles_WhenInspecting_ThenTheyDoNotMutateStateInsideUiUpdaterClosures(t *testing.T) {
	for _, path := range []string{
		"notification_actions.go",
		"pull_request_stage_merge.go",
		"review_story.go",
		"gh_command_status.go",
		"pull_request_assignee.go",
		"pull_request_build.go",
	} {
		contents, actualErr := os.ReadFile(path)
		then_noError(t, actualErr)

		if strings.Contains(string(contents), "uiUpdater.Apply") {
			t.Fatalf("expected %q to dispatch async result messages instead of mutating state inside uiUpdater closures, actual source:\n%s", path, string(contents))
		}
	}
}

func TestUpdate_GivenMsgActionsPopupAsyncGHCommandFinished_WhenSuccessful_ThenItClearsLoadingAppliesSuccessAndClosesThePopup(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.model.OpenActionsPopup(1)
	subject.actionsPopupWidget.searchEditor = newLineEditor("ready")
	subject.actionsPopupWidget.errorMessage = "stale"
	subject.ghCommandLoadingMessage = "Running `gh pr ready`."

	Update(subject, MsgActionsPopupAsyncGHCommandFinished{Success: actionsPopupAsyncFeedbackSuccess{Message: "done"}})

	if actual := subject.ghCommandLoadingMessage; actual != "" {
		t.Fatalf("expected gh command loading message %q, actual %q", "", actual)
	}
	if subject.model.ActionsPopupVisible() {
		t.Fatalf("expected the actions popup to close after a successful async gh command")
	}
	if subject.actionsPopupWidget.searchEditor != nil {
		t.Fatalf("expected the popup search editor to be cleared after success")
	}
	if actual := subject.actionsPopupWidget.errorMessage; actual != "" {
		t.Fatalf("expected popup error message %q, actual %q", "", actual)
	}
	if actual := subject.feedbackMessage; actual != "done" {
		t.Fatalf("expected feedback %q, actual %q", "done", actual)
	}
}

func TestUpdate_GivenMsgNotificationMutationFinished_WhenFailing_ThenItRestoresTheSnapshotAndClearsLoading(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	snapshot := notificationMutationSnapshot{
		rows:          []NotificationRow{{Item: Item{Title: "before-1"}}, {Item: Item{Title: "before-2"}}},
		selectedIndex: 1,
	}
	subject.model.SetNotificationRows([]NotificationRow{{Item: Item{Title: "optimistic"}}})
	subject.model.SelectNotificationIndex(0)
	subject.notificationsLoading = true
	subject.notificationsLoadingDetailMessage = notificationDoneLoadingMessage

	Update(subject, MsgNotificationMutationFinished{Snapshot: snapshot, SuccessFeedbackMessage: "ignored", Err: errors.New("boom")})

	if subject.notificationsLoading {
		t.Fatalf("expected notifications loading to be cleared after the mutation result")
	}
	if actual := subject.notificationsLoadingDetailMessage; actual != "" {
		t.Fatalf("expected notification loading detail %q, actual %q", "", actual)
	}
	actualRows := subject.model.NotificationRows()
	if len(actualRows) != 2 || actualRows[0].Item.Title != "before-1" || actualRows[1].Item.Title != "before-2" {
		t.Fatalf("expected restored rows %+v, actual %+v", snapshot.rows, actualRows)
	}
	if actual := subject.model.SelectedNotificationIndex(); actual != 1 {
		t.Fatalf("expected restored selected notification index %d, actual %d", 1, actual)
	}
	if !subject.transientErrorPopupVisible() {
		t.Fatalf("expected a failing notification mutation to report an error popup")
	}
}

func TestUpdate_GivenMsgAssigneePickerSearchLoaded_WhenRequestIsCurrent_ThenItStoresResultsAndClearsLoading(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.model.OpenActionsPopup(1)
	subject.model.UpdateActionsPopupSearch("ali", nil)
	subject.actionsPopupWidget.assigneePicker = newAssigneePickerState(pullRequestAssigneePickerTarget{repository: "acme/widgets", number: 42}, "viewer", "Viewer")
	subject.actionsPopupWidget.assigneePicker.searchRequestID = 2
	subject.actionsPopupWidget.assigneePicker.searchLoading = true
	subject.actionsPopupWidget.assigneePicker.searchCommand = "gh api graphql"
	subject.actionsPopupWidget.errorMessage = "stale"

	Update(subject, MsgAssigneePickerSearchLoaded{
		RequestID: 2,
		Query:     "ali",
		Results:   []githubdomain.PullRequestAuthor{{Login: "alice", Name: "Alice"}},
	})

	if subject.actionsPopupWidget.assigneePicker.searchLoading {
		t.Fatalf("expected the assignee picker search loading state to be cleared")
	}
	if actual := subject.actionsPopupWidget.assigneePicker.searchCommand; actual != "" {
		t.Fatalf("expected assignee picker search command %q, actual %q", "", actual)
	}
	if actual := subject.actionsPopupWidget.assigneePicker.searchQuery; actual != "ali" {
		t.Fatalf("expected assignee picker search query %q, actual %q", "ali", actual)
	}
	if actual := subject.actionsPopupWidget.errorMessage; actual != "" {
		t.Fatalf("expected assignee picker error message %q, actual %q", "", actual)
	}
	actualResults := subject.actionsPopupWidget.assigneePicker.searchResults
	if len(actualResults) != 1 || actualResults[0].Login != "alice" {
		t.Fatalf("expected assignee picker search results %+v, actual %+v", []githubdomain.PullRequestAuthor{{Login: "alice", Name: "Alice"}}, actualResults)
	}
	if actual := subject.actionsPopupWidget.assigneePicker.knownCandidates["alice"].Name; actual != "Alice" {
		t.Fatalf("expected known candidate %q, actual %q", "Alice", actual)
	}
}

func TestUpdate_GivenMsgStoryReviewPrepared_WhenSuccessful_ThenItClearsLoadingAndStartsStoryReviewSession(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.storyReviewLoading = true
	summary := githubdomain.PullRequest{Title: "First PR", Number: 42, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}}
	prepared := preparedStoryReview{
		summary:         summary,
		diffData:        reviewDiffData{},
		storyData:       reviewStoryData{Summary: "summary", Chapters: []reviewStoryChapter{{ID: "chapter-1", Title: "Chapter 1"}}},
		pendingReviewID: "review-123",
	}

	Update(subject, MsgStoryReviewPrepared{Prepared: prepared})

	if subject.storyReviewLoading {
		t.Fatalf("expected story review loading to be cleared after the async result")
	}
	if !subject.reviewSession.active {
		t.Fatalf("expected the prepared story review to start a review session")
	}
	if actual := subject.reviewSession.mode; actual != reviewSessionModeStory {
		t.Fatalf("expected review session mode %v, actual %v", reviewSessionModeStory, actual)
	}
	if actual := subject.reviewSession.pendingReviewID; actual != "review-123" {
		t.Fatalf("expected pending review id %q, actual %q", "review-123", actual)
	}
}

func TestUpdate_GivenMsgPullRequestBuildRunLoaded_WhenSuccessful_ThenItClearsLoadingAndOpensThePopup(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.pullRequestBuildRunLoad = &pullRequestBuildRunLoadState{command: "gh run view"}
	target := pullRequestBuildRunTarget{
		popupContent: pullRequestBuildRunPopupContent{
			checkTitle: "CI",
			runURL:     "https://example.com/run/1",
			repository: "acme/widgets",
		},
	}

	Update(subject, MsgPullRequestBuildRunLoaded{
		Target:       target,
		RawRunOutput: "run body",
		Jobs:         []githubdomain.PullRequestBuildRunJob{{Name: "job-1"}},
	})

	if subject.pullRequestBuildRunLoad != nil {
		t.Fatalf("expected build run loading state to be cleared after the async result")
	}
	if subject.pullRequestBuildRunPopup == nil {
		t.Fatalf("expected a successful build run load to open the popup")
	}
	if actual := subject.pullRequestBuildRunPopup.title; !strings.Contains(actual, "CI") {
		t.Fatalf("expected popup title to contain %q, actual %q", "CI", actual)
	}
	if actual := subject.pullRequestBuildRunPopup.body; !strings.Contains(actual, "run body") {
		t.Fatalf("expected popup body to contain %q, actual %q", "run body", actual)
	}
}
