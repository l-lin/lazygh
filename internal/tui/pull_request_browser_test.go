package tui

import (
	"reflect"
	"testing"
)

func TestActionsPopup_GivenOpenBrowserActionSelected_WhenExecuting_ThenItUsesTheBrowserHandlerAndClosesThePopup(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("browser", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "browser"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.openBrowserCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected open browser calls %v, actual %v", []string{"acme/widgets#42"}, loader.openBrowserCalls)
	}
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
	then_statusLineContains(t, gui, pullRequestBrowserOpenSuccessMessage)
	then_viewDoesNotExist(t, gui, viewPullRequestsFooterName)
}
