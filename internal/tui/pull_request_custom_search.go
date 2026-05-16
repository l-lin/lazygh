package tui

import (
	"errors"
	"strconv"
	"strings"

	"github.com/jesseduffield/gocui"

	appconfig "github.com/l-lin/lazygh/internal/config"
)

const (
	pullRequestCustomSearchLabel        = "Custom"
	pullRequestCustomSearchActionTitle  = "Custom search"
	pullRequestCustomSearchEditorTitle  = "Search PRs"
	pullRequestCustomSearchEditorHeight = lineModalEditorTotalHeight
)

func (program *Program) openPullRequestCustomSearch(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if program.mainPaneActionBlocked() || program.actionContext().IsReviewContext() {
		return nil
	}

	return program.openPullRequestCustomSearchEditor(gui)
}

func (program *Program) openPullRequestCustomSearchEditor(gui *gocui.Gui) error {
	actualErr := program.openLineModalEditorWithHeight(gui, pullRequestCustomSearchEditorTitle, program.currentPullRequestSearchCriteria(), program.submitPullRequestCustomSearch, pullRequestCustomSearchEditorHeight)
	if actualErr != nil {
		return actualErr
	}
	if program.modalEditor != nil {
		program.modalEditor.afterSubmit = func(gui *gocui.Gui) {
			program.reloadActivePullRequestsTab(gui)
		}
	}
	return nil
}

func (program *Program) pullRequestCustomSearchActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "custom-pull-request-search",
		title:   pullRequestCustomSearchActionTitle,
		icon:    actionsPopupCustomSearchIcon,
		execute: program.executeOpenPullRequestCustomSearchAction,
	}
}

func (program *Program) executeOpenPullRequestCustomSearchAction(gui *gocui.Gui) actionsPopupActionResult {
	wasVisible := program.modalEditorVisible()
	if err := program.openPullRequestCustomSearchEditor(gui); err != nil {
		return actionsPopupActionResult{err: err}
	}
	if !wasVisible && program.modalEditorVisible() {
		return actionsPopupActionResult{closePopup: true}
	}
	return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
}

func (program *Program) currentPullRequestSearchCriteria() string {
	return formatPullRequestSearchCriteria(program.pullRequestSearch(program.model.ActivePullRequestTab()).Command)
}

func (program *Program) submitPullRequestCustomSearch(criteria string) error {
	command := pullRequestCustomSearchCommand(criteria)
	if len(command) == 0 {
		return errors.New("search criteria cannot be empty")
	}

	program.upsertPullRequestCustomSearch(appconfig.PullRequestSearch{Label: pullRequestCustomSearchLabel, Command: command})
	return nil
}

func (program *Program) upsertPullRequestCustomSearch(search appconfig.PullRequestSearch) PullRequestTab {
	searches := append([]appconfig.PullRequestSearch(nil), appconfig.ResolvePullRequestSearches(program.pullRequestSearches)...)
	customTab, customTabExists := pullRequestCustomSearchTab(searches)
	preservedRows := make(map[PullRequestTab][]PullRequestRow, len(program.model.PullRequestTabs()))
	for _, tab := range program.model.PullRequestTabs() {
		if customTabExists && tab == customTab {
			continue
		}
		preservedRows[tab] = program.model.PullRequestRows(tab)
	}

	if customTabExists {
		searches[customTab] = search
	} else {
		customTab = PullRequestTab(len(searches))
		searches = append(searches, search)
	}

	program.pullRequestSearches = searches
	program.model.SetPullRequestTabs(pullRequestTabSeedsForSearches(searches))
	for tab, rows := range preservedRows {
		program.model.SetPullRequestRows(tab, rows)
	}
	program.model.SetPullRequestRows(customTab, []PullRequestRow{{Item: program.pullRequestLoadingItem(customTab)}})
	program.model.pullRequestSearchQueries[customTab] = ""
	program.setPullRequestsLoadStarted(customTab, false)
	program.setPullRequestsLoading(customTab, false)
	program.setPullRequestsCount(customTab, 0, false)

	state := program.model.ScreenState().WithViewTabs(sidePanelPullRequestsViewNumber, int(customTab), program.model.pullRequestScreenTabs()).FocusViewNumber(sidePanelPullRequestsViewNumber)
	program.model.applyBrowserScreenState(state)
	return customTab
}

func pullRequestCustomSearchTab(searches []appconfig.PullRequestSearch) (PullRequestTab, bool) {
	for index, search := range searches {
		if strings.TrimSpace(search.Label) == pullRequestCustomSearchLabel {
			return PullRequestTab(index), true
		}
	}
	return 0, false
}

func pullRequestCustomSearchCommand(criteria string) []string {
	arguments := appconfig.ParseCommandLine(criteria)
	arguments = stripPullRequestSearchCommandArguments(arguments)
	if len(arguments) == 0 {
		return nil
	}

	command := make([]string, 0, len(arguments)+2)
	command = append(command, "search", "prs")
	command = append(command, arguments...)
	return command
}

func formatPullRequestSearchCriteria(command []string) string {
	arguments := stripPullRequestSearchCommandArguments(command)
	if len(arguments) == 0 {
		return ""
	}

	formattedArguments := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		formattedArguments = append(formattedArguments, formatCommandLineArgument(argument))
	}
	return strings.Join(formattedArguments, " ")
}

func stripPullRequestSearchCommandArguments(arguments []string) []string {
	if len(arguments) == 0 {
		return nil
	}

	trimmedArguments := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		trimmedArgument := strings.TrimSpace(argument)
		if trimmedArgument == "" {
			continue
		}
		trimmedArguments = append(trimmedArguments, trimmedArgument)
	}
	if len(trimmedArguments) == 0 {
		return nil
	}
	if len(trimmedArguments) >= 3 && strings.EqualFold(trimmedArguments[0], "gh") && strings.EqualFold(trimmedArguments[1], "search") && strings.EqualFold(trimmedArguments[2], "prs") {
		trimmedArguments = trimmedArguments[3:]
	} else if len(trimmedArguments) >= 2 && ((strings.EqualFold(trimmedArguments[0], "search") && strings.EqualFold(trimmedArguments[1], "prs")) || (strings.EqualFold(trimmedArguments[0], "pr") && strings.EqualFold(trimmedArguments[1], "list"))) {
		trimmedArguments = trimmedArguments[2:]
	}

	filteredArguments := make([]string, 0, len(trimmedArguments))
	for index := 0; index < len(trimmedArguments); index++ {
		argument := trimmedArguments[index]
		switch {
		case argument == "--json":
			index++
		case strings.HasPrefix(argument, "--json="):
			continue
		default:
			filteredArguments = append(filteredArguments, argument)
		}
	}
	if len(filteredArguments) == 0 {
		return nil
	}
	return filteredArguments
}

func formatCommandLineArgument(argument string) string {
	trimmedArgument := strings.TrimSpace(argument)
	if trimmedArgument == "" {
		return strconv.Quote("")
	}
	if !strings.ContainsAny(trimmedArgument, " \t\n\r\"\\") {
		return trimmedArgument
	}
	return strconv.Quote(trimmedArgument)
}
