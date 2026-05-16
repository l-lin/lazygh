package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestAssigneePicker_GivenSelectedAssigneesAndCurrentUser_WhenOpeningAndWarmupFinishes_ThenItPlacesMeFirstAndKeepsOnlyThePinnedAssigneesVisible(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.connectedUserLogin = "bob"
	subject.connectedUserName = "Bob"
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	popupView := given_openAssigneePicker(t, gui, subject)
	given_runQueuedAsync(t, asyncRunner, 0)

	if popupView.Title != assigneePickerTitle {
		t.Fatalf("expected popup title %q, actual %q", assigneePickerTitle, popupView.Title)
	}
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), []string{
		"[ ] @me (Bob)",
		"[x] @alice (Alice)",
	})
	if !reflect.DeepEqual(loader.searchAssignableUserCalls, []string{"acme/widgets|"}) {
		t.Fatalf("expected assignable user search calls %v, actual %v", []string{"acme/widgets|"}, loader.searchAssignableUserCalls)
	}
}

func TestActionsPopup_GivenDescriptionDetailAssignPRAction_WhenOpeningTheAssigneePicker_ThenItShowsAWarmupSpinnerFooterHintAndGraphQLStatusUntilTheWarmupFinishes(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.connectedUserLogin = "bob"
	subject.connectedUserName = "Bob"
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_openAssignPullRequestAction(t, gui, subject)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued assignee warmup search, actual %d", len(asyncRunner.runs))
	}
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), []string{
		"[ ] @me (Bob)",
		"[x] @alice (Alice)",
		string(loadingSpinnerFrames[0]) + " Fetching assignees",
	})
	then_actionsPopupFooterHintIsSet(t, gui, assigneePickerSearchFooterHint)
	then_statusLineContains(t, gui, "Running `gh api graphql -F owner=acme -F name=widgets -F first=20`.")

	asyncRunner.runs[0]()

	popupView, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), []string{
		"[ ] @me (Bob)",
		"[x] @alice (Alice)",
	})
	if strings.Contains(popupView.Buffer(), "@charlie") {
		t.Fatalf("expected the warmup search to keep hidden results like %q, actual %q", "@charlie", popupView.Buffer())
	}
}

func TestAssigneePicker_GivenSearchQuery_WhenSearchingLazily_ThenItShowsTheMatchingAssigneeAndKeepsThePinnedOnesVisible(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.assigneePickerSearchDebounceDelay = 0
	subject.connectedUserLogin = "bob"
	subject.connectedUserName = "Bob"
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_ = given_openAssigneePicker(t, gui, subject)
	given_runQueuedAsync(t, asyncRunner, 0)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "char" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}

	if len(asyncRunner.runs) < 2 {
		t.Fatalf("expected a queued debounce search after typing, actual %d", len(asyncRunner.runs))
	}
	given_runQueuedAsync(t, asyncRunner, len(asyncRunner.runs)-1)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	separatorLine := given_actionsPopupSeparatorLine(t, popupView)
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), []string{
		"[ ] @me (Bob)",
		"[x] @alice (Alice)",
		separatorLine,
		"[ ] @charlie (Charlie)",
	})
	if strings.Contains(popupView.Buffer(), "@dora") {
		t.Fatalf("expected the assignee picker to hide %q, actual %q", "@dora", popupView.Buffer())
	}
	moveDownHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyCtrlN)
	then_noError(t, moveDownHandler(gui, searchView))
	then_noError(t, moveDownHandler(gui, searchView))
	popupView, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	matchLineIndex := -1
	for lineIndex := 0; ; lineIndex++ {
		line, ok := popupView.Line(lineIndex)
		if !ok {
			break
		}
		if strings.Contains(line, "@charlie") {
			matchLineIndex = lineIndex
			break
		}
	}
	if matchLineIndex < 0 {
		t.Fatalf("expected a visible assignee line containing %q, actual %q", "@charlie", strings.Join(popupView.BufferLines(), "\n"))
	}
	then_viewLineSegmentHasSearchHighlightBackground(t, gui, viewActionsPopupName, matchLineIndex, "char")
}

func TestAssigneePicker_GivenPinnedAndSearchResultAssignees_WhenRendering_ThenItSeparatesThemWithAContinuousLine(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.assigneePickerSearchDebounceDelay = 0
	subject.connectedUserLogin = "bob"
	subject.connectedUserName = "Bob"
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_ = given_openAssigneePicker(t, gui, subject)
	given_runQueuedAsync(t, asyncRunner, 0)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "char" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}
	given_runQueuedAsync(t, asyncRunner, len(asyncRunner.runs)-1)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	separatorLineIndex := given_viewLineIndexContaining(t, popupView, "────")
	expectedSeparator := strings.Repeat("─", popupView.InnerWidth())
	actualSeparator, ok := popupView.Line(separatorLineIndex)
	if !ok {
		t.Fatalf("expected the assignee picker to render a separator line at index %d", separatorLineIndex)
	}
	if actualSeparator != expectedSeparator {
		t.Fatalf("expected separator line %q, actual %q", expectedSeparator, actualSeparator)
	}
}

func TestAssigneePicker_GivenSearchFailureWrappedWithTheGhCommand_WhenSearching_ThenItKeepsTheCommandOutOfThePopupTitle(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	loader.searchAssignableUserErr = errors.New("run `gh api graphql -F owner=acme -F name=widgets -F first=20 -F search=char`: exit status 1: Field 'isBot' doesn't exist on type 'User'")
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.assigneePickerSearchDebounceDelay = 0
	subject.connectedUserLogin = "bob"
	subject.connectedUserName = "Bob"
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_ = given_openAssigneePicker(t, gui, subject)
	given_runQueuedAsync(t, asyncRunner, 0)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "char" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}
	given_runQueuedAsync(t, asyncRunner, len(asyncRunner.runs)-1)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Title, "gh api graphql") {
		t.Fatalf("expected popup title to hide the gh command, actual %q", popupView.Title)
	}
	if !strings.Contains(popupView.Title, "Field 'isBot' doesn't exist on type 'User'") {
		t.Fatalf("expected popup title to contain the stripped error, actual %q", popupView.Title)
	}
}

func TestAssigneePicker_GivenServerReturnedAssigneesThatAreVisibleButNotLocalStringMatches_WhenMovingToOneAndPressingEnter_ThenItTogglesThatAssignee(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	loader.searchAssignableUsers["acme/widgets|Léo"] = []githubcli.PullRequestAuthor{
		{Login: "dclaraLeo0808", Name: "Clara Léonard"},
		{Login: "felixleopold", Name: "FELIX LEOPOLD"},
		{Login: "Leo-Vrmed", Name: "Leonardo Manca"},
		{Login: "LeoDocto", Name: "Léo Dedier"},
	}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.assigneePickerSearchDebounceDelay = 0
	subject.connectedUserLogin = "bob"
	subject.connectedUserName = "Bob"
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_ = given_openAssigneePicker(t, gui, subject)
	given_runQueuedAsync(t, asyncRunner, 0)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "Léo" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}
	given_runQueuedAsync(t, asyncRunner, len(asyncRunner.runs)-1)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	separatorLine := given_actionsPopupSeparatorLine(t, popupView)
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), []string{
		"[ ] @me (Bob)",
		"[x] @alice (Alice)",
		separatorLine,
		"[ ] @dclaraLeo0808 (Clara Léonard)",
		"[ ] @felixleopold (FELIX LEOPOLD)",
		"[ ] @Leo-Vrmed (Leonardo Manca)",
		"[ ] @LeoDocto (Léo Dedier)",
	})

	moveDownHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyCtrlN)
	executeHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyEnter)
	then_noError(t, moveDownHandler(gui, searchView))
	then_noError(t, moveDownHandler(gui, searchView))
	then_noError(t, executeHandler(gui, searchView))

	popupView, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "[x] @dclaraLeo0808 (Clara Léonard)") {
		t.Fatalf("expected the visible GitHub search result to toggle, actual %q", popupView.Buffer())
	}
}

func TestAssigneePicker_GivenSearchQueryDoesNotMatchViewer_WhenPressingEnterImmediately_ThenItStillTogglesMe(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.assigneePickerSearchDebounceDelay = 0
	subject.connectedUserLogin = "bob"
	subject.connectedUserName = "Bob"
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_ = given_openAssigneePicker(t, gui, subject)
	given_runQueuedAsync(t, asyncRunner, 0)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "char" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}
	given_runQueuedAsync(t, asyncRunner, len(asyncRunner.runs)-1)

	executeHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyEnter)
	then_noError(t, executeHandler(gui, searchView))

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "[x] @me (Bob)") {
		t.Fatalf("expected @me to stay selectable even when the query does not match it, actual %q", popupView.Buffer())
	}
}

func TestAssigneePicker_GivenSelectedSearchResult_WhenSearchingForAnotherAssignee_ThenItKeepsThatSelectedAssigneeVisibleAndSelectable(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	loader.searchAssignableUsers["acme/widgets|dora"] = []githubcli.PullRequestAuthor{{Login: "dora", Name: "Dora"}}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.assigneePickerSearchDebounceDelay = 0
	subject.connectedUserLogin = "bob"
	subject.connectedUserName = "Bob"
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_ = given_openAssigneePicker(t, gui, subject)
	given_runQueuedAsync(t, asyncRunner, 0)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "char" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}
	given_runQueuedAsync(t, asyncRunner, len(asyncRunner.runs)-1)

	moveDownHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyCtrlN)
	executeHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyEnter)
	then_noError(t, moveDownHandler(gui, searchView))
	then_noError(t, moveDownHandler(gui, searchView))
	then_noError(t, executeHandler(gui, searchView))

	if !subject.editActionsPopupSearch(searchView, gocui.KeyCtrlU, 0, gocui.ModNone) {
		t.Fatal("expected ctrl+u to clear the assignee picker query")
	}
	for _, ch := range "dora" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}
	given_runQueuedAsync(t, asyncRunner, len(asyncRunner.runs)-1)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	separatorLine := given_actionsPopupSeparatorLine(t, popupView)
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), []string{
		"[ ] @me (Bob)",
		"[x] @alice (Alice)",
		"[x] @charlie (Charlie)",
		separatorLine,
		"[ ] @dora (Dora)",
	})
	then_noError(t, executeHandler(gui, searchView))

	popupView, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), "[x] @charlie (Charlie)") {
		t.Fatalf("expected the selected assignee to stay selectable while another search is active, actual %q", popupView.Buffer())
	}
}

func TestAssigneePicker_GivenViewerMatchedOnlyByItsHiddenLogin_WhenPressingEnter_ThenItTogglesMe(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	loader.searchAssignableUsers["acme/widgets|l-lin"] = []githubcli.PullRequestAuthor{{Login: "l-lin", Name: "Louis Lin"}}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.assigneePickerSearchDebounceDelay = 0
	subject.connectedUserLogin = "l-lin"
	subject.connectedUserName = "Louis Lin"
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_ = given_openAssigneePicker(t, gui, subject)
	given_runQueuedAsync(t, asyncRunner, 0)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "l-lin" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}
	given_runQueuedAsync(t, asyncRunner, len(asyncRunner.runs)-1)

	executeHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyEnter)
	then_noError(t, executeHandler(gui, searchView))

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "[x] @me (Louis Lin)") {
		t.Fatalf("expected the viewer row to toggle after pressing Enter, actual %q", popupView.Buffer())
	}
}

func TestAssigneePicker_GivenSearchResultSelected_WhenClearingTheQuery_ThenItKeepsTheToggledAssigneeVisible(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.assigneePickerSearchDebounceDelay = 0
	subject.connectedUserLogin = "bob"
	subject.connectedUserName = "Bob"
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_ = given_openAssigneePicker(t, gui, subject)
	given_runQueuedAsync(t, asyncRunner, 0)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "char" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}
	given_runQueuedAsync(t, asyncRunner, len(asyncRunner.runs)-1)

	moveDownHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyCtrlN)
	executeHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyEnter)
	then_noError(t, moveDownHandler(gui, searchView))
	then_noError(t, moveDownHandler(gui, searchView))
	then_noError(t, executeHandler(gui, searchView))
	if !subject.editActionsPopupSearch(searchView, gocui.KeyCtrlU, 0, gocui.ModNone) {
		t.Fatal("expected ctrl+u to clear the assignee picker query")
	}

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_popupBufferContainsOrderedActionLines(t, popupView.Buffer(), []string{
		"[ ] @me (Bob)",
		"[x] @alice (Alice)",
		"[x] @charlie (Charlie)",
	})
}

func TestAssignPullRequest_GivenChangedSelection_WhenSearchingAndSubmittingWithAltEnter_ThenItAppliesTheAssigneeDiffAndRefreshesTheVisibleDetail(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = asyncRunner
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.assigneePickerSearchDebounceDelay = 0
	subject.connectedUserLogin = "bob"
	subject.connectedUserName = "Bob"
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_ = given_openAssigneePicker(t, gui, subject)
	given_runQueuedAsync(t, asyncRunner, 0)

	searchView, actualErr := gui.View(viewActionsPopupSearchName)
	then_noError(t, actualErr)
	for _, ch := range "bob" {
		actualHandled := subject.editActionsPopupSearch(searchView, 0, ch, gocui.ModNone)
		if !actualHandled {
			t.Fatalf("expected typing %q to be handled", string(ch))
		}
	}
	given_runQueuedAsync(t, asyncRunner, len(asyncRunner.runs)-1)

	executeHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyEnter)
	submitHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyAltEnter)
	then_noError(t, executeHandler(gui, searchView))
	then_noError(t, submitHandler(gui, searchView))
	then_currentViewNameIs(t, gui, viewDetailName)

	if !reflect.DeepEqual(loader.updateAssigneeCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected assignee update calls %v, actual %v", []string{"acme/widgets#42"}, loader.updateAssigneeCalls)
	}
	if !reflect.DeepEqual(loader.updateAssigneeAdditions, [][]string{{"bob"}}) {
		t.Fatalf("expected assignee additions %v, actual %v", [][]string{{"bob"}}, loader.updateAssigneeAdditions)
	}
	if len(loader.updateAssigneeRemovals) != 1 || len(loader.updateAssigneeRemovals[0]) != 0 {
		t.Fatalf("expected no assignee removals, actual %v", loader.updateAssigneeRemovals)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "@bob") {
		t.Fatalf("expected detail buffer to contain %q after assigning, actual %q", "@bob", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "@alice") {
		t.Fatalf("expected detail buffer to keep %q after assigning, actual %q", "@alice", detailView.Buffer())
	}
	then_statusLineContains(t, gui, pullRequestAssigneesUpdatedSuccessMessage)
}

func TestAssignPullRequest_GivenPendingSelectionChanges_WhenCanceling_ThenItLeavesThePullRequestUntouched(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.connectedUserLogin = "bob"
	subject.connectedUserName = "Bob"
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

func TestActionsPopup_GivenReviewSearch_WhenMatchingActions_ThenItDoesNotTreatAssignPROnlyAliasesAsOccurrences(t *testing.T) {
	loader := given_pullRequestAssigneeLoader()
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	actions := subject.currentActionsPopupActions()
	actualIndexes := matchingActionsPopupIndexes(actions, "review")
	for _, actualIndex := range actualIndexes {
		if actualIndex < 0 || actualIndex >= len(actions) {
			continue
		}
		if actions[actualIndex].title == assignPullRequestActionTitle {
			t.Fatalf("expected %q to stay out of the review search matches, actual indexes %v", assignPullRequestActionTitle, actualIndexes)
		}
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

func given_runQueuedAsync(t *testing.T, runner *capturingAsyncRunner, index int) {
	t.Helper()

	if index < 0 || index >= len(runner.runs) {
		t.Fatalf("expected queued async run index %d, actual runs %d", index, len(runner.runs))
	}
	runner.runs[index]()
}

func given_actionsPopupSeparatorLine(t *testing.T, popupView *gocui.View) string {
	t.Helper()

	if popupView == nil {
		t.Fatal("expected a popup view")
	}
	return strings.Repeat("─", popupView.InnerWidth())
}

func then_actionsPopupFooterHintIsSet(t *testing.T, gui *gocui.Gui, expected string) {
	t.Helper()

	popupChromeView, actualErr := gui.View(viewActionsPopupChromeName)
	then_noError(t, actualErr)
	if popupChromeView.Footer != expected {
		t.Fatalf("expected popup footer %q, actual %q", expected, popupChromeView.Footer)
	}
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
				{Login: "dora", Name: "Dora"},
			},
		},
		searchAssignableUsers: map[string][]githubcli.PullRequestAuthor{
			"acme/widgets|": {
				{Login: "alice", Name: "Alice"},
				{Login: "bob", Name: "Bob"},
				{Login: "charlie", Name: "Charlie"},
				{Login: "dora", Name: "Dora"},
			},
			"acme/widgets|bob": {
				{Login: "bob", Name: "Bob"},
			},
			"acme/widgets|char": {
				{Login: "charlie", Name: "Charlie"},
			},
		},
	}
}
