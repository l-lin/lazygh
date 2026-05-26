package tui

import (
	"strings"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const (
	pullRequestReactionRemovedSuccessMessage = "Reaction removed"
	pullRequestReactionAlreadyRemovedMessage = "Reaction already removed"
)

type pullRequestReactionRemovalTarget struct {
	pullRequestReactionActionTarget
	content githubdomain.ReactionContent
}

func (program *Program) currentReactionRemovalAction() (actionsPopupAction, bool) {
	target, ok := program.selectedPullRequestReactionRemovalTarget()
	if !ok {
		return actionsPopupAction{}, false
	}

	reactionTitle := reactionPickerActionMetadata(target.content)
	reactionID := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(string(target.content)), "+", "plus"), "-", "minus")
	return actionsPopupAction{
		id:    "remove-reaction-" + reactionID,
		title: "Remove reaction " + reactionTitle,
		icon:  actionsPopupRemoveReactionIcon,
		execute: actionsPopupExecuteErr(func(gui *gocui.Gui) error {
			return program.executeRemoveReactionAction(gui, target)
		}),
	}.withGroup(target.popupGroup()), true
}

func (program *Program) executeRemoveReactionAction(gui *gocui.Gui, target pullRequestReactionRemovalTarget) error {
	if strings.TrimSpace(target.subjectID) == "" {
		return errActionsPopupActionUnavailable
	}
	return program.dispatch(gui, MsgReactionRemovalRequested{Target: target})
}

func (program *Program) selectedPullRequestReactionRemovalTarget() (pullRequestReactionRemovalTarget, bool) {
	if program.model.Focus() != FocusDetailView {
		return pullRequestReactionRemovalTarget{}, false
	}

	target, ok := program.selectedPullRequestReactionActionTarget()
	if !ok {
		return pullRequestReactionRemovalTarget{}, false
	}

	selection := program.currentDetailCursorSelection()
	content, ok := viewerReactionContentAtCursor(selection.document, selection.state.cursor, target.reactionGroups)
	if !ok {
		return pullRequestReactionRemovalTarget{}, false
	}

	return pullRequestReactionRemovalTarget{pullRequestReactionActionTarget: target, content: content}, true
}

func viewerReactionContentAtCursor(document detailDocument, position detailPosition, groups []githubdomain.ReactionGroup) (githubdomain.ReactionContent, bool) {
	if len(document.lines) == 0 || len(groups) == 0 {
		return "", false
	}

	position = document.clampPosition(position)
	if position.line < 0 || position.line >= len(document.lines) {
		return "", false
	}

	return viewerReactionContentAtLineColumn(string(document.lines[position.line]), position.column, groups)
}

func viewerReactionContentAtLineColumn(line string, column int, groups []githubdomain.ReactionGroup) (githubdomain.ReactionContent, bool) {
	searchStartByte := 0
	for _, content := range githubdomain.SupportedReactionContents {
		group, ok := reactionGroupForContent(groups, content)
		if !ok || !group.ViewerHasReacted || group.TotalCount <= 0 {
			continue
		}

		visiblePill := styledTextVisibleString(renderReactionGroup(group))
		if visiblePill == "" {
			continue
		}

		matchByte := strings.Index(line[searchStartByte:], visiblePill)
		if matchByte >= 0 {
			matchByte += searchStartByte
		} else {
			matchByte = strings.Index(line, visiblePill)
		}
		if matchByte < 0 {
			continue
		}

		startColumn := utf8.RuneCountInString(line[:matchByte])
		endColumn := startColumn + utf8.RuneCountInString(visiblePill)
		if column >= startColumn && column < endColumn {
			return group.Content, true
		}

		searchStartByte = matchByte + len(visiblePill)
	}

	return "", false
}

func styledTextVisibleString(text string) string {
	styledLines := splitStyledTextLines(text)
	visibleLines := make([]string, 0, len(styledLines))
	for _, styledLine := range styledLines {
		visibleLines = append(visibleLines, string(styledLine.runes))
	}
	return strings.Join(visibleLines, "\n")
}
