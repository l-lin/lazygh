package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestRefreshKeybinding_GivenProgram_WhenListingBindings_ThenAltRIsAvailableInEveryMainPane(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewNotificationsName, viewDetailName} {
		then_bindingExists(t, actual, keybindingSpec{viewName: viewName, key: 'r', mod: gocui.ModAlt, handler: subject.refreshActiveView})
	}
}

func TestRefreshActiveView_GivenConnectedUserFocus_WhenPressingAltR_ThenItDoesNothing(t *testing.T) {
	subject := NewProgramWithModel(NewModel(DefaultSeedData()))
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	userView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	expected := userView.Buffer()

	handler := given_handlerForBindingWithModifier(t, subject.keybindingSpecs(), viewUserName, 'r', gocui.ModAlt)
	actualErr = handler(gui, userView)
	then_noError(t, actualErr)

	userView, actualErr = gui.View(viewUserName)
	then_noError(t, actualErr)
	if actual := userView.Buffer(); actual != expected {
		t.Fatalf("expected connected user buffer %q, actual %q", expected, actual)
	}
	then_statusLineIs(t, gui, "")
}

func TestRefreshActiveView_GivenUserBackedDetailFocus_WhenPressingAltR_ThenItDoesNothing(t *testing.T) {
	model := NewModel(DefaultSeedData())
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	expected := detailView.Buffer()

	handler := given_handlerForBindingWithModifier(t, subject.keybindingSpecs(), viewDetailName, 'r', gocui.ModAlt)
	actualErr = handler(gui, detailView)
	then_noError(t, actualErr)

	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if actual := detailView.Buffer(); actual != expected {
		t.Fatalf("expected user detail buffer %q, actual %q", expected, actual)
	}
	then_statusLineIs(t, gui, "")
}

func TestRefreshActiveView_GivenPullRequestsFocus_WhenPressingAltR_ThenItReloadsTheActivePullRequestList(t *testing.T) {
	loader := &fakePullRequestDetailLoader{myPullRequests: []githubcli.PullRequest{{
		Title:      "Refreshed list PR",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		URL:        "https://github.com/acme/widgets/pull/42",
		State:      "OPEN",
	}}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), "First PR") {
		t.Fatalf("expected pull requests buffer to contain %q before refreshing, actual %q", "First PR", pullRequestsView.Buffer())
	}

	handler := given_handlerForBindingWithModifier(t, subject.keybindingSpecs(), viewPullRequestsName, 'r', gocui.ModAlt)
	actualErr = handler(gui, pullRequestsView)
	then_noError(t, actualErr)

	if len(loader.listPullRequestCommands) != 1 {
		t.Fatalf("expected one pull request list refresh call, actual %d", len(loader.listPullRequestCommands))
	}

	pullRequestsView, actualErr = gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), "Refreshed list PR") {
		t.Fatalf("expected pull requests buffer to contain %q after refreshing, actual %q", "Refreshed list PR", pullRequestsView.Buffer())
	}
	if strings.Contains(pullRequestsView.Buffer(), "First PR") {
		t.Fatalf("expected pull requests buffer to drop %q after refreshing, actual %q", "First PR", pullRequestsView.Buffer())
	}
	then_statusLineContains(t, gui, pullRequestListRefreshSuccessMessage)
}

func TestRefreshActiveView_GivenNotificationsFocus_WhenPressingAltR_ThenItReloadsTheNotificationsList(t *testing.T) {
	initialNotifications := []githubcli.Notification{given_notificationValue(t, given_pullRequestNotificationRow())}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), initialNotifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	notificationsView, actualErr := gui.View(viewNotificationsName)
	then_noError(t, actualErr)
	if !strings.Contains(notificationsView.Buffer(), "Add notifications") {
		t.Fatalf("expected notifications buffer to contain %q before refreshing, actual %q", "Add notifications", notificationsView.Buffer())
	}

	loader.notifications = []githubcli.Notification{given_notificationValue(t, given_issueNotificationRow())}
	handler := given_handlerForBindingWithModifier(t, subject.keybindingSpecs(), viewNotificationsName, 'r', gocui.ModAlt)
	actualErr = handler(gui, notificationsView)
	then_noError(t, actualErr)

	notificationsView, actualErr = gui.View(viewNotificationsName)
	then_noError(t, actualErr)
	if !strings.Contains(notificationsView.Buffer(), "Support notifications in issue detail") {
		t.Fatalf("expected notifications buffer to contain %q after refreshing, actual %q", "Support notifications in issue detail", notificationsView.Buffer())
	}
	if strings.Contains(notificationsView.Buffer(), "Add notifications") {
		t.Fatalf("expected notifications buffer to drop %q after refreshing, actual %q", "Add notifications", notificationsView.Buffer())
	}
	then_statusLineContains(t, gui, notificationsRefreshSuccessMessage)
}

func TestRefreshActiveView_GivenNotificationsFocus_WhenTheReloadFails_ThenItShowsATransientErrorPopup(t *testing.T) {
	initialNotifications := []githubcli.Notification{given_notificationValue(t, given_pullRequestNotificationRow())}
	loader := &fakePullRequestDetailLoader{
		notifications:    append([]githubcli.Notification(nil), initialNotifications...),
		notificationsErr: errors.New("notification refresh refused"),
	}
	subject := given_notificationActionProgram(loader.notifications, loader)
	subject.timingState.transientErrorPopupDuration = 0
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	notificationsView, actualErr := gui.View(viewNotificationsName)
	then_noError(t, actualErr)
	handler := given_handlerForBindingWithModifier(t, subject.keybindingSpecs(), viewNotificationsName, 'r', gocui.ModAlt)
	actualErr = handler(gui, notificationsView)
	then_noError(t, actualErr)

	then_statusLineDoesNotContain(t, gui, notificationsRefreshSuccessMessage)
	then_transientErrorPopupContains(t, gui, "notification refresh refused")
	if !strings.Contains(notificationsView.Buffer(), "Add notifications") {
		t.Fatalf("expected notifications buffer to preserve %q after refresh failure, actual %q", "Add notifications", notificationsView.Buffer())
	}
}

func TestRefreshActiveView_GivenPullRequestDetailFocus_WhenPressingAltR_ThenItReloadsTheSelectedPullRequest(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		myPullRequests: []githubcli.PullRequest{{
			Title:      "Refreshed PR",
			Number:     42,
			Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
			URL:        "https://github.com/acme/widgets/pull/42",
			Body:       "Refreshed summary body",
			State:      "OPEN",
		}},
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "Refreshed PR",
				Number:      42,
				Body:        "Refreshed body",
				BaseRefName: "main",
				HeadRefName: "refresh-branch",
				State:       "OPEN",
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{
		Title:       "Old PR",
		Number:      42,
		Body:        "Old body",
		BaseRefName: "main",
		HeadRefName: "old-branch",
		State:       "OPEN",
	})}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Old body") {
		t.Fatalf("expected detail buffer to contain %q before refreshing, actual %q", "Old body", detailView.Buffer())
	}

	handler := given_handlerForBindingWithModifier(t, subject.keybindingSpecs(), viewDetailName, 'r', gocui.ModAlt)
	actualErr = handler(gui, detailView)
	then_noError(t, actualErr)

	if len(loader.detailCalls) != 1 || loader.detailCalls[0] != "acme/widgets#42" {
		t.Fatalf("expected detail refresh calls %v, actual %v", []string{"acme/widgets#42"}, loader.detailCalls)
	}

	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Refreshed body") {
		t.Fatalf("expected detail buffer to contain %q after refreshing, actual %q", "Refreshed body", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Old body") {
		t.Fatalf("expected detail buffer to drop %q after refreshing, actual %q", "Old body", detailView.Buffer())
	}
	then_statusLineContains(t, gui, pullRequestRefreshSuccessMessage)
}

func TestRefreshActiveView_GivenReviewDiffFocus_WhenPressingAltR_ThenItReloadsTheActivePullRequestDetailAndDiff(t *testing.T) {
	staleDiff := given_reviewSessionPullRequestDiff()
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "Stale PR",
				Number:      42,
				Body:        "Stale body",
				BaseRefName: "main",
				HeadRefName: "feature-old",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": staleDiff,
		},
	}
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

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "new line") {
		t.Fatalf("expected detail buffer to contain %q before refreshing, actual %q", "new line", detailView.Buffer())
	}

	loader.details["acme/widgets#42"] = githubcli.PullRequestDetail{
		Title:       "Refreshed PR",
		Number:      42,
		Body:        "Refreshed body",
		BaseRefName: "main",
		HeadRefName: "feature-refresh",
		State:       "OPEN",
	}
	refreshedDiff := staleDiff
	refreshedDiff.UnifiedDiff = strings.ReplaceAll(refreshedDiff.UnifiedDiff, "+new line", "+fresh line")
	loader.diffs["acme/widgets#42"] = refreshedDiff

	handler := given_handlerForBindingWithModifier(t, subject.keybindingSpecs(), viewDetailName, 'r', gocui.ModAlt)
	actualErr = handler(gui, detailView)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.detailCalls)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected diff calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.diffCalls)
	}

	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "fresh line") {
		t.Fatalf("expected detail buffer to contain %q after refreshing, actual %q", "fresh line", detailView.Buffer())
	}
	then_statusLineContains(t, gui, pullRequestRefreshSuccessMessage)
}

func TestRefreshActiveView_GivenPullRequestsFocus_WhenPressingAltRAndTheListReloadIsAsync_ThenItShowsTheCommandUntilTheReloadFinishes(t *testing.T) {
	loader := &fakePullRequestDetailLoader{myPullRequests: []githubcli.PullRequest{{
		Title:      "Refreshed list PR",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		URL:        "https://github.com/acme/widgets/pull/42",
		State:      "OPEN",
	}}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Cached body", State: "OPEN"})}
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	handler := given_handlerForBindingWithModifier(t, subject.keybindingSpecs(), viewPullRequestsName, 'r', gocui.ModAlt)
	actualErr = handler(gui, pullRequestsView)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued pull request list refresh, actual %d", len(asyncRunner.runs))
	}
	if len(loader.listPullRequestCommands) != 0 {
		t.Fatalf("expected the pull request list command to wait for the queued run, actual %v", loader.listPullRequestCommands)
	}
	then_statusLineContains(t, gui, string(loadingSpinnerFrames[0]))
	then_statusLineContains(t, gui, myPullRequestsLoadingDetail)
	then_statusLineDoesNotContain(t, gui, pullRequestListRefreshSuccessMessage)

	given_runQueuedAsync(t, asyncRunner, 0)

	if len(loader.listPullRequestCommands) != 1 {
		t.Fatalf("expected one pull request list refresh call after running the queue, actual %d", len(loader.listPullRequestCommands))
	}
	then_statusLineContains(t, gui, pullRequestListRefreshSuccessMessage)
}

func TestRefreshActiveView_GivenNotificationsFocus_WhenPressingAltRAndTheReloadIsAsync_ThenItShowsTheCommandUntilTheReloadFinishes(t *testing.T) {
	initialNotifications := []githubcli.Notification{given_notificationValue(t, given_pullRequestNotificationRow())}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), initialNotifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{Title: "Add notifications", Number: 42, Body: "Cached body", State: "OPEN"})}
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	notificationsView, actualErr := gui.View(viewNotificationsName)
	then_noError(t, actualErr)
	loader.notifications = []githubcli.Notification{given_notificationValue(t, given_issueNotificationRow())}
	handler := given_handlerForBindingWithModifier(t, subject.keybindingSpecs(), viewNotificationsName, 'r', gocui.ModAlt)
	actualErr = handler(gui, notificationsView)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued notification refresh, actual %d", len(asyncRunner.runs))
	}
	then_statusLineContains(t, gui, string(loadingSpinnerFrames[0]))
	then_statusLineContains(t, gui, notificationsLoadingDetail)
	then_statusLineDoesNotContain(t, gui, notificationsRefreshSuccessMessage)

	given_runQueuedAsync(t, asyncRunner, 0)

	then_statusLineContains(t, gui, notificationsRefreshSuccessMessage)
}

func TestRefreshActiveView_GivenPullRequestDetailFocus_WhenPressingAltRAndTheRefreshIsAsync_ThenItKeepsShowingCommandsUntilTheListAndDetailReloadFinish(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		myPullRequests: []githubcli.PullRequest{{
			Title:      "Refreshed PR",
			Number:     42,
			Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
			URL:        "https://github.com/acme/widgets/pull/42",
			Body:       "Refreshed summary body",
			State:      "OPEN",
		}},
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "Refreshed PR",
				Number:      42,
				Body:        "Refreshed body",
				BaseRefName: "main",
				HeadRefName: "refresh-branch",
				State:       "OPEN",
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{
		Title:       "Old PR",
		Number:      42,
		Body:        "Old body",
		BaseRefName: "main",
		HeadRefName: "old-branch",
		State:       "OPEN",
	})}
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	handler := given_handlerForBindingWithModifier(t, subject.keybindingSpecs(), viewDetailName, 'r', gocui.ModAlt)
	actualErr = handler(gui, detailView)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 2 {
		t.Fatalf("expected one queued list refresh and one queued detail refresh, actual %d", len(asyncRunner.runs))
	}
	if len(loader.listPullRequestCommands) != 0 {
		t.Fatalf("expected the pull request list command to wait for the queued run, actual %v", loader.listPullRequestCommands)
	}
	if len(loader.detailCalls) != 0 {
		t.Fatalf("expected the pull request detail call to wait for the queued run, actual %v", loader.detailCalls)
	}
	then_statusLineContains(t, gui, string(loadingSpinnerFrames[0]))
	then_statusLineContains(t, gui, subject.selectedPullRequestDetailLoadingStatus())
	then_statusLineDoesNotContain(t, gui, pullRequestRefreshSuccessMessage)

	given_runQueuedAsync(t, asyncRunner, 0)

	then_statusLineContains(t, gui, string(loadingSpinnerFrames[0]))
	then_statusLineContains(t, gui, "Running `gh")
	then_statusLineDoesNotContain(t, gui, pullRequestRefreshSuccessMessage)

	given_runQueuedAsync(t, asyncRunner, 1)

	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected detail refresh calls %v, actual %v", []string{"acme/widgets#42"}, loader.detailCalls)
	}
	if len(loader.listPullRequestCommands) != 1 {
		t.Fatalf("expected one pull request list refresh call, actual %d", len(loader.listPullRequestCommands))
	}
	then_statusLineContains(t, gui, pullRequestRefreshSuccessMessage)
}

func TestRefreshActiveView_GivenReviewDiffFocus_WhenPressingAltRAndTheRefreshIsAsync_ThenItShowsTheDiffCommandUntilTheRefreshFinishes(t *testing.T) {
	staleDiff := given_reviewSessionPullRequestDiff()
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "Stale PR",
				Number:      42,
				Body:        "Stale body",
				BaseRefName: "main",
				HeadRefName: "feature-old",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": staleDiff,
		},
	}
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

	loader.details["acme/widgets#42"] = githubcli.PullRequestDetail{
		Title:       "Refreshed PR",
		Number:      42,
		Body:        "Refreshed body",
		BaseRefName: "main",
		HeadRefName: "feature-refresh",
		State:       "OPEN",
	}
	refreshedDiff := staleDiff
	refreshedDiff.UnifiedDiff = strings.ReplaceAll(refreshedDiff.UnifiedDiff, "+new line", "+fresh line")
	loader.diffs["acme/widgets#42"] = refreshedDiff
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	handler := given_handlerForBindingWithModifier(t, subject.keybindingSpecs(), viewDetailName, 'r', gocui.ModAlt)
	actualErr = handler(gui, detailView)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 2 {
		t.Fatalf("expected one queued detail refresh and one queued diff refresh, actual %d", len(asyncRunner.runs))
	}
	then_statusLineContains(t, gui, string(loadingSpinnerFrames[0]))
	then_statusLineContains(t, gui, subject.selectedPullRequestDetailLoadingStatus())
	then_statusLineDoesNotContain(t, gui, pullRequestRefreshSuccessMessage)

	given_runQueuedAsync(t, asyncRunner, 0)

	then_statusLineContains(t, gui, string(loadingSpinnerFrames[0]))
	then_statusLineContains(t, gui, "Running `gh api repos/acme/widgets/pulls/42 -H 'Accept: application/vnd.github.v3.diff'`.")
	then_statusLineDoesNotContain(t, gui, pullRequestRefreshSuccessMessage)

	given_runQueuedAsync(t, asyncRunner, 1)

	then_statusLineContains(t, gui, pullRequestRefreshSuccessMessage)
}

func TestHelpPopup_GivenPullRequestsFocus_WhenTogglingHelp_ThenItShowsTheRefreshShortcut(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	then_helpEntryUsesKey(t, helpView.Buffer(), "Refresh PR list", "Alt+R")
}

func TestHelpPopup_GivenNotificationsFocus_WhenTogglingHelp_ThenItShowsTheRefreshShortcut(t *testing.T) {
	model := given_model()
	model.FocusNotificationsView()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	then_helpEntryUsesKey(t, helpView.Buffer(), "Refresh notifications", "Alt+R")
}

func TestHelpPopup_GivenPullRequestDetailFocus_WhenTogglingHelp_ThenItShowsTheRefreshShortcut(t *testing.T) {
	loader := &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Body 42", State: "OPEN"}}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	then_helpEntryUsesKey(t, helpView.Buffer(), "Refresh PR", "Alt+R")
}

func TestHelpPopup_GivenReviewDiffFocus_WhenTogglingHelp_ThenItShowsTheRefreshShortcut(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
		},
	}
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
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	then_helpEntryUsesKey(t, helpView.Buffer(), "Refresh PR", "Alt+R")
}
