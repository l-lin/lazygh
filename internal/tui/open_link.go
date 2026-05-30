package tui

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"
)

var (
	ErrNoLinkUnderCursor = errors.New("no link under cursor")
	visibleURLPattern    = regexp.MustCompile(`https?://[^\s<>"']+`)
)

const (
	openLinkSuccessMessage           = iconLink + " Link opened"
	openLinkFailureMessage           = iconWarning + " Open failed"
	openLinkUnavailableMessage       = iconUnavailable + " No link under cursor"
	openLinkOpenerUnavailableMessage = iconUnavailable + " Link opener unavailable"
)

func (program *Program) openLinkUnderCursor(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgOpenLinkUnderCursorRequested{})
}

func (program *Program) currentDetailCursorLink() (string, bool) {
	selection := program.currentDetailCursorSelection()
	if actual, ok := selection.document.linkAt(selection.state.cursor); ok {
		return actual, true
	}
	return program.buildLinkUnderCursor(selection.document)
}

func (program *Program) buildLinkUnderCursor(document detailDocument) (string, bool) {
	entry, ok := program.browserOverviewBuildEntryAtDetailCursorDocument(document)
	if !ok {
		return "", false
	}

	actual := strings.TrimSpace(entry.Link)
	if actual == "" {
		return "", false
	}
	return actual, true
}

func pullRequestOverviewEntryAtBodyLine(section browserDetailSection, bodyLine int) (pullRequestOverviewEntry, bool) {
	contentLineIndex := bodyLine - 1
	if contentLineIndex < 0 {
		return pullRequestOverviewEntry{}, false
	}

	for _, entry := range section.overviewEntries {
		if strings.TrimSpace(entry.Label) != "" {
			if contentLineIndex == 0 {
				return entry, true
			}
			contentLineIndex--
		}
		if strings.TrimSpace(entry.Detail) != "" {
			if contentLineIndex == 0 {
				return pullRequestOverviewEntry{}, false
			}
			contentLineIndex--
		}
	}
	return pullRequestOverviewEntry{}, false
}

func (program *Program) detailCursorHasLink() bool {
	_, ok := program.currentDetailCursorLink()
	return ok
}

func (program *Program) openLinkUnderCursorActionsPopupAction() actionsPopupAction {
	capabilities := program.currentInteractionCapabilitySnapshot()
	var requested Msg = MsgOpenLinkUnderCursorRequested{}
	if !capabilities.linkOpenerAvailable {
		requested = actionsPopupErrorRequested(errors.New(openLinkOpenerUnavailableMessage))
	}
	return actionsPopupAction{
		id:        "open-link-under-cursor",
		title:     "Open link under cursor",
		icon:      actionsPopupOpenLinkIcon,
		requested: requested,
	}
}

func (document detailDocument) linkAt(position detailPosition) (string, bool) {
	position = document.clampPosition(position)
	if target := strings.TrimSpace(document.hyperlinkTargetAt(position)); target != "" {
		return target, true
	}

	return document.visibleURLAt(position)
}

func (document detailDocument) hyperlinkTargetAt(position detailPosition) string {
	if position.line < 0 || position.line >= len(document.lineHyperlinkTargets) {
		return ""
	}
	if position.column < 0 || position.column >= len(document.lineHyperlinkTargets[position.line]) {
		return ""
	}

	return document.lineHyperlinkTargets[position.line][position.column]
}

func (document detailDocument) visibleURLAt(position detailPosition) (string, bool) {
	if position.line < 0 || position.line >= len(document.lines) {
		return "", false
	}

	line := string(document.lines[position.line])
	for _, byteRange := range visibleURLPattern.FindAllStringIndex(line, -1) {
		candidate := trimVisibleURL(line[byteRange[0]:byteRange[1]])
		if candidate == "" {
			continue
		}

		startColumn := utf8.RuneCountInString(line[:byteRange[0]])
		endColumn := startColumn + utf8.RuneCountInString(candidate)
		if position.column < startColumn || position.column >= endColumn {
			continue
		}

		return candidate, true
	}

	return "", false
}

func trimVisibleURL(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimRight(trimmed, ".,:;!?)]}\"'")
	return trimmed
}
