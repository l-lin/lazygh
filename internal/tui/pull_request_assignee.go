package tui

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const (
	assignPullRequestActionTitle              = "Assign PR"
	assigneePickerTitle                       = "Select assignees"
	assigneePickerSearchFooterHint            = "Type to search"
	pullRequestAssigneesUpdatedSuccessMessage = "PR assignees updated"
	assigneePickerSearchResultLimit           = 20
	defaultAssigneePickerSearchDebounceDelay  = 250 * time.Millisecond
)

type pullRequestAssigneePickerTarget struct {
	repository string
	number     int
	assignees  []githubdomain.PullRequestAuthor
}

type assigneePickerState struct {
	target                 pullRequestAssigneePickerTarget
	selectedLogins         map[string]bool
	originalSelectedLogins map[string]bool
	knownCandidates        map[string]githubdomain.PullRequestAuthor
	viewerLogin            string
	viewerName             string
	searchQuery            string
	searchResults          []githubdomain.PullRequestAuthor
	searchLoading          bool
	searchCommand          string
	searchRequestID        int
}

type assigneePickerLoadState struct {
	target  pullRequestAssigneePickerTarget
	command string
}

type assigneePickerCandidateSections struct {
	pinned        []githubdomain.PullRequestAuthor
	searchResults []githubdomain.PullRequestAuthor
}

func (sections assigneePickerCandidateSections) visibleCandidates() []githubdomain.PullRequestAuthor {
	visible := append([]githubdomain.PullRequestAuthor(nil), sections.pinned...)
	visible = append(visible, sections.searchResults...)
	return visible
}

func (program *Program) assigneePickerVisible() bool {
	return program.actionsPopupWidget.assigneePicker != nil
}

func (program *Program) assigneePickerLoading() bool {
	return program != nil && program.actionsPopupWidget.assigneePicker != nil && program.actionsPopupWidget.assigneePicker.searchLoading
}

func (program *Program) currentAssignPullRequestAction() (actionsPopupAction, bool) {
	if _, ok := program.selectedPullRequestAssigneePickerTarget(); !ok {
		return actionsPopupAction{}, false
	}
	return program.assignPullRequestAction(), true
}

func (program *Program) assignPullRequestAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	target, ok := program.selectedPullRequestAssigneePickerTarget()
	if ok {
		if !program.hasPullRequestMutations() {
			requested = actionsPopupErrorRequested(errors.New("github loader is unavailable"))
		} else {
			requested = MsgOpenAssigneePickerRequested{Target: target}
		}
	}
	return actionsPopupAction{
		id:        "assign-pull-request",
		title:     assignPullRequestActionTitle,
		icon:      actionsPopupEditPullRequestIcon,
		requested: requested,
	}
}

func (program *Program) currentAssigneePickerActionCount() int {
	actionCount := len(program.currentActionsPopupActions())
	if actionCount > 0 {
		return actionCount
	}
	return 1
}

func newAssigneePickerState(target pullRequestAssigneePickerTarget, viewerLogin string, viewerName string) *assigneePickerState {
	selectedLogins := map[string]bool{}
	knownCandidates := map[string]githubdomain.PullRequestAuthor{}
	for _, assignee := range target.assignees {
		normalizedCandidate := normalizedAssigneePickerCandidate(assignee)
		if normalizedCandidate.Login == "" {
			continue
		}
		selectedLogins[normalizedCandidate.Login] = true
		knownCandidates[normalizedCandidate.Login] = normalizedCandidate
	}

	viewerLogin = strings.TrimSpace(viewerLogin)
	viewerName = strings.TrimSpace(viewerName)
	if viewerLogin != "" {
		knownCandidates[viewerLogin] = normalizedAssigneePickerCandidate(githubdomain.PullRequestAuthor{Login: viewerLogin, Name: viewerName})
	}

	originalSelectedLogins := map[string]bool{}
	for login, selected := range selectedLogins {
		originalSelectedLogins[login] = selected
	}

	return &assigneePickerState{
		target:                 target,
		selectedLogins:         selectedLogins,
		originalSelectedLogins: originalSelectedLogins,
		knownCandidates:        knownCandidates,
		viewerLogin:            viewerLogin,
		viewerName:             viewerName,
		searchQuery:            "",
		searchResults:          nil,
	}
}

func (program *Program) resetAssigneePickerSearch(query string) int {
	if !program.assigneePickerVisible() {
		return 0
	}

	trimmedQuery := strings.TrimSpace(query)
	program.actionsPopupWidget.assigneePicker.searchRequestID++
	program.actionsPopupWidget.assigneePicker.searchQuery = trimmedQuery
	program.actionsPopupWidget.assigneePicker.searchResults = nil
	program.actionsPopupWidget.assigneePicker.searchLoading = false
	program.actionsPopupWidget.assigneePicker.searchCommand = ""
	return program.actionsPopupWidget.assigneePicker.searchRequestID
}

func (program *Program) markAssigneePickerSearchLoading(query string) {
	if !program.assigneePickerVisible() {
		return
	}

	program.actionsPopupWidget.assigneePicker.searchLoading = true
	program.actionsPopupWidget.assigneePicker.searchCommand = formatAssigneeSearchCommand(program.actionsPopupWidget.assigneePicker.target.repository, query)
}

func (program *Program) assigneePickerSearchRequestCurrent(requestID int, query string) bool {
	if !program.assigneePickerVisible() {
		return false
	}
	if program.actionsPopupWidget.assigneePicker.searchRequestID != requestID {
		return false
	}
	return strings.TrimSpace(program.model.ActionsPopupSearchQuery()) == strings.TrimSpace(query)
}

func (state *assigneePickerState) rememberCandidates(candidates []githubdomain.PullRequestAuthor) {
	if state == nil {
		return
	}
	if state.knownCandidates == nil {
		state.knownCandidates = map[string]githubdomain.PullRequestAuthor{}
	}

	for _, candidate := range candidates {
		normalizedCandidate := normalizedAssigneePickerCandidate(candidate)
		if normalizedCandidate.Login == "" {
			continue
		}
		existing := state.knownCandidates[normalizedCandidate.Login]
		if normalizedCandidate.Name == "" {
			normalizedCandidate.Name = existing.Name
		}
		if normalizedCandidate.Name == "" && normalizedCandidate.Login == state.viewerLogin {
			normalizedCandidate.Name = state.viewerName
		}
		state.knownCandidates[normalizedCandidate.Login] = normalizedCandidate
		if normalizedCandidate.Login == state.viewerLogin && normalizedCandidate.Name != "" {
			state.viewerName = normalizedCandidate.Name
		}
	}
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
		assignees:  append([]githubdomain.PullRequestAuthor(nil), result.detail.Assignees...),
	}, true
}

func (program *Program) pullRequestAssigneeStateVisible() bool {
	actionContext := program.actionContext()
	if !actionContext.IsPullRequestContext() || !actionContext.ShowsPullRequestDescription() {
		return false
	}
	if actionContext.IsReviewContext() {
		return actionContext.ActiveView.Focus == FocusDetailView
	}
	return actionContext.ActiveView.Focus == FocusPullRequestsView || actionContext.ActiveView.Focus == FocusDetailView
}

func (program *Program) currentAssigneePickerActions() []actionsPopupAction {
	return program.currentAssigneePickerActionsForQuery(program.model.ActionsPopupSearchQuery())
}

func (program *Program) currentAssigneePickerActionsForQuery(query string) []actionsPopupAction {
	if !program.assigneePickerVisible() {
		return nil
	}

	candidates := program.currentAssigneePickerVisibleCandidatesForQuery(query)
	actions := make([]actionsPopupAction, 0, len(candidates))
	for _, candidate := range candidates {
		candidate := candidate
		actions = append(actions, actionsPopupAction{
			id:        "assignee-" + strings.ToLower(strings.TrimSpace(candidate.Login)),
			title:     program.assigneePickerLabel(candidate),
			keywords:  program.assigneePickerSearchKeywords(candidate),
			requested: MsgToggleAssigneePickerSelectionRequested{Candidate: candidate},
		})
	}
	return actions
}

func (program *Program) currentAssigneePickerVisibleCandidates() []githubdomain.PullRequestAuthor {
	return program.currentAssigneePickerVisibleCandidatesForQuery(program.model.ActionsPopupSearchQuery())
}

func (program *Program) currentAssigneePickerVisibleCandidatesForQuery(query string) []githubdomain.PullRequestAuthor {
	return program.currentAssigneePickerCandidateSections(query).visibleCandidates()
}

func (program *Program) currentAssigneePickerSearchResultCandidatesForQuery(query string) []githubdomain.PullRequestAuthor {
	return program.currentAssigneePickerCandidateSections(query).searchResults
}

func (program *Program) currentAssigneePickerCandidateSections(query string) assigneePickerCandidateSections {
	if !program.assigneePickerVisible() {
		return assigneePickerCandidateSections{}
	}

	pinnedCandidates := program.currentPinnedAssigneePickerCandidates()
	return assigneePickerCandidateSections{
		pinned:        pinnedCandidates,
		searchResults: program.actionsPopupWidget.assigneePicker.searchResultCandidatesForQuery(query, pinnedCandidates),
	}
}

func (state *assigneePickerState) searchResultCandidatesForQuery(query string, pinnedCandidates []githubdomain.PullRequestAuthor) []githubdomain.PullRequestAuthor {
	if state == nil {
		return nil
	}

	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" || strings.TrimSpace(state.searchQuery) != trimmedQuery {
		return nil
	}

	seenLogins := map[string]bool{}
	for _, candidate := range pinnedCandidates {
		trimmedLogin := strings.TrimSpace(candidate.Login)
		if trimmedLogin == "" {
			continue
		}
		seenLogins[trimmedLogin] = true
	}

	searchResultCandidates := make([]githubdomain.PullRequestAuthor, 0, len(state.searchResults))
	for _, candidate := range state.searchResults {
		normalizedCandidate := normalizedAssigneePickerCandidate(candidate)
		if normalizedCandidate.Login == "" || seenLogins[normalizedCandidate.Login] {
			continue
		}
		seenLogins[normalizedCandidate.Login] = true
		searchResultCandidates = append(searchResultCandidates, normalizedCandidate)
	}
	return searchResultCandidates
}

func (program *Program) matchingAssigneePickerIndexes(query string) []int {
	return actionIndexes(len(program.currentAssigneePickerActionsForQuery(query)))
}

func (program *Program) currentPinnedAssigneePickerCandidates() []githubdomain.PullRequestAuthor {
	if !program.assigneePickerVisible() {
		return nil
	}

	logins := make([]string, 0, len(program.actionsPopupWidget.assigneePicker.selectedLogins)+1)
	for login := range program.actionsPopupWidget.assigneePicker.selectedLogins {
		trimmedLogin := strings.TrimSpace(login)
		if trimmedLogin == "" {
			continue
		}
		logins = append(logins, trimmedLogin)
	}
	if viewerLogin := strings.TrimSpace(program.actionsPopupWidget.assigneePicker.viewerLogin); viewerLogin != "" && !program.actionsPopupWidget.assigneePicker.selectedLogins[viewerLogin] {
		logins = append(logins, viewerLogin)
	}
	if len(logins) == 0 {
		return nil
	}

	candidates := make([]githubdomain.PullRequestAuthor, 0, len(logins))
	seenLogins := map[string]bool{}
	for _, login := range logins {
		trimmedLogin := strings.TrimSpace(login)
		if trimmedLogin == "" || seenLogins[trimmedLogin] {
			continue
		}
		seenLogins[trimmedLogin] = true
		candidates = append(candidates, program.actionsPopupWidget.assigneePicker.candidateForLogin(trimmedLogin))
	}
	sortAssigneePickerCandidates(candidates, program.actionsPopupWidget.assigneePicker.selectedLogins, program.actionsPopupWidget.assigneePicker.viewerLogin)
	return candidates
}

func (state *assigneePickerState) candidateForLogin(login string) githubdomain.PullRequestAuthor {
	trimmedLogin := strings.TrimSpace(login)
	if trimmedLogin == "" {
		return githubdomain.PullRequestAuthor{}
	}

	candidate := normalizedAssigneePickerCandidate(state.knownCandidates[trimmedLogin])
	candidate.Login = trimmedLogin
	if candidate.Name == "" && trimmedLogin == state.viewerLogin {
		candidate.Name = strings.TrimSpace(state.viewerName)
	}
	return candidate
}

func sortAssigneePickerCandidates(candidates []githubdomain.PullRequestAuthor, selectedLogins map[string]bool, viewerLogin string) {
	sort.SliceStable(candidates, func(i int, j int) bool {
		left := candidates[i]
		right := candidates[j]
		leftPriority := assigneePickerCandidatePriority(strings.TrimSpace(left.Login), selectedLogins, viewerLogin)
		rightPriority := assigneePickerCandidatePriority(strings.TrimSpace(right.Login), selectedLogins, viewerLogin)
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		return strings.TrimSpace(left.Login) < strings.TrimSpace(right.Login)
	})
}

func assigneePickerCandidatePriority(login string, selectedLogins map[string]bool, viewerLogin string) int {
	switch {
	case login != "" && login == viewerLogin:
		return 0
	case selectedLogins[login]:
		return 1
	default:
		return 2
	}
}

func (program *Program) assigneePickerSearchKeywords(candidate githubdomain.PullRequestAuthor) []string {
	if !program.assigneePickerVisible() {
		return nil
	}

	trimmedLogin := strings.TrimSpace(candidate.Login)
	trimmedName := strings.TrimSpace(candidate.Name)
	keywords := filterEmptyStrings([]string{trimmedLogin, "@" + trimmedLogin, trimmedName})
	if trimmedLogin != "" && trimmedLogin == program.actionsPopupWidget.assigneePicker.viewerLogin {
		keywords = append(keywords, "@me")
	}
	return keywords
}

func (program *Program) assigneePickerLabel(candidate githubdomain.PullRequestAuthor) string {
	if !program.assigneePickerVisible() {
		return ""
	}

	checkbox := iconCheckboxUnchecked
	trimmedLogin := strings.TrimSpace(candidate.Login)
	if program.actionsPopupWidget.assigneePicker.selectedLogins[trimmedLogin] {
		checkbox = iconCheckboxChecked
	}

	identityLabel := "@" + trimmedLogin
	if trimmedLogin != "" && trimmedLogin == program.actionsPopupWidget.assigneePicker.viewerLogin {
		identityLabel = "@me"
	}
	trimmedName := strings.TrimSpace(candidate.Name)
	if trimmedName == "" && trimmedLogin == program.actionsPopupWidget.assigneePicker.viewerLogin {
		trimmedName = strings.TrimSpace(program.actionsPopupWidget.assigneePicker.viewerName)
	}
	if trimmedName == "" && trimmedLogin == program.actionsPopupWidget.assigneePicker.viewerLogin {
		trimmedName = trimmedLogin
	}
	if trimmedName == "" || trimmedName == trimmedLogin {
		if identityLabel == "@me" {
			return checkbox + " " + identityLabel + " (" + trimmedLogin + ")"
		}
		return checkbox + " " + identityLabel
	}
	return checkbox + " " + identityLabel + " (" + trimmedName + ")"
}

func (program *Program) submitAssigneePickerSelection(gui *gocui.Gui) error {
	if !program.assigneePickerVisible() {
		return errActionsPopupActionUnavailable
	}
	if !program.hasPullRequestMutations() {
		return errors.New("github loader is unavailable")
	}

	repository := program.actionsPopupWidget.assigneePicker.target.repository
	number := program.actionsPopupWidget.assigneePicker.target.number
	addLogins, removeLogins := program.actionsPopupWidget.assigneePicker.selectedDiff()
	return program.dispatch(gui, MsgSubmitAssigneePickerRequested{
		Repository:   repository,
		Number:       number,
		AddLogins:    addLogins,
		RemoveLogins: removeLogins,
	})
}

func updatePullRequestAssigneesCommand(repository string, number int, addLogins []string, removeLogins []string) string {
	arguments := []string{"gh", "pr", "edit", fmt.Sprintf("%d", number), "-R", repository}
	if len(addLogins) > 0 {
		arguments = append(arguments, "--add-assignee", strings.Join(addLogins, ","))
	}
	if len(removeLogins) > 0 {
		arguments = append(arguments, "--remove-assignee", strings.Join(removeLogins, ","))
	}
	return formatStatusLineCommand(arguments...)
}

func (program *Program) optimisticallyUpdatePullRequestAssignees(repository string, number int, addLogins []string, removeLogins []string) {
	if program == nil {
		return
	}

	key := pullRequestMutationCacheKey(repository, number)
	if key == "" {
		return
	}

	identity := githubdomain.PullRequest{Repository: githubdomain.Repository{NameWithOwner: strings.TrimSpace(repository)}, Number: number}
	result, ok := program.pullRequestDetailCache[key]
	if !ok || result.err != nil {
		result = pullRequestDetailResult{detail: program.optimisticPullRequestDetailSeed(identity)}
	}

	updatedAssignees := append([]githubdomain.PullRequestAuthor(nil), result.detail.Assignees...)
	for _, login := range removeLogins {
		trimmedLogin := strings.TrimSpace(login)
		if trimmedLogin == "" {
			continue
		}
		filteredAssignees := updatedAssignees[:0]
		for _, assignee := range updatedAssignees {
			if strings.TrimSpace(assignee.Login) == trimmedLogin {
				continue
			}
			filteredAssignees = append(filteredAssignees, assignee)
		}
		updatedAssignees = filteredAssignees
	}
	for _, login := range addLogins {
		trimmedLogin := strings.TrimSpace(login)
		if trimmedLogin == "" {
			continue
		}
		alreadyAssigned := false
		for _, assignee := range updatedAssignees {
			if strings.TrimSpace(assignee.Login) == trimmedLogin {
				alreadyAssigned = true
				break
			}
		}
		if alreadyAssigned {
			continue
		}
		candidate := githubdomain.PullRequestAuthor{Login: trimmedLogin}
		if program.assigneePickerVisible() {
			candidate = program.actionsPopupWidget.assigneePicker.candidateForLogin(trimmedLogin)
		}
		updatedAssignees = append(updatedAssignees, normalizedAssigneePickerCandidate(candidate))
	}

	result.detail.Assignees = updatedAssignees
	result.sourceUpdatedAt = ""
	result.needsRefresh = true
	program.pullRequestDetailCache[key] = result
	program.invalidatePullRequestDetailDocumentCache()
	program.invalidatePersistentPullRequest(repository, number)
}

func (state *assigneePickerState) selectedDiff() ([]string, []string) {
	if state == nil {
		return nil, nil
	}

	addLogins := make([]string, 0)
	removeLogins := make([]string, 0)
	for login := range state.selectedLogins {
		trimmedLogin := strings.TrimSpace(login)
		if trimmedLogin == "" || state.originalSelectedLogins[trimmedLogin] {
			continue
		}
		addLogins = append(addLogins, trimmedLogin)
	}
	for login := range state.originalSelectedLogins {
		trimmedLogin := strings.TrimSpace(login)
		if trimmedLogin == "" || state.selectedLogins[trimmedLogin] {
			continue
		}
		removeLogins = append(removeLogins, trimmedLogin)
	}
	sort.Strings(addLogins)
	sort.Strings(removeLogins)
	return addLogins, removeLogins
}

func normalizedAssigneePickerCandidate(candidate githubdomain.PullRequestAuthor) githubdomain.PullRequestAuthor {
	candidate.Login = strings.TrimSpace(candidate.Login)
	candidate.Name = strings.TrimSpace(candidate.Name)
	return candidate
}

func normalizedAssigneePickerError(err error) error {
	return normalizeGHCommandError(err)
}
