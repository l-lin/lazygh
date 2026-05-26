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
	return program.dispatch(gui, MsgOpenLinkUnderCursorRequested{View: view})
}

func (program *Program) openCurrentLink(_ *gocui.View) error {
	if program.linkOpener == nil {
		return ErrLinkOpenerUnavailable
	}

	url, ok := program.currentDetailCursorLink(nil)
	if !ok {
		return ErrNoLinkUnderCursor
	}

	return program.linkOpener.Open(url)
}

func (program *Program) currentDetailCursorLink(_ *gocui.View) (string, bool) {
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
	_, ok := program.currentDetailCursorLink(nil)
	return ok
}

func (program *Program) openLinkUnderCursorActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "open-link-under-cursor",
		title:   "Open link under cursor",
		icon:    actionsPopupOpenLinkIcon,
		execute: actionsPopupExecuteErr(program.executeOpenLinkUnderCursorAction),
	}
}

func (program *Program) executeOpenLinkUnderCursorAction(gui *gocui.Gui) error {
	if program.linkOpener == nil {
		return errors.New(openLinkOpenerUnavailableMessage)
	}
	url, ok := program.currentDetailCursorLink(nil)
	if !ok {
		return errors.New(openLinkUnavailableMessage)
	}
	if err := program.dispatch(gui, MsgOpenBrowserURLRequested{URL: url, SuccessMessage: openLinkSuccessMessage, FailureMessage: openLinkFailureMessage, Target: program.model.Focus()}); err != nil {
		return err
	}
	return program.closeActionsPopupIfVisible(gui)
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
