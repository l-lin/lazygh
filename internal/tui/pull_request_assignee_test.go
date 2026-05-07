package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestAssigneePicker_GivenSelectedAssigneesAndCurrentUser_WhenOpening_ThenItPlacesAtMeFirstAndKeepsSelectedAssigneesAtTheTop(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	loader.details["acme/widgets#42"] = githubcli.PullRequestDetail{
		Title:       "First PR",
		Number:      42,
		Body:        "Body 42",
		BaseRefName: "main",
		HeadRefName: "feature/assignees",
		State:       "OPEN",
		Assignees: []githubcli.PullRequestAuthor{
			{Login: "alice", Name: "Alice"},
			{Login: "bob", Name: "Bob"},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.connectedUserLogin = "bob"
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	popupView := given_openAssigneePicker(t, gui, subject)

	if popupView.Title != assigneePickerTitle {
		t.Fatalf("expected popup title %q, actual %q", assigneePickerTitle, popupView.Title)
	}
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), []string{
		"[x] @me (Bob)",
		"[x] @alice (Alice)",
		"[ ] @charlie (Charlie)",
	})
	if !reflect.DeepEqual(loader.assignableUserCalls, []string{"acme/widgets"}) {
		t.Fatalf("expected assignable user calls %v, actual %v", []string{"acme/widgets"}, loader.assignableUserCalls)
	}
}

func TestActionsPopup_GivenDescriptionDetailAssignPRAction_WhenLoadingAssignees_ThenItShowsASpinnerUntilTheAsyncLoadFinishes(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: loader.details["acme/widgets#42"]}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_openAssignPullRequestAction(t, gui, subject)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued assignee load, actual %d", len(asyncRunner.runs))
	}
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if popupView.Title != assigneePickerTitle {
		t.Fatalf("expected popup title %q while loading, actual %q", assigneePickerTitle, popupView.Title)
	}
	if actual := strings.TrimSpace(popupView.Buffer()); actual != string(loadingSpinnerFrames[0]) {
		t.Fatalf("expected popup buffer %q while loading, actual %q", string(loadingSpinnerFrames[0]), popupView.Buffer())
	}
	then_statusLineContains(t, gui, "Running `gh api repos/acme/widgets/assignees?per_page=100 --paginate --slurp`.")

	asyncRunner.runs[0]()

	popupView, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), []string{
		"[x] @alice (Alice)",
		"[ ] @bob (Bob)",
		"[ ] @charlie (Charlie)",
	})
}

func TestActionsPopup_GivenCachedAssignableUsers_WhenOpeningTheAssigneePickerAgain_ThenItReusesTheCache(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: loader.details["acme/widgets#42"]}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_openAssignPullRequestAction(t, gui, subject)
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued assignee load, actual %d", len(asyncRunner.runs))
	}
	asyncRunner.runs[0]()

	actualErr := subject.closeActionsPopup(gui, nil)
	then_noError(t, actualErr)
	given_openAssignPullRequestAction(t, gui, subject)

	if !reflect.DeepEqual(loader.assignableUserCalls, []string{"acme/widgets"}) {
		t.Fatalf("expected cached assignee list to avoid reloads, actual calls %v", loader.assignableUserCalls)
	}
	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected cached assignee list to avoid new async runs, actual %d", len(asyncRunner.runs))
	}
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), []string{
		"[x] @alice (Alice)",
		"[ ] @bob (Bob)",
		"[ ] @charlie (Charlie)",
	})
}

func TestAssigneePicker_GivenSearchQuery_WhenFiltering_ThenItShowsMatchingAssigneesOnly(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_ = given_openAssigneePicker(t, gui, subject)
	actualErr := subject.focusActionsPopupSearch(gui, nil)
	then_noError(t, actualErr)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "char" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), []string{"[ ] @charlie (Charlie)"})
}

func TestAssignPullRequest_GivenChangedSelection_WhenTogglingWithEnterAndSubmittingWithAltEnter_ThenItAppliesTheAssigneeDiffAndRefreshesTheVisibleDetail(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	popupView := given_openAssigneePicker(t, gui, subject)
	moveDownHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, 'j')
	enterHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, gocui.KeyEnter)
	submitHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, gocui.KeyAltEnter)

	actualErr := enterHandler(gui, popupView)
	then_noError(t, actualErr)
	actualErr = moveDownHandler(gui, popupView)
	then_noError(t, actualErr)
	actualErr = enterHandler(gui, popupView)
	then_noError(t, actualErr)
	actualErr = submitHandler(gui, popupView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if !reflect.DeepEqual(loader.updateAssigneeCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected assignee update calls %v, actual %v", []string{"acme/widgets#42"}, loader.updateAssigneeCalls)
	}
	if !reflect.DeepEqual(loader.updateAssigneeAdditions, [][]string{{"bob"}}) {
		t.Fatalf("expected assignee additions %v, actual %v", [][]string{{"bob"}}, loader.updateAssigneeAdditions)
	}
	if !reflect.DeepEqual(loader.updateAssigneeRemovals, [][]string{{"alice"}}) {
		t.Fatalf("expected assignee removals %v, actual %v", [][]string{{"alice"}}, loader.updateAssigneeRemovals)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "@bob") {
		t.Fatalf("expected detail buffer to contain %q after assigning, actual %q", "@bob", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "@alice") {
		t.Fatalf("expected detail buffer to drop %q after assigning, actual %q", "@alice", detailView.Buffer())
	}
	then_statusLineContains(t, gui, pullRequestAssigneesUpdatedSuccessMessage)
}

func TestAssignPullRequest_GivenPendingSelectionChanges_WhenCanceling_ThenItLeavesThePullRequestUntouched(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	popupView := given_openAssigneePicker(t, gui, subject)
	enterHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, gocui.KeyEnter)
	closeHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, gocui.KeyEsc)

	actualErr := enterHandler(gui, popupView)
	then_noError(t, actualErr)
	actualErr = closeHandler(gui, popupView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if len(loader.updateAssigneeCalls) != 0 {
		t.Fatalf("expected no assignee update calls, actual %v", loader.updateAssigneeCalls)
	}
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "@alice") {
		t.Fatalf("expected detail buffer to keep %q after canceling, actual %q", "@alice", detailView.Buffer())
	}
}

func TestActionsPopup_GivenBrowserCommentsTab_WhenOpening_ThenItHidesAssignPR(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), assignPullRequestActionTitle) {
		t.Fatalf("expected popup buffer to hide %q, actual %q", assignPullRequestActionTitle, popupView.Buffer())
	}
}

func TestActionsPopup_GivenReviewDiff_WhenOpening_ThenItHidesAssignPR(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	loader := given_pullRequestAssigneeLoader()
	loader.startReviewID = "PRR_pending"
	loader.diffs = map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), assignPullRequestActionTitle) {
		t.Fatalf("expected popup buffer to hide %q, actual %q", assignPullRequestActionTitle, popupView.Buffer())
	}
}

func TestActionsPopup_GivenReviewDescription_WhenOpening_ThenItShowsAssignPR(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	loader := given_pullRequestAssigneeLoader()
	loader.startReviewID = "PRR_pending"
	loader.diffs = map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusUserView(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), assignPullRequestActionTitle) {
		t.Fatalf("expected popup buffer to contain %q, actual %q", assignPullRequestActionTitle, popupView.Buffer())
	}
}

func given_openAssignPullRequestAction(t *testing.T, gui *gocui.Gui, subject *Program) {
	t.Helper()

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("assign", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "assign"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
}

func given_openAssigneePicker(t *testing.T, gui *gocui.Gui, subject *Program) *gocui.View {
	t.Helper()

	given_openAssignPullRequestAction(t, gui, subject)
	actual, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	return actual
}

func given_pullRequestAssigneeLoader() *fakePullRequestDetailLoader {
	return &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/assignees",
				State:       "OPEN",
				Assignees:   []githubcli.PullRequestAuthor{{Login: "alice", Name: "Alice"}},
			},
		},
		assignableUsers: map[string][]githubcli.PullRequestAuthor{
			"acme/widgets": {
				{Login: "alice", Name: "Alice"},
				{Login: "bob", Name: "Bob"},
				{Login: "charlie", Name: "Charlie"},
			},
		},
	}
}
