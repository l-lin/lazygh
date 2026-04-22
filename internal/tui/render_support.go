package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) detailViewContent() string {
	if program.reviewSession.active {
		return program.reviewSessionDetailContent()
	}
	if program.model.currentSideFocus() == FocusPullRequestsView {
		row, ok := program.model.SelectedPullRequestRow()
		if ok && row.Summary != nil && pullRequestDetailKey(row.Summary.Repository, row.Summary.Number) != "" {
			if result, ok := program.pullRequestDetailForSummary(*row.Summary); ok {
				if result.err != nil {
					return renderPullRequestDetailError(*row.Summary, result.err)
				}

				header := renderPullRequestDetailHeader(*row.Summary, result.detail)
				content := renderPullRequestDescription(*row.Summary, result.detail, program.markdownRenderer, program.detailWrapWidth)
				if program.activeDetailTab == CommentsDetailTab {
					content = renderPullRequestCommentsTab(result.detail.Comments, result.detail.InlineComments, program.markdownRenderer, program.detailWrapWidth)
				}
				return renderPullRequestDetailContent(header, content)
			}
			return renderPullRequestDetailLoading(*row.Summary)
		}
	}

	item, ok := program.model.detailItem()
	if !ok {
		return "No detail available."
	}

	return program.fallbackDetailViewContent(item)
}

func (program *Program) fallbackDetailViewContent(item Item) string {
	header := program.detailHeader(item)
	body := strings.TrimSpace(item.Detail)
	if body == "" {
		body = "No description available. Even the dummy data is disappointed."
	}

	return renderPullRequestDetailContent(header, body)
}

func (program *Program) detailHeader(item Item) string {
	source := "Connected user"
	if program.model.currentSideFocus() == FocusPullRequestsView {
		source = fmt.Sprintf("%s tab", program.model.ActivePullRequestTab().Label())
	}

	return fmt.Sprintf("%s\n%s", source, item.Title)
}

func (program *Program) currentDetailDocument(view *gocui.View) detailDocument {
	width := program.detailWrapWidth
	if view != nil && view.InnerWidth() > 0 {
		width = view.InnerWidth()
	}
	if width < 1 {
		width = 1
	}

	return newDetailDocument(program.detailViewContent(), width)
}

func (program *Program) syncDetailViewState(detailDocument detailDocument, viewportHeight int) {
	identity := program.currentDetailIdentity()
	if identity != program.lastDetailIdentity {
		program.lastDetailIdentity = identity
		program.detailViewState.reset()
	}

	program.detailViewState.sync(detailDocument, viewportHeight)
	program.detailViewState.syncSearch(detailDocument, program.model.DetailSearchQuery())
}

func (program *Program) pullRequestsTabLabels() []string {
	return []string{
		program.pullRequestsTabLabel(MyPullRequestsTab),
		program.pullRequestsTabLabel(RequestedPullRequestsTab),
	}
}

func (program *Program) pullRequestsTabLabel(tab PullRequestTab) string {
	label := tab.Label()
	count, ok := program.pullRequestsCount(tab)
	if !ok {
		return label
	}

	return fmt.Sprintf("%s (%d)", label, count)
}

func (program *Program) pullRequestsCount(tab PullRequestTab) (int, bool) {
	switch tab {
	case RequestedPullRequestsTab:
		return program.requestedPullRequestsCount, program.requestedPullRequestsCountKnown
	default:
		return program.myPullRequestsCount, program.myPullRequestsCountKnown
	}
}

func (program *Program) shouldHighlightSelection(focus Focus, selectable bool) bool {
	if !selectable {
		return false
	}

	if program.model.Focus() == focus {
		return true
	}

	return program.model.Focus() == FocusDetailView && program.model.currentSideFocus() == focus
}

func searchNoMatchesMessage(query string) string {
	return fmt.Sprintf("No matches for %q.", strings.TrimSpace(query))
}

func (program *Program) layoutContentHeight(maxY int) int {
	if maxY < 1 {
		return 1
	}
	if program.bottomPromptVisible() && maxY > 1 {
		return maxY - 1
	}
	return maxY
}

func (program *Program) bottomPromptVisible() bool {
	return program.model.ActionsPopupSearchActive()
}
