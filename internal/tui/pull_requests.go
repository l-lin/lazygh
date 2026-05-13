package tui

import (
	"fmt"
	"strings"

	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/theme"
)

const (
	myPullRequestsLoadingTitle                   = "Loading my pull requests..."
	myPullRequestsEmptyTitle                     = "No open pull requests"
	myPullRequestsEmptyDetail                    = "GitHub returned no open pull requests authored by the authenticated user."
	myPullRequestsUnauthenticatedTitle           = "GitHub authentication required"
	myPullRequestsUnauthenticatedDetail          = "GitHub CLI is not authenticated.\n\nRun `gh auth login`, then restart `lazygh`."
	myPullRequestsUnavailableTitle               = "`gh` not found"
	myPullRequestsUnavailableDetail              = "Install GitHub CLI and make sure `gh` is in your `PATH`, then restart `lazygh`."
	myPullRequestsGenericErrorTitle              = "Could not load my pull requests"
	requestedPullRequestsLoadingTitle            = "Loading my reviews..."
	requestedPullRequestsEmptyTitle              = "No reviewed pull requests"
	requestedPullRequestsEmptyDetail             = "GitHub returned no open pull requests reviewed by the authenticated user."
	requestedPullRequestsUnauthenticatedTitle    = "GitHub authentication required"
	requestedPullRequestsUnauthenticatedDetail   = "GitHub CLI is not authenticated.\n\nRun `gh auth login`, then restart `lazygh`."
	requestedPullRequestsUnavailableTitle        = "`gh` not found"
	requestedPullRequestsUnavailableDetail       = "Install GitHub CLI and make sure `gh` is in your `PATH`, then restart `lazygh`."
	requestedPullRequestsGenericErrorTitle       = "Could not load my reviews"
	reviewRequestedPullRequestsLoadingTitle      = "Loading review requests..."
	reviewRequestedPullRequestsEmptyTitle        = "No requested pull requests"
	reviewRequestedPullRequestsEmptyDetail       = "GitHub returned no open pull requests requesting review from the authenticated user."
	reviewRequestedPullRequestsGenericErrorTitle = "Could not load review requests"
)

type pullRequestListState struct {
	loadingTitle          string
	loadingDetail         string
	emptyTitle            string
	emptyDetail           string
	unauthenticatedTitle  string
	unauthenticatedDetail string
	unavailableTitle      string
	unavailableDetail     string
	genericErrorTitle     string
	genericErrorPrefix    string
}

var (
	defaultPullRequestSearches              = appconfig.DefaultPullRequestSearches()
	myPullRequestsLoadingDetail             = buildPullRequestListState(defaultPullRequestSearches[0]).loadingDetail
	myPullRequestsGenericErrorPrefix        = buildPullRequestListState(defaultPullRequestSearches[0]).genericErrorPrefix
	requestedPullRequestsLoadingDetail      = buildPullRequestListState(defaultPullRequestSearches[1]).loadingDetail
	requestedPullRequestsGenericErrorPrefix = buildPullRequestListState(defaultPullRequestSearches[1]).genericErrorPrefix
	myPullRequestsState                     = buildPullRequestListState(defaultPullRequestSearches[0])
	requestedPullRequestsState              = buildPullRequestListState(defaultPullRequestSearches[1])
)

func buildPullRequestListState(search appconfig.PullRequestSearch) pullRequestListState {
	commandLine := formatPullRequestSearchCommand(search.Command)
	label := strings.TrimSpace(search.Label)
	switch pullRequestSearchKind(search.Command) {
	case pullRequestSearchKindReviewRequested:
		return newPullRequestListState(
			reviewRequestedPullRequestsLoadingTitle,
			fmt.Sprintf("Running `%s` to load review requests.", commandLine),
			reviewRequestedPullRequestsEmptyTitle,
			reviewRequestedPullRequestsEmptyDetail,
			reviewRequestedPullRequestsGenericErrorTitle,
			commandLine,
		)
	case pullRequestSearchKindReviewed:
		return newPullRequestListState(
			requestedPullRequestsLoadingTitle,
			fmt.Sprintf("Running `%s` to load reviewed pull requests.", commandLine),
			requestedPullRequestsEmptyTitle,
			requestedPullRequestsEmptyDetail,
			requestedPullRequestsGenericErrorTitle,
			commandLine,
		)
	case pullRequestSearchKindAuthored:
		return newPullRequestListState(
			myPullRequestsLoadingTitle,
			fmt.Sprintf("Running `%s` to load authored pull requests.", commandLine),
			myPullRequestsEmptyTitle,
			myPullRequestsEmptyDetail,
			myPullRequestsGenericErrorTitle,
			commandLine,
		)
	default:
		return newPullRequestListState(
			fmt.Sprintf("Loading %s...", label),
			fmt.Sprintf("Running `%s` to load pull requests for %s.", commandLine, label),
			fmt.Sprintf("No pull requests for %s", label),
			fmt.Sprintf("GitHub returned no open pull requests for %s.", label),
			fmt.Sprintf("Could not load %s", label),
			commandLine,
		)
	}
}

func newPullRequestListState(loadingTitle string, loadingDetail string, emptyTitle string, emptyDetail string, genericErrorTitle string, commandLine string) pullRequestListState {
	return pullRequestListState{
		loadingTitle:          loadingTitle,
		loadingDetail:         loadingDetail,
		emptyTitle:            emptyTitle,
		emptyDetail:           emptyDetail,
		unauthenticatedTitle:  requestedPullRequestsUnauthenticatedTitle,
		unauthenticatedDetail: requestedPullRequestsUnauthenticatedDetail,
		unavailableTitle:      requestedPullRequestsUnavailableTitle,
		unavailableDetail:     requestedPullRequestsUnavailableDetail,
		genericErrorTitle:     genericErrorTitle,
		genericErrorPrefix:    fmt.Sprintf("Failed to run `%s`.", commandLine),
	}
}

type pullRequestSearchKindType int

const (
	pullRequestSearchKindGeneric pullRequestSearchKindType = iota
	pullRequestSearchKindAuthored
	pullRequestSearchKindReviewed
	pullRequestSearchKindReviewRequested
)

func pullRequestSearchKind(command []string) pullRequestSearchKindType {
	if pullRequestSearchMatches(command, "--review-requested", "review-requested:") {
		return pullRequestSearchKindReviewRequested
	}
	if pullRequestSearchMatches(command, "--reviewed-by", "reviewed-by:") {
		return pullRequestSearchKindReviewed
	}
	if pullRequestSearchMatches(command, "--author", "author:") {
		return pullRequestSearchKindAuthored
	}
	return pullRequestSearchKindGeneric
}

func pullRequestSearchMatches(command []string, flag string, queryPrefix string) bool {
	normalizedCommand := normalizedPullRequestSearchCommand(command)
	for index, argument := range normalizedCommand {
		if argument == flag || strings.HasPrefix(argument, flag+"=") {
			return true
		}
		if argument == "--search" && index+1 < len(normalizedCommand) && pullRequestSearchQueryContains(normalizedCommand[index+1], queryPrefix) {
			return true
		}
		if strings.HasPrefix(argument, "--search=") && pullRequestSearchQueryContains(strings.TrimPrefix(argument, "--search="), queryPrefix) {
			return true
		}
	}
	return false
}

func normalizedPullRequestSearchCommand(command []string) []string {
	normalized := make([]string, 0, len(command))
	for _, argument := range command {
		trimmedArgument := strings.ToLower(strings.TrimSpace(argument))
		if trimmedArgument == "" {
			continue
		}
		normalized = append(normalized, trimmedArgument)
	}
	return normalized
}

func pullRequestSearchQueryContains(query string, prefix string) bool {
	for _, term := range strings.Fields(strings.TrimSpace(query)) {
		if strings.HasPrefix(term, prefix) {
			return true
		}
	}
	return false
}

func myPullRequestsLoadingItem() Item {
	return pullRequestLoadingItem(myPullRequestsState)
}

func requestedPullRequestsLoadingItem() Item {
	return pullRequestLoadingItem(requestedPullRequestsState)
}

func myPullRequestsStateRows(pullRequests []githubdomain.PullRequest, err error) []PullRequestRow {
	return pullRequestStateRows(myPullRequestsState, pullRequests, err)
}

func requestedPullRequestsStateRows(pullRequests []githubdomain.PullRequest, err error) []PullRequestRow {
	return pullRequestStateRows(requestedPullRequestsState, pullRequests, err)
}

func myPullRequestsStateItems(pullRequests []githubdomain.PullRequest, err error) []Item {
	return pullRequestItems(myPullRequestsStateRows(pullRequests, err))
}

func requestedPullRequestsStateItems(pullRequests []githubdomain.PullRequest, err error) []Item {
	return pullRequestItems(requestedPullRequestsStateRows(pullRequests, err))
}

func myPullRequestRow(pullRequest any) PullRequestRow {
	return pullRequestRow(pullRequest)
}

func requestedPullRequestRow(pullRequest any) PullRequestRow {
	return pullRequestRow(pullRequest)
}

func myPullRequestItem(pullRequest any) Item {
	return myPullRequestRow(pullRequest).Item
}

func requestedPullRequestItem(pullRequest any) Item {
	return requestedPullRequestRow(pullRequest).Item
}

func myPullRequestsErrorItem(err error) Item {
	return pullRequestErrorItem(myPullRequestsState, err)
}

func requestedPullRequestsErrorItem(err error) Item {
	return pullRequestErrorItem(requestedPullRequestsState, err)
}

func pullRequestLoadingItem(state pullRequestListState) Item {
	return Item{Title: state.loadingTitle, Detail: state.loadingDetail}
}

func pullRequestStateRows(state pullRequestListState, pullRequests []githubdomain.PullRequest, err error) []PullRequestRow {
	if err != nil {
		return []PullRequestRow{{Item: pullRequestErrorItem(state, err)}}
	}
	if len(pullRequests) == 0 {
		return []PullRequestRow{{Item: pullRequestEmptyItem(state)}}
	}

	rows := make([]PullRequestRow, 0, len(pullRequests))
	for _, pullRequest := range pullRequests {
		rows = append(rows, pullRequestRow(pullRequest))
	}
	return rows
}

func pullRequestRow(pullRequest any) PullRequestRow {
	pullRequestValue, ok := toDomainPullRequestSummary(pullRequest)
	if !ok {
		return PullRequestRow{}
	}
	repositoryName := pullRequestRepositoryName(pullRequestValue.Repository)
	body := strings.TrimSpace(pullRequestValue.Body)
	if body == "" {
		body = "No description available."
	}

	detailLines := []string{
		fmt.Sprintf("Repository: %s", repositoryName),
		fmt.Sprintf("Number: #%d", pullRequestValue.Number),
		fmt.Sprintf("State: %s", valueOrDash(pullRequestValue.State)),
		fmt.Sprintf("Draft: %s", yesNo(pullRequestValue.IsDraft)),
		fmt.Sprintf("Updated: %s", valueOrDash(pullRequestValue.UpdatedAt)),
		fmt.Sprintf("URL: %s", valueOrDash(pullRequestValue.URL)),
		"",
		body,
	}

	mergeChecksBackgroundHex := pullRequestMergeChecksBackgroundHex(pullRequestValue)
	mergeChecksBackgroundPrefix := backgroundColorEscape(mergeChecksBackgroundHex)
	status := effectivePullRequestStatus(pullRequestValue.State, pullRequestValue.IsDraft)
	statusIconSegment := ItemTitleSegment{Text: pullRequestStatusIcon(status) + " ", Prefix: mergeChecksBackgroundPrefix, BackgroundHex: mergeChecksBackgroundHex, MinimumContrast: 3.0}
	if statusStyle, ok := pullRequestStatusStyleFor(status); ok {
		statusIconSegment.ForegroundHex = statusStyle.foregroundHex
		statusIconSegment.Prefix = foregroundColorEscape(statusStyle.foregroundHex) + mergeChecksBackgroundPrefix
	}

	titlePrefix := fmt.Sprintf("%s#%d", repositoryName, pullRequestValue.Number)
	titleSuffix := " " + valueOrDash(pullRequestValue.Title)

	summaryCopy := pullRequestValue
	return PullRequestRow{
		Item: Item{
			Title:  statusIconSegment.Text + titlePrefix + titleSuffix,
			Detail: strings.Join(detailLines, "\n"),
			TitleSegments: []ItemTitleSegment{
				statusIconSegment,
				{Text: titlePrefix, Prefix: foregroundColorEscape(theme.PullRequestReferenceHex) + mergeChecksBackgroundPrefix, ForegroundHex: theme.PullRequestReferenceHex, BackgroundHex: mergeChecksBackgroundHex},
				{Text: titleSuffix, Prefix: foregroundColorEscape(theme.PullRequestTitleHex) + mergeChecksBackgroundPrefix, ForegroundHex: theme.PullRequestTitleHex, BackgroundHex: mergeChecksBackgroundHex},
			},
		},
		Summary: &summaryCopy,
	}
}

func pullRequestMergeChecksBackgroundHex(pullRequest githubdomain.PullRequest) string {
	switch pullRequestMergeChecksStatus(pullRequest) {
	case pullRequestOverviewStatusSuccess:
		return theme.SuccessBackgroundHex
	case pullRequestOverviewStatusFailure:
		return theme.FailureBackgroundHex
	default:
		return ""
	}
}

func (program *Program) restylePullRequestRows() {
	if program == nil || program.model == nil {
		return
	}

	for _, tab := range program.model.PullRequestTabs() {
		rows := program.model.PullRequestRows(tab)
		if len(rows) == 0 {
			continue
		}
		program.model.SetPullRequestRows(tab, restyledPullRequestRows(rows))
	}
}

func restyledPullRequestRows(rows []PullRequestRow) []PullRequestRow {
	restyledRows := make([]PullRequestRow, 0, len(rows))
	for _, row := range rows {
		if row.Summary == nil {
			restyledRows = append(restyledRows, row)
			continue
		}
		restyledRows = append(restyledRows, pullRequestRow(*row.Summary))
	}
	return restyledRows
}

func pullRequestMergeChecksStatus(pullRequest githubdomain.PullRequest) pullRequestOverviewStatus {
	entries := []pullRequestOverviewEntry{
		{Status: pullRequestMergeChecksReviewStatus(pullRequest)},
		{Status: pullRequestOverviewStatusForStatusCheckRollupState(pullRequest.StatusCheckRollupState)},
		{Status: pullRequestOverviewStatusForMergeability(pullRequest.Mergeable, pullRequest.MergeStateStatus)},
	}
	return pullRequestOverviewBlockStatus(entries)
}

func pullRequestMergeChecksReviewStatus(pullRequest githubdomain.PullRequest) pullRequestOverviewStatus {
	reviewDecisionStatus := pullRequestOverviewStatusForReviewDecision(pullRequest.ReviewDecision)
	if reviewDecisionStatus == pullRequestOverviewStatusFailure {
		return pullRequestOverviewStatusFailure
	}
	if len(pullRequest.ReviewRequests) > 0 {
		return pullRequestOverviewStatusPending
	}
	return reviewDecisionStatus
}

func pullRequestErrorItem(state pullRequestListState, err error) Item {
	switch {
	case isProviderUnauthenticatedError(err):
		return Item{Title: state.unauthenticatedTitle, Detail: state.unauthenticatedDetail}
	case isProviderUnavailableError(err):
		return Item{Title: state.unavailableTitle, Detail: state.unavailableDetail}
	default:
		return Item{Title: state.genericErrorTitle, Detail: formatPullRequestErrorDetail(state.genericErrorPrefix, err)}
	}
}

func pullRequestEmptyItem(state pullRequestListState) Item {
	return Item{Title: state.emptyTitle, Detail: state.emptyDetail}
}

func formatPullRequestErrorDetail(prefix string, err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return prefix
	}

	return fmt.Sprintf("%s\n\n%s", prefix, message)
}

func pullRequestRepositoryName(repository any) string {
	repositoryRef, ok := toDomainRepository(repository)
	if !ok {
		return "-"
	}
	if repositoryRef.NameWithOwner != "" {
		return repositoryRef.NameWithOwner
	}
	if repositoryRef.Name != "" {
		return repositoryRef.Name
	}
	return "-"
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
