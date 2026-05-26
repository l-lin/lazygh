package tui

import (
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
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	if target, ok := program.selectedPullRequestReactionActionTarget(); ok {
		requested = MsgOpenReactionPickerRequested{Target: target}
	}
	return actionsPopupAction{
		id:        "add-reaction",
		title:     reactionPickerTitle,
		icon:      actionsPopupAddReactionIcon,
		requested: requested,
	}
}

func (program *Program) executeOpenReactionPickerAction(gui *gocui.Gui) error {
	target, ok := program.selectedPullRequestReactionActionTarget()
	if !ok {
		return errActionsPopupActionUnavailable
	}
	return program.dispatch(gui, MsgOpenReactionPickerRequested{Target: target})
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
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	if program.reactionPickerVisible() {
		requested = MsgAddReactionRequested{Target: program.actionsPopupWidget.reactionPicker.target, Content: content}
	}
	return actionsPopupAction{
		id:        "reaction-" + strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(string(content)), "+", "plus"), "-", "minus"),
		title:     title,
		requested: requested,
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

func (program *Program) executeReactionPickerAction(gui *gocui.Gui, content githubdomain.ReactionContent) error {
	if !program.reactionPickerVisible() {
		return errActionsPopupActionUnavailable
	}
	return program.dispatch(gui, MsgAddReactionRequested{Target: program.actionsPopupWidget.reactionPicker.target, Content: content})
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
