package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	viewUserFooterName         = "user-footer"
	viewPullRequestsFooterName = "pull-requests-footer"
	viewDetailFooterName       = "detail-footer"

	pullRequestDetailLoadingTitle = "Loading pull request detail..."
)

type paneFooterState struct {
	text string
}

func (state paneFooterState) Visible() bool {
	return strings.TrimSpace(state.text) != ""
}

func (program *Program) layoutPaneFooterViews(gui *gocui.Gui) error {
	for _, focus := range []Focus{FocusUserView, FocusPullRequestsView, FocusDetailView} {
		if err := program.layoutPaneFooterView(gui, focus); err != nil {
			return err
		}
	}

	return nil
}

func (program *Program) layoutPaneFooterView(gui *gocui.Gui, focus Focus) error {
	viewName := paneFooterViewName(focus)
	if !program.model.PaneVisible(focus) {
		return deleteViewIfPresent(gui, viewName)
	}

	state := program.paneFooterStateFor(focus)
	if !state.Visible() {
		return deleteViewIfPresent(gui, viewName)
	}

	view, err := program.layoutPaneBottomOverlayView(gui, viewName, paneViewName(focus))
	if err != nil {
		return err
	}

	program.configurePaneFooterView(view)
	program.renderPaneFooterView(view, state.text)
	_, err = gui.SetViewOnTop(viewName)
	if isUnknownViewError(err) {
		return nil
	}

	return err
}

func (program *Program) paneFooterStateFor(focus Focus) paneFooterState {
	if program.model.SearchActive() && program.model.SearchTarget() == focus {
		return paneFooterState{}
	}

	if message := strings.TrimSpace(program.feedbackMessageFor(focus)); message != "" {
		return paneFooterState{text: message}
	}

	if message := strings.TrimSpace(program.loadingFooterText(focus)); message != "" {
		return paneFooterState{text: message}
	}

	if message := strings.TrimSpace(program.appliedSearchFooterText(focus)); message != "" {
		return paneFooterState{text: message}
	}

	return paneFooterState{}
}

func (program *Program) appliedSearchFooterText(focus Focus) string {
	if program.reviewSession.active {
		if focus != FocusDetailView {
			return ""
		}
		query := program.model.appliedSearchQuery(FocusDetailView, MyPullRequestsTab)
		return searchSummaryText(query, countSearchMatches(program.detailViewContent(), query))
	}

	switch focus {
	case FocusPullRequestsView:
		query := program.model.appliedSearchQuery(FocusPullRequestsView, program.model.ActivePullRequestTab())
		return searchSummaryText(query, len(program.model.VisiblePullRequests()))
	case FocusDetailView:
		query := program.model.appliedSearchQuery(FocusDetailView, MyPullRequestsTab)
		return searchSummaryText(query, countSearchMatches(program.detailViewContent(), query))
	default:
		query := program.model.appliedSearchQuery(FocusUserView, MyPullRequestsTab)
		return searchSummaryText(query, len(program.model.VisibleUsers()))
	}
}

func searchSummaryText(query string, count int) string {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return ""
	}

	return fmt.Sprintf("/%s (%d %s)", trimmedQuery, count, pluralize(count, "match", "matches"))
}

func (program *Program) loadingFooterText(focus Focus) string {
	if program.reviewSession.active {
		if focus == FocusDetailView && program.selectedPullRequestDetailLoading() {
			return pullRequestDetailLoadingTitle
		}
		return ""
	}

	switch focus {
	case FocusPullRequestsView:
		return program.activePullRequestsLoadingText()
	case FocusDetailView:
		if program.selectedPullRequestDetailLoading() {
			return pullRequestDetailLoadingTitle
		}
	}

	return ""
}

func (program *Program) activePullRequestsLoadingText() string {
	switch program.model.ActivePullRequestTab() {
	case RequestedPullRequestsTab:
		if program.requestedPullRequestsLoading {
			return requestedPullRequestsLoadingTitle
		}
	default:
		if program.myPullRequestsLoading {
			return myPullRequestsLoadingTitle
		}
	}

	return ""
}

func (program *Program) selectedPullRequestDetailLoading() bool {
	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return false
	}

	return program.pullRequestDetailLoadInFlight[pullRequestDetailKey(summary.Repository, summary.Number)]
}

func (program *Program) configurePaneFooterView(view *gocui.View) {
	program.configureBottomPromptView(view, nil, false)
	view.Editable = false
	view.Editor = nil
}

func (program *Program) renderPaneFooterView(view *gocui.View, text string) {
	if view == nil {
		return
	}

	view.Clear()
	view.SetOrigin(0, 0)
	view.SetCursor(0, 0)
	fmt.Fprint(view, strings.TrimSpace(text))
}

func paneFooterViewName(focus Focus) string {
	switch focus {
	case FocusPullRequestsView:
		return viewPullRequestsFooterName
	case FocusDetailView:
		return viewDetailFooterName
	default:
		return viewUserFooterName
	}
}

func paneViewName(focus Focus) string {
	switch focus {
	case FocusPullRequestsView:
		return viewPullRequestsName
	case FocusDetailView:
		return viewDetailName
	default:
		return viewUserName
	}
}
