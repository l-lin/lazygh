package tui

import (
	"reflect"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenOpenBrowserActionSelected_WhenExecuting_ThenItKeepsThePopupOpenShowsTheStatusLineSpinnerAndDelaysTheGitHubCall(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Original body", State: "OPEN"})}
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("browser", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "browser"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(loader.openBrowserCalls) != 0 {
		t.Fatalf("expected the open-browser call to wait for the queued run, actual %v", loader.openBrowserCalls)
	}
	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
	_, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_statusLineContains(t, gui, string(loadingSpinnerFrames[0]))
	then_statusLineContains(t, gui, formatRunningCommandStatus(openPullRequestInBrowserCommand("acme/widgets", 42)))
}

func TestActionsPopup_GivenOpenBrowserActionSelected_WhenExecuting_ThenItUsesTheBrowserHandlerAndClosesThePopup(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("browser", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "browser"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.openBrowserCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected open browser calls %v, actual %v", []string{"acme/widgets#42"}, loader.openBrowserCalls)
	}
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
	then_statusLineContains(t, gui, pullRequestBrowserOpenSuccessMessage)
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, a: action")
	then_viewDoesNotExist(t, gui, viewPullRequestsFooterName)
}
