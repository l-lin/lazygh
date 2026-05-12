package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
)

type panelViewNumber int

const (
	mainPanelViewZero panelViewNumber = iota
	sidePanelViewOne
	sidePanelViewTwo
	sidePanelViewThree
)

func given_panelViewContractBrowserModel() *Model {
	return given_pullRequestCommentModel()
}

func panelViewName(number panelViewNumber) string {
	switch number {
	case mainPanelViewZero:
		return viewDetailName
	case sidePanelViewTwo:
		return viewPullRequestsName
	case sidePanelViewThree:
		return viewNotificationsName
	default:
		return viewUserName
	}
}

func focusForPanelViewNumber(number panelViewNumber) Focus {
	switch number {
	case mainPanelViewZero:
		return FocusDetailView
	case sidePanelViewTwo:
		return FocusPullRequestsView
	case sidePanelViewThree:
		return FocusNotificationsView
	default:
		return FocusUserView
	}
}

func when_focusingPanelViewNumber(subject *Model, target panelViewNumber) {
	switch target {
	case mainPanelViewZero:
		subject.FocusDetailView()
	case sidePanelViewTwo:
		subject.FocusPullRequestsView()
	case sidePanelViewThree:
		subject.FocusNotificationsView()
	default:
		subject.FocusUserView()
	}
}

func then_modelFocusesPanelViewNumber(t *testing.T, subject *Model, expected panelViewNumber) {
	t.Helper()

	if actual := subject.Focus(); actual != focusForPanelViewNumber(expected) {
		t.Fatalf("expected focus %v for panel view %d, actual %v", focusForPanelViewNumber(expected), expected, actual)
	}
}

func then_currentPanelViewIs(t *testing.T, gui *gocui.Gui, expected panelViewNumber) {
	t.Helper()

	then_currentViewNameIs(t, gui, panelViewName(expected))
}

func then_panelViewShowsVisibleNumber(t *testing.T, gui *gocui.Gui, expected panelViewNumber) {
	t.Helper()

	view, actualErr := gui.View(panelViewName(expected))
	then_noError(t, actualErr)

	visibleTitle := strings.TrimSpace(view.TitlePrefix + view.Title)
	expectedLabel := fmt.Sprintf("[%d]", expected)
	if !strings.Contains(visibleTitle, expectedLabel) {
		t.Fatalf("expected %q to show %q, actual title prefix=%q title=%q", panelViewName(expected), expectedLabel, view.TitlePrefix, view.Title)
	}
}
