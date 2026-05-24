package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const (
	reactionPickerTitle                    = "Add reaction"
	pullRequestReactionAddedSuccessMessage = "Reaction added"
	pullRequestReactionAlreadyAddedMessage = "Reaction already added"
)

func (program *Program) reactionPickerVisible() bool {
	return program.actionsPopupWidget.reactionPicker != nil
}

func (program *Program) currentReactionAction() (actionsPopupAction, bool) {
	target, ok := program.selectedPullRequestReactionActionTarget()
	if !ok {
		return actionsPopupAction{}, false
	}
	return program.addReactionAction().withGroup(target.popupGroup()), true
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
	program.actionsPopupWidget.reactionPicker = &reactionPickerState{target: target}
	program.actionsPopupWidget.searchEditor = nil
	program.actionsPopupWidget.errorMessage = ""
	program.model.OpenActionsPopup(len(program.currentActionsPopupActions()))
	return actionsPopupActionResult{}
}

func (program *Program) currentReactionPickerActions() []actionsPopupAction {
	if !program.reactionPickerVisible() {
		return nil
	}

	actions := make([]actionsPopupAction, 0, len(githubdomain.SupportedReactionContents))
	for _, content := range githubdomain.SupportedReactionContents {
		actions = append(actions, program.reactionPickerAction(content))
	}
	return actions
}

func (program *Program) reactionPickerAction(content githubdomain.ReactionContent) actionsPopupAction {
	title := reactionPickerActionMetadata(content)
	return actionsPopupAction{
		id:    "reaction-" + strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(string(content)), "+", "plus"), "-", "minus"),
		title: title,
		execute: func(_ *gocui.Gui) actionsPopupActionResult {
			return program.executeReactionPickerAction(content)
		},
	}
}

func reactionPickerActionMetadata(content githubdomain.ReactionContent) string {
	switch content {
	case githubdomain.ReactionContentThumbsUp:
		return "👍 Thumbs up (+1)"
	case githubdomain.ReactionContentThumbsDown:
		return "👎 Thumbs down (-1)"
	case githubdomain.ReactionContentLaugh:
		return "😄 Laugh"
	case githubdomain.ReactionContentHooray:
		return "🎉 Hooray"
	case githubdomain.ReactionContentConfused:
		return "😕 Confused"
	case githubdomain.ReactionContentHeart:
		return "❤️ Heart"
	case githubdomain.ReactionContentRocket:
		return "🚀 Rocket"
	case githubdomain.ReactionContentEyes:
		return "👀 Eyes"
	default:
		return strings.TrimSpace(string(content))
	}
}

func (program *Program) executeReactionPickerAction(content githubdomain.ReactionContent) actionsPopupActionResult {
	if !program.reactionPickerVisible() {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	target := program.actionsPopupWidget.reactionPicker.target
	if reactionGroupViewerHasReacted(target.reactionGroups, content) {
		program.setFeedback(program.model.Focus(), pullRequestReactionAlreadyAddedMessage)
		return actionsPopupActionResult{closePopup: true}
	}
	if !program.hasReactionMutations() {
		return actionsPopupActionResult{err: errors.New("github loader is unavailable")}
	}
	if err := program.reactionMutations.AddReaction(target.subjectID, content); err != nil {
		return actionsPopupActionResult{err: newTransientErrorPopupActionError(err)}
	}

	program.optimisticallyAddReaction(target, content)
	program.setFeedback(program.model.Focus(), pullRequestReactionAddedSuccessMessage)
	return actionsPopupActionResult{closePopup: true}
}

func reactionGroupViewerHasReacted(groups []githubdomain.ReactionGroup, content githubdomain.ReactionContent) bool {
	for _, group := range groups {
		if strings.TrimSpace(string(group.Content)) != strings.TrimSpace(string(content)) {
			continue
		}
		return group.ViewerHasReacted
	}
	return false
}
