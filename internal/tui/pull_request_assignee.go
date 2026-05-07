package tui

import (
	"errors"
	"sort"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	assignPullRequestActionTitle              = "Assign PR"
	assigneePickerTitle                       = "Select assignees"
	pullRequestAssigneesUpdatedSuccessMessage = "PR assignees updated"
)

var errNoAssignableUsers = errors.New("no assignable users available")

type pullRequestAssigneePickerTarget struct {
	repository string
	number     int
	assignees  []githubcli.PullRequestAuthor
}

type assigneePickerState struct {
	target                 pullRequestAssigneePickerTarget
	candidates             []githubcli.PullRequestAuthor
	selectedLogins         map[string]bool
	originalSelectedLogins map[string]bool
}

func (program *Program) assigneePickerVisible() bool {
	return program.assigneePicker != nil
}

func (program *Program) currentAssignPullRequestAction() (actionsPopupAction, bool) {
	if _, ok := program.selectedPullRequestAssigneePickerTarget(); !ok {
		return actionsPopupAction{}, false
	}
	return program.assignPullRequestAction(), true
}

func (program *Program) assignPullRequestAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "assign-pull-request",
		title:    assignPullRequestActionTitle,
		icon:     actionsPopupEditPullRequestIcon,
		keywords: []string{"assign", "assignee", "owner", "reviewer", "people"},
		execute:  program.executeOpenAssigneePickerAction,
	}
}

func (program *Program) executeOpenAssigneePickerAction(_ *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestAssigneePickerTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if program.githubLoader == nil {
		return actionsPopupActionResult{err: errors.New("github loader is unavailable")}
	}

	candidates, err := program.githubLoader.ListAssignableUsers(target.repository)
	if err != nil {
		return actionsPopupActionResult{err: err}
	}

	picker, err := newAssigneePickerState(target, candidates)
	if err != nil {
		return actionsPopupActionResult{err: err}
	}

	program.assigneePicker = picker
	program.actionsPopupSearchEditor = nil
	program.actionsPopupErrorMessage = ""
	program.model.OpenActionsPopup(len(program.currentActionsPopupActions()))
	return actionsPopupActionResult{}
}

func newAssigneePickerState(target pullRequestAssigneePickerTarget, candidates []githubcli.PullRequestAuthor) (*assigneePickerState, error) {
	mergedCandidates := mergedAssigneePickerCandidates(target.assignees, candidates)
	if len(mergedCandidates) == 0 {
		return nil, errNoAssignableUsers
	}

	selectedLogins := map[string]bool{}
	for _, assignee := range target.assignees {
		trimmedLogin := strings.TrimSpace(assignee.Login)
		if trimmedLogin == "" {
			continue
		}
		selectedLogins[trimmedLogin] = true
	}

	originalSelectedLogins := map[string]bool{}
	for login, selected := range selectedLogins {
		originalSelectedLogins[login] = selected
	}

	return &assigneePickerState{
		target:                 target,
		candidates:             mergedCandidates,
		selectedLogins:         selectedLogins,
		originalSelectedLogins: originalSelectedLogins,
	}, nil
}

func mergedAssigneePickerCandidates(currentAssignees []githubcli.PullRequestAuthor, assignableUsers []githubcli.PullRequestAuthor) []githubcli.PullRequestAuthor {
	mergedCandidates := make([]githubcli.PullRequestAuthor, 0, len(currentAssignees)+len(assignableUsers))
	seenLogins := map[string]bool{}
	for _, candidate := range append(append([]githubcli.PullRequestAuthor(nil), currentAssignees...), assignableUsers...) {
		normalizedCandidate := normalizedAssigneePickerCandidate(candidate)
		if normalizedCandidate.Login == "" || seenLogins[normalizedCandidate.Login] {
			continue
		}
		seenLogins[normalizedCandidate.Login] = true
		mergedCandidates = append(mergedCandidates, normalizedCandidate)
	}
	sort.SliceStable(mergedCandidates, func(i int, j int) bool {
		return mergedCandidates[i].Login < mergedCandidates[j].Login
	})
	return mergedCandidates
}

func (program *Program) selectedPullRequestAssigneePickerTarget() (pullRequestAssigneePickerTarget, bool) {
	if !program.pullRequestAssigneeStateVisible() {
		return pullRequestAssigneePickerTarget{}, false
	}

	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return pullRequestAssigneePickerTarget{}, false
	}
	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return pullRequestAssigneePickerTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return pullRequestAssigneePickerTarget{}, false
	}

	return pullRequestAssigneePickerTarget{
		repository: repository,
		number:     summary.Number,
		assignees:  append([]githubcli.PullRequestAuthor(nil), result.detail.Assignees...),
	}, true
}

func (program *Program) pullRequestAssigneeStateVisible() bool {
	if !program.isPullRequestContext() {
		return false
	}
	if program.reviewSession.active {
		return program.model.Focus() == FocusDetailView && program.reviewSessionShowsDescription()
	}
	if program.model.Focus() != FocusPullRequestsView && program.model.Focus() != FocusDetailView {
		return false
	}
	return program.shouldShowPullRequestDetailTabs() && program.activeDetailTab == DescriptionDetailTab
}

func (program *Program) currentAssigneePickerActions() []actionsPopupAction {
	if !program.assigneePickerVisible() {
		return nil
	}

	actions := make([]actionsPopupAction, 0, len(program.assigneePicker.candidates))
	for _, candidate := range program.assigneePicker.candidates {
		candidate := candidate
		actions = append(actions, actionsPopupAction{
			id:       "assignee-" + strings.ToLower(strings.TrimSpace(candidate.Login)),
			title:    program.assigneePickerLabel(candidate),
			keywords: assigneePickerKeywords(candidate),
			execute: func(_ *gocui.Gui) actionsPopupActionResult {
				return program.toggleAssigneePickerSelection(candidate)
			},
		})
	}
	return actions
}

func assigneePickerKeywords(candidate githubcli.PullRequestAuthor) []string {
	keywords := []string{strings.TrimSpace(candidate.Login)}
	if trimmedName := strings.TrimSpace(candidate.Name); trimmedName != "" {
		keywords = append(keywords, trimmedName)
	}
	return keywords
}

func (program *Program) assigneePickerLabel(candidate githubcli.PullRequestAuthor) string {
	if !program.assigneePickerVisible() {
		return ""
	}
	checkbox := "[ ]"
	trimmedLogin := strings.TrimSpace(candidate.Login)
	if program.assigneePicker.selectedLogins[trimmedLogin] {
		checkbox = "[x]"
	}
	trimmedName := strings.TrimSpace(candidate.Name)
	if trimmedName == "" || trimmedName == trimmedLogin {
		return checkbox + " @" + trimmedLogin
	}
	return checkbox + " @" + trimmedLogin + " (" + trimmedName + ")"
}

func (program *Program) toggleAssigneePickerSelection(candidate githubcli.PullRequestAuthor) actionsPopupActionResult {
	if !program.assigneePickerVisible() {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	trimmedLogin := strings.TrimSpace(candidate.Login)
	if trimmedLogin == "" {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if program.assigneePicker.selectedLogins[trimmedLogin] {
		delete(program.assigneePicker.selectedLogins, trimmedLogin)
	} else {
		program.assigneePicker.selectedLogins[trimmedLogin] = true
	}
	return actionsPopupActionResult{}
}

func (program *Program) executeSubmitAssigneePickerAction(_ *gocui.Gui) actionsPopupActionResult {
	if !program.assigneePickerVisible() {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if program.githubLoader == nil {
		return actionsPopupActionResult{err: errors.New("github loader is unavailable")}
	}

	addLogins, removeLogins := program.assigneePicker.selectedDiff()
	if len(addLogins) == 0 && len(removeLogins) == 0 {
		return actionsPopupActionResult{closePopup: true}
	}
	if err := program.githubLoader.UpdatePullRequestAssignees(program.assigneePicker.target.repository, program.assigneePicker.target.number, addLogins, removeLogins); err != nil {
		return actionsPopupActionResult{err: normalizedAssigneePickerError(err)}
	}

	program.invalidatePullRequestDetail(program.assigneePicker.target.repository, program.assigneePicker.target.number)
	program.setFeedback(program.model.Focus(), pullRequestAssigneesUpdatedSuccessMessage)
	return actionsPopupActionResult{closePopup: true}
}

func (state *assigneePickerState) selectedDiff() ([]string, []string) {
	if state == nil {
		return nil, nil
	}

	addLogins := make([]string, 0)
	removeLogins := make([]string, 0)
	for _, candidate := range state.candidates {
		trimmedLogin := strings.TrimSpace(candidate.Login)
		if trimmedLogin == "" {
			continue
		}
		selectedNow := state.selectedLogins[trimmedLogin]
		selectedOriginally := state.originalSelectedLogins[trimmedLogin]
		switch {
		case selectedNow && !selectedOriginally:
			addLogins = append(addLogins, trimmedLogin)
		case selectedOriginally && !selectedNow:
			removeLogins = append(removeLogins, trimmedLogin)
		}
	}
	return addLogins, removeLogins
}

func normalizedAssigneePickerCandidate(candidate githubcli.PullRequestAuthor) githubcli.PullRequestAuthor {
	candidate.Login = strings.TrimSpace(candidate.Login)
	candidate.Name = strings.TrimSpace(candidate.Name)
	return candidate
}

func normalizedAssigneePickerError(err error) error {
	if err == nil {
		return nil
	}

	message := strings.TrimSpace(err.Error())
	message = strings.TrimPrefix(message, "run `gh pr edit`:")
	message = strings.TrimSpace(message)
	if strings.HasPrefix(message, "exit status ") {
		if separatorIndex := strings.Index(message, ":"); separatorIndex >= 0 {
			message = strings.TrimSpace(message[separatorIndex+1:])
		}
	}
	if message == "" {
		return err
	}
	return errors.New(message)
}
