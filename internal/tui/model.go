package tui

import "codeberg.org/l-lin/lazygh/internal/githubcli"

type Focus int

const (
	FocusUserView Focus = iota
	FocusPullRequestsView
	FocusDetailView
)

type PullRequestTab int

const (
	MyPullRequestsTab PullRequestTab = iota
	RequestedPullRequestsTab
)

type ItemTitleSegment struct {
	Text               string
	Prefix             string
	ForegroundHex      string
	BackgroundHex      string
	MinimumContrast    float64
	PreserveForeground bool
}

type Item struct {
	Title         string
	Detail        string
	TitleSegments []ItemTitleSegment
}

type PullRequestRow struct {
	Item    Item
	Summary *githubcli.PullRequest
}

type PullRequestTabSeed struct {
	Label        string
	PullRequests []Item
}

type SeedData struct {
	Users                 []Item
	MyPullRequests        []Item
	RequestedPullRequests []Item
	PullRequestTabs       []PullRequestTabSeed
}

type pullRequestTabState struct {
	label string
	rows  []PullRequestRow
}

type Model struct {
	focus                      Focus
	lastSideFocus              Focus
	users                      []Item
	pullRequestTabs            []pullRequestTabState
	selectedUserIndex          int
	activePullRequestTab       PullRequestTab
	selectedPullRequestIndexes map[PullRequestTab]int
	paneLayoutSize             PaneLayoutSize
	fullscreenPane             Focus
	detailFullscreenReturnSize PaneLayoutSize
	searchActive               bool
	searchTarget               Focus
	searchTargetPullRequestTab PullRequestTab
	searchDraft                string
	userSearchQuery            string
	detailSearchQuery          string
	pullRequestSearchQueries   map[PullRequestTab]string
	actionsPopup               actionsPopupState
}

func NewModel(seed SeedData) *Model {
	model := &Model{
		focus:                      FocusUserView,
		lastSideFocus:              FocusUserView,
		users:                      copyItems(seed.Users),
		selectedPullRequestIndexes: map[PullRequestTab]int{},
		pullRequestSearchQueries:   map[PullRequestTab]string{},
	}
	model.SetPullRequestTabs(seed.pullRequestTabs())
	return model
}

func DefaultSeedData() SeedData {
	return SeedData{
		Users: []Item{connectedUserLoadingItem()},
		PullRequestTabs: []PullRequestTabSeed{
			{Label: "My PRs", PullRequests: []Item{myPullRequestsLoadingItem()}},
			{Label: "Requested", PullRequests: []Item{requestedPullRequestsLoadingItem()}},
		},
	}
}

func (seed SeedData) pullRequestTabs() []PullRequestTabSeed {
	if len(seed.PullRequestTabs) > 0 {
		return copyPullRequestTabSeeds(seed.PullRequestTabs)
	}

	return []PullRequestTabSeed{
		{Label: "My PRs", PullRequests: copyItems(seed.MyPullRequests)},
		{Label: "Requested", PullRequests: copyItems(seed.RequestedPullRequests)},
	}
}

func copyPullRequestTabSeeds(seeds []PullRequestTabSeed) []PullRequestTabSeed {
	copied := make([]PullRequestTabSeed, 0, len(seeds))
	for _, seed := range seeds {
		copied = append(copied, PullRequestTabSeed{Label: seed.Label, PullRequests: copyItems(seed.PullRequests)})
	}
	return copied
}

func (model *Model) SetPullRequestTabs(seeds []PullRequestTabSeed) {
	normalizedSeeds := normalizePullRequestTabSeeds(seeds)
	if len(normalizedSeeds) == 0 {
		normalizedSeeds = DefaultSeedData().pullRequestTabs()
	}

	previousSelectedIndexes := model.selectedPullRequestIndexes
	previousSearchQueries := model.pullRequestSearchQueries
	previousActiveTab := model.activePullRequestTab
	previousSearchTargetTab := model.searchTargetPullRequestTab

	model.pullRequestTabs = make([]pullRequestTabState, 0, len(normalizedSeeds))
	model.selectedPullRequestIndexes = make(map[PullRequestTab]int, len(normalizedSeeds))
	model.pullRequestSearchQueries = make(map[PullRequestTab]string, len(normalizedSeeds))
	for index, seed := range normalizedSeeds {
		tab := PullRequestTab(index)
		rows := pullRequestRowsFromItems(seed.PullRequests)
		model.pullRequestTabs = append(model.pullRequestTabs, pullRequestTabState{label: seed.Label, rows: rows})
		model.selectedPullRequestIndexes[tab] = clampIndex(previousSelectedIndexes[tab], len(rows))
		model.pullRequestSearchQueries[tab] = previousSearchQueries[tab]
	}

	model.activePullRequestTab = PullRequestTab(clampIndex(int(previousActiveTab), len(model.pullRequestTabs)))
	model.searchTargetPullRequestTab = PullRequestTab(clampIndex(int(previousSearchTargetTab), len(model.pullRequestTabs)))
	for _, tab := range model.PullRequestTabs() {
		model.clampSearchSelectionForPullRequestTab(tab)
	}
}

func normalizePullRequestTabSeeds(seeds []PullRequestTabSeed) []PullRequestTabSeed {
	if len(seeds) == 0 {
		return nil
	}

	normalized := make([]PullRequestTabSeed, 0, len(seeds))
	for _, seed := range seeds {
		if seed.Label == "" {
			continue
		}
		normalized = append(normalized, PullRequestTabSeed{Label: seed.Label, PullRequests: copyItems(seed.PullRequests)})
	}

	return normalized
}

func (model *Model) Focus() Focus {
	return model.focus
}

func (model *Model) ActivePullRequestTab() PullRequestTab {
	return model.activePullRequestTab
}

func (model *Model) PullRequestTabs() []PullRequestTab {
	tabs := make([]PullRequestTab, 0, len(model.pullRequestTabs))
	for index := range model.pullRequestTabs {
		tabs = append(tabs, PullRequestTab(index))
	}
	return tabs
}

func (model *Model) PullRequestTabLabel(tab PullRequestTab) string {
	index := int(tab)
	if index < 0 || index >= len(model.pullRequestTabs) {
		return tab.Label()
	}
	return model.pullRequestTabs[index].label
}

func (model *Model) SelectedUserIndex() int {
	return model.selectedUserIndex
}

func (model *Model) SelectedPullRequestIndex(tab PullRequestTab) int {
	return model.selectedPullRequestIndexes[tab]
}

func (model *Model) Users() []Item {
	return copyItems(model.users)
}

func (model *Model) SetUsers(users []Item) {
	model.users = copyItems(users)
	model.selectedUserIndex = clampIndex(model.selectedUserIndex, len(model.users))
	model.clampSearchSelectionForUserView()
}

func (model *Model) PullRequests(tab PullRequestTab) []Item {
	return pullRequestItems(model.pullRequestRows(tab))
}

func (model *Model) PullRequestRows(tab PullRequestTab) []PullRequestRow {
	return copyPullRequestRows(model.pullRequestRows(tab))
}

func (model *Model) SetPullRequests(tab PullRequestTab, pullRequests []Item) {
	model.SetPullRequestRows(tab, pullRequestRowsFromItems(pullRequests))
}

func (model *Model) SetPullRequestRows(tab PullRequestTab, pullRequests []PullRequestRow) {
	index := int(tab)
	if index < 0 || index >= len(model.pullRequestTabs) {
		return
	}

	model.pullRequestTabs[index].rows = copyPullRequestRows(pullRequests)
	model.selectedPullRequestIndexes[tab] = clampIndex(model.selectedPullRequestIndexes[tab], len(model.pullRequestTabs[index].rows))
	model.clampSearchSelectionForPullRequestTab(tab)
}

func (model *Model) CurrentPullRequests() []Item {
	return model.PullRequests(model.activePullRequestTab)
}

func (model *Model) SelectedPullRequestRow() (PullRequestRow, bool) {
	rows := model.pullRequestRows(model.activePullRequestTab)
	selectedIndex := model.selectedPullRequestIndexes[model.activePullRequestTab]
	return pullRequestRowAt(rows, selectedIndex)
}

func (model *Model) SelectedPullRequestSummary() (githubcli.PullRequest, bool) {
	row, ok := model.SelectedPullRequestRow()
	if !ok || row.Summary == nil {
		return githubcli.PullRequest{}, false
	}

	return *row.Summary, true
}

func (model *Model) DetailContent() string {
	item, ok := model.detailItem()
	if !ok {
		return ""
	}

	return item.Detail
}

func (model *Model) NextSideView() {
	if model.focus == FocusDetailView || model.paneLayoutSize == PaneLayoutFullscreen {
		return
	}

	switch model.currentSideFocus() {
	case FocusPullRequestsView:
		model.setSideFocus(FocusUserView)
	default:
		model.setSideFocus(FocusPullRequestsView)
	}
}

func (model *Model) PreviousSideView() {
	if model.focus == FocusDetailView {
		return
	}

	model.NextSideView()
}

func (model *Model) OpenDetail() {
	model.FocusDetailView()
}

func (model *Model) FocusDetailView() {
	if model.paneLayoutSize == PaneLayoutFullscreen && model.fullscreenPane != FocusDetailView {
		return
	}

	model.lastSideFocus = model.currentSideFocus()
	model.focus = FocusDetailView
}

func (model *Model) FocusUserView() {
	model.setSideFocus(FocusUserView)
}

func (model *Model) FocusPullRequestsView() {
	model.setSideFocus(FocusPullRequestsView)
}

func (model *Model) CloseDetail() {
	if model.focus != FocusDetailView {
		return
	}
	if model.paneLayoutSize == PaneLayoutFullscreen && model.fullscreenPane == FocusDetailView {
		model.paneLayoutSize = model.detailFullscreenReturnSize
	}

	model.focus = model.currentSideFocus()
}

func (model *Model) MoveSelectionDown() {
	switch model.focus {
	case FocusUserView:
		model.selectedUserIndex = adjustVisibleSelection(model.selectedUserIndex, model.visibleUserIndexes(), 1)
	case FocusPullRequestsView:
		model.adjustPullRequestSelection(1)
	}
}

func (model *Model) MoveSelectionUp() {
	switch model.focus {
	case FocusUserView:
		model.selectedUserIndex = adjustVisibleSelection(model.selectedUserIndex, model.visibleUserIndexes(), -1)
	case FocusPullRequestsView:
		model.adjustPullRequestSelection(-1)
	}
}

func (model *Model) MoveSelectionToTop() {
	switch model.focus {
	case FocusUserView:
		model.selectedUserIndex = firstVisibleIndex(model.selectedUserIndex, model.visibleUserIndexes())
	case FocusPullRequestsView:
		model.selectedPullRequestIndexes[model.activePullRequestTab] = firstVisibleIndex(model.selectedPullRequestIndexes[model.activePullRequestTab], model.visiblePullRequestIndexes(model.activePullRequestTab))
	}
}

func (model *Model) MoveSelectionToBottom() {
	switch model.focus {
	case FocusUserView:
		model.selectedUserIndex = lastVisibleIndex(model.selectedUserIndex, model.visibleUserIndexes())
	case FocusPullRequestsView:
		model.selectedPullRequestIndexes[model.activePullRequestTab] = lastVisibleIndex(model.selectedPullRequestIndexes[model.activePullRequestTab], model.visiblePullRequestIndexes(model.activePullRequestTab))
	}
}

func (model *Model) PageDown(pageSize int) {
	model.adjustSelectionBy(pageDelta(pageSize))
}

func (model *Model) PageUp(pageSize int) {
	model.adjustSelectionBy(-pageDelta(pageSize))
}

func (model *Model) NextPullRequestTab() {
	if model.focus != FocusPullRequestsView || len(model.pullRequestTabs) <= 1 {
		return
	}

	model.activePullRequestTab = PullRequestTab((int(model.activePullRequestTab) + 1) % len(model.pullRequestTabs))
}

func (model *Model) PreviousPullRequestTab() {
	if model.focus != FocusPullRequestsView || len(model.pullRequestTabs) <= 1 {
		return
	}

	activeIndex := int(model.activePullRequestTab) - 1
	if activeIndex < 0 {
		activeIndex = len(model.pullRequestTabs) - 1
	}
	model.activePullRequestTab = PullRequestTab(activeIndex)
}
