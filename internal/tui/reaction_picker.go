package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	actionsPopupAddReactionIcon            = "󰞅"
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
		id:       "add-reaction",
		title:    reactionPickerTitle,
		icon:     actionsPopupAddReactionIcon,
		keywords: []string{"reaction", "react", "emoji", "+1", "thumbs up", "thumbs down", "laugh", "hooray", "confused", "heart", "rocket", "eyes"},
		execute:  program.executeOpenReactionPickerAction,
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
	title, keywords := reactionPickerActionMetadata(content)
	return actionsPopupAction{
		id:       "reaction-" + strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(string(content)), "+", "plus"), "-", "minus"),
		title:    title,
		keywords: keywords,
		execute: func(_ *gocui.Gui) actionsPopupActionResult {
			return program.executeReactionPickerAction(content)
		},
	}
}

func reactionPickerActionMetadata(content githubcli.ReactionContent) (string, []string) {
	switch content {
	case githubcli.ReactionContentThumbsUp:
		return "👍 Thumbs up (+1)", []string{"thumbs up", "+1", "approve", "like"}
	case githubcli.ReactionContentThumbsDown:
		return "👎 Thumbs down (-1)", []string{"thumbs down", "-1", "dislike"}
	case githubcli.ReactionContentLaugh:
		return "😄 Laugh", []string{"laugh", "smile", "funny"}
	case githubcli.ReactionContentHooray:
		return "🎉 Hooray", []string{"hooray", "celebrate", "party"}
	case githubcli.ReactionContentConfused:
		return "😕 Confused", []string{"confused", "question", "unsure"}
	case githubcli.ReactionContentHeart:
		return "❤️ Heart", []string{"heart", "love"}
	case githubcli.ReactionContentRocket:
		return "🚀 Rocket", []string{"rocket", "ship", "launch"}
	case githubcli.ReactionContentEyes:
		return "👀 Eyes", []string{"eyes", "watching", "look"}
	default:
		trimmedContent := strings.TrimSpace(string(content))
		return trimmedContent, []string{trimmedContent}
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
