package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
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
	expectedReactionGroups := toDomainReactionGroups([]githubcli.ReactionGroup{{Content: githubcli.ReactionContentEyes, TotalCount: 2}})
	if !reflect.DeepEqual(actual.reactionGroups, expectedReactionGroups) {
		t.Fatalf("expected comment reaction groups %+v, actual %+v", expectedReactionGroups, actual.reactionGroups)
	}
}

func TestSelectedPullRequestReactionActionTarget_GivenCommentsTabCursorOnSubmittedReviewBody_WhenResolving_ThenItUsesTheReviewID(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				ID:     "PR_kwDOA",
				Title:  "First PR",
				Number: 42,
				Body:   "Body 42",
				State:  "OPEN",
				Reviews: []githubcli.PullRequestReview{{
					ID:             "PRR_1",
					Author:         &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					Body:           "Review summary",
					State:          "COMMENTED",
					SubmittedAt:    "2026-06-15T06:54:59Z",
					ReactionGroups: []githubcli.ReactionGroup{{Content: githubcli.ReactionContentRocket, TotalCount: 1}},
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Review summary": "Rendered review summary"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered review summary")

	actual, ok := subject.selectedPullRequestReactionActionTarget()
	if !ok {
		t.Fatal("expected a submitted review reaction target")
	}
	if actual.subjectID != "PRR_1" {
		t.Fatalf("expected subject id %q, actual %q", "PRR_1", actual.subjectID)
	}
	if actual.invalidateDiff {
		t.Fatal("expected submitted review reactions to avoid diff invalidation")
	}
	expectedReactionGroups := toDomainReactionGroups([]githubcli.ReactionGroup{{Content: githubcli.ReactionContentRocket, TotalCount: 1}})
	if !reflect.DeepEqual(actual.reactionGroups, expectedReactionGroups) {
		t.Fatalf("expected review reaction groups %+v, actual %+v", expectedReactionGroups, actual.reactionGroups)
	}
}

func TestActionsPopup_GivenBrowserChangesCursorOnInlineComment_WhenOpening_ThenItShowsAddReaction(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithInlineThreadForReplyTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionDiffWithInlineThreadForReplyTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Inline thread body": "Rendered inline thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_browserChangesDetailFocusForInlineComment(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")

	actualErr := subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), reactionPickerTitle) {
		t.Fatalf("expected the popup to contain %q, actual %q", reactionPickerTitle, popupView.Buffer())
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

func TestSelectedPullRequestReactionRemovalTarget_GivenDescriptionTabCursorOnViewerReactionPill_WhenResolving_ThenItUsesThePullRequestReactionAndContent(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				ID:             "PR_kwDOA",
				Title:          "First PR",
				Number:         42,
				Body:           "Body 42",
				State:          "OPEN",
				ReactionGroups: []githubcli.ReactionGroup{{Content: githubcli.ReactionContentThumbsUp, TotalCount: 1, ViewerHasReacted: true}, {Content: githubcli.ReactionContentEyes, TotalCount: 2}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	given_cursorOnDetailText(t, subject, detailView, given_visibleReactionPillText(githubcli.ReactionGroup{Content: githubcli.ReactionContentThumbsUp, TotalCount: 1, ViewerHasReacted: true}))

	actual, ok := subject.selectedPullRequestReactionRemovalTarget()
	if !ok {
		t.Fatal("expected a pull request reaction removal target")
	}
	if actual.subjectID != "PR_kwDOA" {
		t.Fatalf("expected subject id %q, actual %q", "PR_kwDOA", actual.subjectID)
	}
	if string(actual.content) != string(githubcli.ReactionContentThumbsUp) {
		t.Fatalf("expected reaction content %q, actual %q", githubcli.ReactionContentThumbsUp, actual.content)
	}
}

func TestSelectedPullRequestReactionRemovalTarget_GivenCommentsTabCursorOnViewerReactionPill_WhenResolving_ThenItUsesTheCommentReactionAndContent(t *testing.T) {
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
					ReactionGroups: []githubcli.ReactionGroup{{Content: githubcli.ReactionContentEyes, TotalCount: 2, ViewerHasReacted: true}},
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
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	given_cursorOnDetailText(t, subject, detailView, given_visibleReactionPillText(githubcli.ReactionGroup{Content: githubcli.ReactionContentEyes, TotalCount: 2, ViewerHasReacted: true}))

	actual, ok := subject.selectedPullRequestReactionRemovalTarget()
	if !ok {
		t.Fatal("expected a comment reaction removal target")
	}
	if actual.subjectID != "IC_kwDOA" {
		t.Fatalf("expected subject id %q, actual %q", "IC_kwDOA", actual.subjectID)
	}
	if string(actual.content) != string(githubcli.ReactionContentEyes) {
		t.Fatalf("expected reaction content %q, actual %q", githubcli.ReactionContentEyes, actual.content)
	}
}

func TestSelectedPullRequestReactionRemovalTarget_GivenReviewModeCursorOnViewerReactionPill_WhenResolving_ThenItUsesTheInlineCommentReactionAndContent(t *testing.T) {
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
			ReactionGroups: []githubcli.ReactionGroup{{Content: githubcli.ReactionContentHeart, TotalCount: 1, ViewerHasReacted: true}},
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
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	given_cursorOnDetailText(t, subject, detailView, given_visibleReactionPillText(githubcli.ReactionGroup{Content: githubcli.ReactionContentHeart, TotalCount: 1, ViewerHasReacted: true}))

	actual, ok := subject.selectedPullRequestReactionRemovalTarget()
	if !ok {
		t.Fatal("expected an inline comment reaction removal target")
	}
	if actual.subjectID != "PRRC_1" {
		t.Fatalf("expected subject id %q, actual %q", "PRRC_1", actual.subjectID)
	}
	if string(actual.content) != string(githubcli.ReactionContentHeart) {
		t.Fatalf("expected reaction content %q, actual %q", githubcli.ReactionContentHeart, actual.content)
	}
	if !actual.invalidateDiff {
		t.Fatal("expected inline comment reaction removal to invalidate the diff cache")
	}
}

func TestActionsPopup_GivenDescriptionCursorOnOtherUsersReactionPill_WhenListingActions_ThenItDoesNotShowRemoveReaction(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				ID:             "PR_kwDOA",
				Title:          "First PR",
				Number:         42,
				Body:           "Body 42",
				State:          "OPEN",
				ReactionGroups: []githubcli.ReactionGroup{{Content: githubcli.ReactionContentThumbsUp, TotalCount: 1, ViewerHasReacted: true}, {Content: githubcli.ReactionContentEyes, TotalCount: 2}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	given_cursorOnDetailText(t, subject, detailView, given_visibleReactionPillText(githubcli.ReactionGroup{Content: githubcli.ReactionContentEyes, TotalCount: 2}))

	if given_hasActionTitle(subject.currentActionsPopupActions(), "Remove reaction") {
		t.Fatalf("expected remove reaction to stay hidden when the cursor is on someone else's reaction, actual %v", given_actionTitles(subject.currentActionsPopupActions()))
	}
}

func TestActionsPopup_GivenReviewSearchOnTheCommentsTab_WhenMatchingOccurrences_ThenItSkipsTheAddReactionAction(t *testing.T) {
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
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	actions := subject.currentActionsPopupActions()
	actualIndexes := matchingActionsPopupIndexes(actions, "review")
	for _, actualIndex := range actualIndexes {
		if actualIndex < 0 || actualIndex >= len(actions) {
			continue
		}
		if actions[actualIndex].title == reactionPickerTitle {
			t.Fatalf("expected %q to stay out of the review search matches, actual indexes %v", reactionPickerTitle, actualIndexes)
		}
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
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
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
	actualErr = subject.afterStateChange(gui)
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

func TestAddReaction_GivenGitHubRejectsTheReaction_WhenSubmitting_ThenItShowsATransientErrorPopup(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		addReactionErr: errors.New("reaction refused"),
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {ID: "PR_kwDOA", Title: "First PR", Number: 42, Body: "Body 42", State: "OPEN"},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.asyncRunner = &capturingAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("add reaction", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "add reaction"))
	actualErr = subject.afterStateChange(gui)
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
	if subject.actionsPopupWidget.errorMessage != "" {
		t.Fatalf("expected popup error message to stay empty, actual %q", subject.actionsPopupWidget.errorMessage)
	}
	then_transientErrorPopupContains(t, gui, "reaction refused")
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Title, reactionPickerTitle) {
		t.Fatalf("expected popup title to stay on %q, actual %q", reactionPickerTitle, popupView.Title)
	}
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
	actualErr = subject.afterStateChange(gui)
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
	actualErr = subject.afterStateChange(gui)
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
	actualErr = subject.afterStateChange(gui)
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

func TestAddReaction_GivenCommentsTabCommentReaction_WhenSubmittingOptimistically_ThenItKeepsTheRenderedCommentVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				ID:     "PR_kwDOA",
				Title:  "First PR",
				Number: 42,
				Body:   "Body 42",
				State:  "OPEN",
				Comments: []githubcli.PullRequestComment{{
					ID:        "IC_kwDOA",
					Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					Body:      "General feedback",
					CreatedAt: "2026-04-18T10:00:00Z",
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

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("add reaction", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "add reaction"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager detail refresh before the queued run, actual %v", loader.detailCalls)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered general feedback") {
		t.Fatalf("expected detail buffer to keep the rendered comment, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "👍 1") {
		t.Fatalf("expected detail buffer to contain the optimistic reaction pill, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail buffer to avoid the loading spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}
}

func TestAddReaction_GivenCommentsTabSubmittedReviewReaction_WhenSubmittingOptimistically_ThenItKeepsTheRenderedReviewVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				ID:     "PR_kwDOA",
				Title:  "First PR",
				Number: 42,
				Body:   "Body 42",
				State:  "OPEN",
				Reviews: []githubcli.PullRequestReview{{
					ID:          "PRR_1",
					Author:      &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					Body:        "Review summary",
					State:       "COMMENTED",
					SubmittedAt: "2026-06-15T06:54:59Z",
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Review summary": "Rendered review summary"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered review summary")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	if !given_hasActionTitle(subject.currentActionsPopupActions(), reactionPickerTitle) {
		t.Fatalf("expected actions %v to contain %q", given_actionTitles(subject.currentActionsPopupActions()), reactionPickerTitle)
	}
	subject.model.UpdateActionsPopupSearch("add reaction", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "add reaction"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager detail refresh before the queued run, actual %v", loader.detailCalls)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered review summary") {
		t.Fatalf("expected detail buffer to keep the rendered review, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "👍 1") {
		t.Fatalf("expected detail buffer to contain the optimistic reaction pill, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail buffer to avoid the loading spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}
}

func TestAddReaction_GivenBrowserChangesInlineCommentReaction_WhenSubmittingOptimistically_ThenItKeepsTheRenderedDiffVisibleWhileQueueingBackgroundRefreshes(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithInlineThreadForReplyTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionDiffWithInlineThreadForReplyTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Inline thread body": "Rendered inline thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_browserChangesDetailFocusForInlineComment(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr := subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("add reaction", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "add reaction"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 2 {
		t.Fatalf("expected two queued background refreshes, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager detail refresh before the queued runs, actual %v", loader.detailCalls)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager diff refresh before the queued runs, actual %v", loader.diffCalls)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered inline thread body") {
		t.Fatalf("expected detail buffer to keep the rendered thread body, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "👍 1") {
		t.Fatalf("expected detail buffer to contain the optimistic reaction pill, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Loading pull request diff...") {
		t.Fatalf("expected detail buffer to avoid the diff loading state, actual %q", detailView.Buffer())
	}
	then_statusLineContains(t, gui, pullRequestReactionAddedSuccessMessage)
}

func TestAddReaction_GivenReviewModeInlineCommentReaction_WhenSubmittingOptimistically_ThenItKeepsTheRenderedDiffVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
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

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("add reaction", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "add reaction"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager diff refresh before the queued run, actual %v", loader.diffCalls)
	}

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if strings.Contains(filesView.Buffer(), "Loading file tree...") {
		t.Fatalf("expected files buffer to avoid the loading state, actual %q", filesView.Buffer())
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered thread body") {
		t.Fatalf("expected detail buffer to keep the rendered thread body, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "👍 1") {
		t.Fatalf("expected detail buffer to contain the optimistic reaction pill, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Loading pull request diff...") {
		t.Fatalf("expected detail buffer to avoid the diff loading state, actual %q", detailView.Buffer())
	}
}

func TestRemoveReaction_GivenPullRequestReactionUnderCursor_WhenSubmitting_ThenItRemovesTheReactionRefreshesDetailAndShowsFeedback(t *testing.T) {
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
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	given_cursorOnDetailText(t, subject, detailView, given_visibleReactionPillText(githubcli.ReactionGroup{Content: githubcli.ReactionContentThumbsUp, TotalCount: 1, ViewerHasReacted: true}))
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("remove reaction", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "remove reaction"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.removeReactionSubjectIDs, []string{"PR_kwDOA"}) {
		t.Fatalf("expected reaction subject ids %v, actual %v", []string{"PR_kwDOA"}, loader.removeReactionSubjectIDs)
	}
	if !reflect.DeepEqual(loader.removeReactionContents, []githubcli.ReactionContent{githubcli.ReactionContentThumbsUp}) {
		t.Fatalf("expected reaction contents %v, actual %v", []githubcli.ReactionContent{githubcli.ReactionContentThumbsUp}, loader.removeReactionContents)
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.detailCalls)
	}
	if strings.Contains(detailView.Buffer(), "👍 1") {
		t.Fatalf("expected detail buffer to remove the reaction pill, actual %q", detailView.Buffer())
	}
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_statusLineContains(t, gui, pullRequestReactionRemovedSuccessMessage)
}

func TestRemoveReaction_GivenReviewModeInlineCommentReactionUnderCursor_WhenSubmitting_ThenItRemovesTheReactionRefreshesTheDiffAndShowsFeedback(t *testing.T) {
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
			ReactionGroups: []githubcli.ReactionGroup{{Content: githubcli.ReactionContentHeart, TotalCount: 1, ViewerHasReacted: true}},
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
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	given_cursorOnDetailText(t, subject, detailView, given_visibleReactionPillText(githubcli.ReactionGroup{Content: githubcli.ReactionContentHeart, TotalCount: 1, ViewerHasReacted: true}))
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("remove reaction", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "remove reaction"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.removeReactionSubjectIDs, []string{"PRRC_1"}) {
		t.Fatalf("expected reaction subject ids %v, actual %v", []string{"PRRC_1"}, loader.removeReactionSubjectIDs)
	}
	if !reflect.DeepEqual(loader.removeReactionContents, []githubcli.ReactionContent{githubcli.ReactionContentHeart}) {
		t.Fatalf("expected reaction contents %v, actual %v", []githubcli.ReactionContent{githubcli.ReactionContentHeart}, loader.removeReactionContents)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected diff calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.diffCalls)
	}
	if strings.Contains(detailView.Buffer(), "❤️ 1") {
		t.Fatalf("expected detail buffer to remove the inline reaction pill, actual %q", detailView.Buffer())
	}
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_statusLineContains(t, gui, pullRequestReactionRemovedSuccessMessage)
}

func TestRemoveReaction_GivenCommentsTabCommentReaction_WhenSubmittingOptimistically_ThenItKeepsTheRenderedCommentVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
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
					ReactionGroups: []githubcli.ReactionGroup{{Content: githubcli.ReactionContentEyes, TotalCount: 2, ViewerHasReacted: true}},
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
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	given_cursorOnDetailText(t, subject, detailView, given_visibleReactionPillText(githubcli.ReactionGroup{Content: githubcli.ReactionContentEyes, TotalCount: 2, ViewerHasReacted: true}))

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("remove reaction", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "remove reaction"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager detail refresh before the queued run, actual %v", loader.detailCalls)
	}

	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered general feedback") {
		t.Fatalf("expected detail buffer to keep the rendered comment, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "👀 2") {
		t.Fatalf("expected detail buffer to remove the optimistic reaction pill, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail buffer to avoid the loading spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}
}

func TestRemoveReaction_GivenReviewModeInlineCommentReaction_WhenSubmittingOptimistically_ThenItKeepsTheRenderedDiffVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
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
			ReactionGroups: []githubcli.ReactionGroup{{Content: githubcli.ReactionContentHeart, TotalCount: 1, ViewerHasReacted: true}},
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
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	given_cursorOnDetailText(t, subject, detailView, given_visibleReactionPillText(githubcli.ReactionGroup{Content: githubcli.ReactionContentHeart, TotalCount: 1, ViewerHasReacted: true}))

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("remove reaction", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "remove reaction"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager diff refresh before the queued run, actual %v", loader.diffCalls)
	}

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if strings.Contains(filesView.Buffer(), "Loading file tree...") {
		t.Fatalf("expected files buffer to avoid the loading state, actual %q", filesView.Buffer())
	}

	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered thread body") {
		t.Fatalf("expected detail buffer to keep the rendered thread body, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "❤️ 1") {
		t.Fatalf("expected detail buffer to remove the optimistic reaction pill, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Loading pull request diff...") {
		t.Fatalf("expected detail buffer to avoid the diff loading state, actual %q", detailView.Buffer())
	}
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

func given_visibleReactionPillText(group githubcli.ReactionGroup) string {
	document := newDetailDocumentWithWrap(renderReactionGroup(group), 120, false)
	if len(document.lines) == 0 {
		return ""
	}
	return string(document.lines[0])
}

func given_hasActionTitle(actions []actionsPopupAction, expected string) bool {
	for _, actual := range given_actionTitles(actions) {
		if strings.HasPrefix(actual, expected) {
			return true
		}
	}
	return false
}

func given_actionTitles(actions []actionsPopupAction) []string {
	actual := make([]string, 0, len(actions))
	for _, action := range actions {
		actual = append(actual, action.title)
	}
	return actual
}
