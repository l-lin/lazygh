package tui

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	pullRequestReactionRemovedSuccessMessage = "Reaction removed"
	pullRequestReactionAlreadyRemovedMessage = "Reaction already removed"
)

type pullRequestReactionRemovalTarget struct {
	pullRequestReactionActionTarget
	content githubcli.ReactionContent
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
		execute: func(_ *gocui.Gui) actionsPopupActionResult {
			return program.executeRemoveReactionAction(target)
		},
	}, true
}

func (program *Program) executeRemoveReactionAction(target pullRequestReactionRemovalTarget) actionsPopupActionResult {
	if strings.TrimSpace(target.subjectID) == "" {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if !reactionGroupViewerHasReacted(target.reactionGroups, target.content) {
		program.setFeedback(program.model.Focus(), pullRequestReactionAlreadyRemovedMessage)
		return actionsPopupActionResult{closePopup: true}
	}
	if !program.hasReactionMutations() {
		return actionsPopupActionResult{err: errors.New("github loader is unavailable")}
	}
	if err := program.reactionMutations.RemoveReaction(target.subjectID, githubcli.ToDomainReactionContent(target.content)); err != nil {
		return actionsPopupActionResult{err: err}
	}

	program.optimisticallyRemoveReaction(target)
	program.setFeedback(program.model.Focus(), pullRequestReactionRemovedSuccessMessage)
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) selectedPullRequestReactionRemovalTarget() (pullRequestReactionRemovalTarget, bool) {
	if program.model.Focus() != FocusDetailView {
		return pullRequestReactionRemovalTarget{}, false
	}

	target, ok := program.selectedPullRequestReactionActionTarget()
	if !ok {
		return pullRequestReactionRemovalTarget{}, false
	}

	document := program.currentDetailDocument(nil)
	content, ok := viewerReactionContentAtCursor(document, program.detailViewState.cursor, target.reactionGroups)
	if !ok {
		return pullRequestReactionRemovalTarget{}, false
	}

	return pullRequestReactionRemovalTarget{pullRequestReactionActionTarget: target, content: content}, true
}

func viewerReactionContentAtCursor(document detailDocument, position detailPosition, groups []githubcli.ReactionGroup) (githubcli.ReactionContent, bool) {
	if len(document.lines) == 0 || len(groups) == 0 {
		return "", false
	}

	position = document.clampPosition(position)
	if position.line < 0 || position.line >= len(document.lines) {
		return "", false
	}

	return viewerReactionContentAtLineColumn(string(document.lines[position.line]), position.column, groups)
}

func viewerReactionContentAtLineColumn(line string, column int, groups []githubcli.ReactionGroup) (githubcli.ReactionContent, bool) {
	searchStartByte := 0
	for _, content := range githubcli.SupportedReactionContents {
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
