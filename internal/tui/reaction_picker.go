package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	reactionPickerTitle                    = "Add reaction"
	pullRequestReactionAddedSuccessMessage = "Reaction added"
	pullRequestReactionAlreadyAddedMessage = "Reaction already added"
)

func (program *Program) reactionPickerVisible() bool {
	return program.reactionPicker != nil
}

func (program *Program) currentReactionAction() (actionsPopupAction, bool) {
	if _, ok := program.selectedPullRequestReactionActionTarget(); !ok {
		return actionsPopupAction{}, false
	}
	return program.addReactionAction(), true
}

func (program *Program) addReactionAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "add-reaction",
		title:   reactionPickerTitle,
		icon:    actionsPopupAddReactionIcon,
		execute: program.executeOpenReactionPickerAction,
	}
}

func (program *Program) executeOpenReactionPickerAction(_ *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestReactionActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	program.reactionPicker = &reactionPickerState{target: target}
	program.actionsPopupSearchEditor = nil
	program.actionsPopupErrorMessage = ""
	program.model.OpenActionsPopup(len(program.currentActionsPopupActions()))
	return actionsPopupActionResult{}
}

func (program *Program) currentReactionPickerActions() []actionsPopupAction {
	if !program.reactionPickerVisible() {
		return nil
	}

	actions := make([]actionsPopupAction, 0, len(githubcli.SupportedReactionContents))
	for _, content := range githubcli.SupportedReactionContents {
		actions = append(actions, program.reactionPickerAction(content))
	}
	return actions
}

func (program *Program) reactionPickerAction(content githubcli.ReactionContent) actionsPopupAction {
	title := reactionPickerActionMetadata(content)
	return actionsPopupAction{
		id:    "reaction-" + strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(string(content)), "+", "plus"), "-", "minus"),
		title: title,
		execute: func(_ *gocui.Gui) actionsPopupActionResult {
			return program.executeReactionPickerAction(content)
		},
	}
}

func reactionPickerActionMetadata(content githubcli.ReactionContent) string {
	switch content {
	case githubcli.ReactionContentThumbsUp:
		return "👍 Thumbs up (+1)"
	case githubcli.ReactionContentThumbsDown:
		return "👎 Thumbs down (-1)"
	case githubcli.ReactionContentLaugh:
		return "😄 Laugh"
	case githubcli.ReactionContentHooray:
		return "🎉 Hooray"
	case githubcli.ReactionContentConfused:
		return "😕 Confused"
	case githubcli.ReactionContentHeart:
		return "❤️ Heart"
	case githubcli.ReactionContentRocket:
		return "🚀 Rocket"
	case githubcli.ReactionContentEyes:
		return "👀 Eyes"
	default:
		return strings.TrimSpace(string(content))
	}
}

func (program *Program) executeReactionPickerAction(content githubcli.ReactionContent) actionsPopupActionResult {
	if !program.reactionPickerVisible() {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	target := program.reactionPicker.target
	if reactionGroupViewerHasReacted(target.reactionGroups, content) {
		program.setFeedback(program.model.Focus(), pullRequestReactionAlreadyAddedMessage)
		return actionsPopupActionResult{closePopup: true}
	}
	if program.githubLoader == nil {
		return actionsPopupActionResult{err: errors.New("github loader is unavailable")}
	}
	if err := program.githubLoader.AddReaction(target.subjectID, content); err != nil {
		return actionsPopupActionResult{err: err}
	}

	program.invalidatePullRequestDetail(target.repository, target.number)
	if target.invalidateDiff {
		program.invalidatePullRequestDiff(target.repository, target.number)
	}
	program.setFeedback(program.model.Focus(), pullRequestReactionAddedSuccessMessage)
	return actionsPopupActionResult{closePopup: true}
}

func reactionGroupViewerHasReacted(groups []githubcli.ReactionGroup, content githubcli.ReactionContent) bool {
	for _, group := range groups {
		if strings.TrimSpace(string(group.Content)) != strings.TrimSpace(string(content)) {
			continue
		}
		return group.ViewerHasReacted
	}
	return false
}
