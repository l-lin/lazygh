package tui

import (
	"reflect"
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestSelectedPullRequestReactionActionTarget_GivenDescriptionTab_WhenResolving_ThenItUsesThePullRequestReactionTargetID(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {ID: "PR_kwDOA", Title: "First PR", Number: 42, Body: "Body 42", State: "OPEN"},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	actual, ok := subject.selectedPullRequestReactionActionTarget()
	if !ok {
		t.Fatal("expected a pull request reaction target")
	}
	if actual.subjectID != "PR_kwDOA" {
		t.Fatalf("expected subject id %q, actual %q", "PR_kwDOA", actual.subjectID)
	}
	if actual.invalidateDiff {
		t.Fatal("expected pull request reactions to avoid diff invalidation")
	}
}

func TestSelectedPullRequestReactionActionTarget_GivenCommentsTabCursorOnPullRequestComment_WhenResolving_ThenItUsesTheCommentID(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				ID:     "PR_kwDOA",
				Title:  "First PR",
				Number: 42,
				Body:   "Body 42",
				State:  "OPEN",
				Comments: []githubcli.PullRequestComment{{
					ID:             "IC_kwDOA",
					Author:         &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					Body:           "General feedback",
					CreatedAt:      "2026-04-18T10:00:00Z",
					ReactionGroups: []githubcli.ReactionGroup{{Content: githubcli.ReactionContentEyes, TotalCount: 2}},
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"General feedback": "Rendered general feedback"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered general feedback")

	actual, ok := subject.selectedPullRequestReactionActionTarget()
	if !ok {
		t.Fatal("expected a pull request comment reaction target")
	}
	if actual.subjectID != "IC_kwDOA" {
		t.Fatalf("expected subject id %q, actual %q", "IC_kwDOA", actual.subjectID)
	}
	if actual.invalidateDiff {
		t.Fatal("expected pull request comment reactions to avoid diff invalidation")
	}
	if !reflect.DeepEqual(actual.reactionGroups, []githubcli.ReactionGroup{{Content: githubcli.ReactionContentEyes, TotalCount: 2}}) {
		t.Fatalf("expected comment reaction groups %+v, actual %+v", []githubcli.ReactionGroup{{Content: githubcli.ReactionContentEyes, TotalCount: 2}}, actual.reactionGroups)
	}
}

func TestSelectedPullRequestReactionActionTarget_GivenReviewModeCursorOnInlineComment_WhenResolving_ThenItUsesTheInlineCommentID(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			ID:             "PRRC_1",
			Author:         &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
			Body:           "Thread body",
			CreatedAt:      "2026-04-20T10:00:00Z",
			ReactionGroups: []githubcli.ReactionGroup{{Content: githubcli.ReactionContentThumbsUp, TotalCount: 1}},
			DiffHunk:       "@@ -1,2 +1,3 @@\n context\n-old line\n+new line",
		}},
	}}
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": diff,
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered thread body")

	actual, ok := subject.selectedPullRequestReactionActionTarget()
	if !ok {
		t.Fatal("expected an inline comment reaction target")
	}
	if actual.subjectID != "PRRC_1" {
		t.Fatalf("expected subject id %q, actual %q", "PRRC_1", actual.subjectID)
	}
	if !actual.invalidateDiff {
		t.Fatal("expected inline comment reactions to invalidate the diff cache")
	}
}

func TestActionsPopup_GivenPullRequestAddReactionAction_WhenExecuting_ThenItShowsTheFullReactionPicker(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {ID: "PR_kwDOA", Title: "First PR", Number: 42, Body: "Body 42", State: "OPEN"},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("add reaction", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "add reaction"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewActionsPopupName)
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Title, reactionPickerTitle) {
		t.Fatalf("expected popup title to contain %q, actual %q", reactionPickerTitle, popupView.Title)
	}
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), expectedReactionPickerLabels())
}

func TestAddReaction_GivenPullRequestReactionPickerSelection_WhenSubmitting_ThenItAddsTheReactionRefreshesDetailAndShowsFeedback(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {ID: "PR_kwDOA", Title: "First PR", Number: 42, Body: "Body 42", State: "OPEN"},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("add reaction", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "add reaction"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.addReactionSubjectIDs, []string{"PR_kwDOA"}) {
		t.Fatalf("expected reaction subject ids %v, actual %v", []string{"PR_kwDOA"}, loader.addReactionSubjectIDs)
	}
	if !reflect.DeepEqual(loader.addReactionContents, []githubcli.ReactionContent{githubcli.ReactionContentThumbsUp}) {
		t.Fatalf("expected reaction contents %v, actual %v", []githubcli.ReactionContent{githubcli.ReactionContentThumbsUp}, loader.addReactionContents)
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.detailCalls)
	}
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "👍 1") {
		t.Fatalf("expected detail buffer to contain the refreshed reaction group, actual %q", detailView.Buffer())
	}
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_statusLineContains(t, gui, pullRequestReactionAddedSuccessMessage)
}

func TestAddReaction_GivenPullRequestNotificationDetailFocus_WhenSubmitting_ThenItUsesTheNotificationPullRequestIdentity(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/rocket#7":   {ID: "PR_list", Title: "List PR", Number: 7, Body: "Body 7", State: "OPEN"},
			"acme/widgets#42": {ID: "PR_notification", Title: "Notification PR", Number: 42, Body: "Body 42", State: "OPEN"},
		},
	}
	model := NewModel(DefaultSeedData())
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "List PR", Number: 7, Repository: githubcli.Repository{NameWithOwner: "acme/rocket"}, URL: "https://github.com/acme/rocket/pull/7", Body: "Body 7"}),
	})
	model.SetNotificationRows([]NotificationRow{given_pullRequestNotificationRow()})
	model.FocusPullRequestsView()
	subject := given_pullRequestCommentProgram(model, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.focusNotificationsView(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("add reaction", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "add reaction"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.addReactionSubjectIDs, []string{"PR_notification"}) {
		t.Fatalf("expected reaction subject ids %v, actual %v", []string{"PR_notification"}, loader.addReactionSubjectIDs)
	}
	if !reflect.DeepEqual(loader.addReactionContents, []githubcli.ReactionContent{githubcli.ReactionContentThumbsUp}) {
		t.Fatalf("expected reaction contents %v, actual %v", []githubcli.ReactionContent{githubcli.ReactionContentThumbsUp}, loader.addReactionContents)
	}
}

func TestAddReaction_GivenReviewModeInlineCommentReactionPickerSelection_WhenSubmitting_ThenItAddsTheReactionRefreshesTheDiffAndShowsFeedback(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			ID:        "PRRC_1",
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
			Body:      "Thread body",
			CreatedAt: "2026-04-20T10:00:00Z",
			DiffHunk:  "@@ -1,2 +1,3 @@\n context\n-old line\n+new line",
		}},
	}}
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": diff,
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered thread body")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("add reaction", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "add reaction"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.addReactionSubjectIDs, []string{"PRRC_1"}) {
		t.Fatalf("expected reaction subject ids %v, actual %v", []string{"PRRC_1"}, loader.addReactionSubjectIDs)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected diff calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.diffCalls)
	}
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "👍 1") {
		t.Fatalf("expected detail buffer to contain the refreshed inline reaction group, actual %q", detailView.Buffer())
	}
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_statusLineContains(t, gui, pullRequestReactionAddedSuccessMessage)
}

func TestAddReaction_GivenViewerAlreadyAddedTheReaction_WhenSubmitting_ThenItShowsANoOpMessageWithoutCallingGitHub(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				ID:             "PR_kwDOA",
				Title:          "First PR",
				Number:         42,
				Body:           "Body 42",
				State:          "OPEN",
				ReactionGroups: []githubcli.ReactionGroup{{Content: githubcli.ReactionContentThumbsUp, TotalCount: 1, ViewerHasReacted: true}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("add reaction", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "add reaction"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(loader.addReactionSubjectIDs) != 0 {
		t.Fatalf("expected no reaction calls, actual %v", loader.addReactionSubjectIDs)
	}
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_statusLineContains(t, gui, pullRequestReactionAlreadyAddedMessage)
}

func expectedReactionPickerLabels() []string {
	return []string{
		"👍 Thumbs up (+1)",
		"👎 Thumbs down (-1)",
		"😄 Laugh",
		"🎉 Hooray",
		"😕 Confused",
		"❤️ Heart",
		"🚀 Rocket",
		"👀 Eyes",
	}
}
