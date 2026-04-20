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

type Item struct {
	Title  string
	Detail string
}

type PullRequestRow struct {
	Item    Item
	Summary *githubcli.PullRequest
}

type SeedData struct {
	Users                 []Item
	MyPullRequests        []Item
	RequestedPullRequests []Item
}

type Model struct {
	focus                      Focus
	lastSideFocus              Focus
	users                      []Item
	myPullRequests             []PullRequestRow
	requestedPullRequests      []PullRequestRow
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
	return &Model{
		focus:                 FocusUserView,
		lastSideFocus:         FocusUserView,
		users:                 copyItems(seed.Users),
		myPullRequests:        pullRequestRowsFromItems(seed.MyPullRequests),
		requestedPullRequests: pullRequestRowsFromItems(seed.RequestedPullRequests),
		selectedPullRequestIndexes: map[PullRequestTab]int{
			MyPullRequestsTab:        0,
			RequestedPullRequestsTab: 0,
		},
		pullRequestSearchQueries: map[PullRequestTab]string{
			MyPullRequestsTab:        "",
			RequestedPullRequestsTab: "",
		},
	}
}

func DefaultSeedData() SeedData {
	return SeedData{
		Users:                 []Item{connectedUserLoadingItem()},
		MyPullRequests:        []Item{myPullRequestsLoadingItem()},
		RequestedPullRequests: []Item{requestedPullRequestsLoadingItem()},
	}
}

func (model *Model) Focus() Focus {
	return model.focus
}

func (model *Model) ActivePullRequestTab() PullRequestTab {
	return model.activePullRequestTab
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
	switch tab {
	case RequestedPullRequestsTab:
		model.requestedPullRequests = copyPullRequestRows(pullRequests)
		model.selectedPullRequestIndexes[RequestedPullRequestsTab] = clampIndex(model.selectedPullRequestIndexes[RequestedPullRequestsTab], len(model.requestedPullRequests))
	default:
		model.myPullRequests = copyPullRequestRows(pullRequests)
		model.selectedPullRequestIndexes[MyPullRequestsTab] = clampIndex(model.selectedPullRequestIndexes[MyPullRequestsTab], len(model.myPullRequests))
	}

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
		model.selectedUserIndex = model.adjustVisibleSelection(model.selectedUserIndex, model.visibleUserIndexes(), 1)
	case FocusPullRequestsView:
		model.adjustPullRequestSelection(1)
	}
}

func (model *Model) MoveSelectionUp() {
	switch model.focus {
	case FocusUserView:
		model.selectedUserIndex = model.adjustVisibleSelection(model.selectedUserIndex, model.visibleUserIndexes(), -1)
	case FocusPullRequestsView:
		model.adjustPullRequestSelection(-1)
	}
}

func (model *Model) PageDown(pageSize int) {
	model.adjustSelectionBy(pageDelta(pageSize))
}

func (model *Model) PageUp(pageSize int) {
	model.adjustSelectionBy(-pageDelta(pageSize))
}

func (model *Model) NextPullRequestTab() {
	if model.focus != FocusPullRequestsView {
		return
	}

	switch model.activePullRequestTab {
	case RequestedPullRequestsTab:
		model.activePullRequestTab = MyPullRequestsTab
	default:
		model.activePullRequestTab = RequestedPullRequestsTab
	}
}

func (model *Model) PreviousPullRequestTab() {
	model.NextPullRequestTab()
}

func (model *Model) detailItem() (Item, bool) {
	switch model.currentSideFocus() {
	case FocusPullRequestsView:
		row, ok := model.SelectedPullRequestRow()
		if !ok {
			return Item{}, false
		}
		return row.Item, true
	default:
		return itemAt(model.users, model.selectedUserIndex)
	}
}

func (model *Model) currentSideFocus() Focus {
	if model.focus == FocusDetailView {
		return model.lastSideFocus
	}

	if model.focus == FocusPullRequestsView {
		return FocusPullRequestsView
	}

	return FocusUserView
}

func (model *Model) setSideFocus(focus Focus) {
	if focus != FocusUserView && focus != FocusPullRequestsView {
		return
	}
	if model.paneLayoutSize == PaneLayoutFullscreen && focus != model.fullscreenPane {
		return
	}

	model.focus = focus
	model.lastSideFocus = focus
}

func (model *Model) adjustSelectionBy(change int) {
	switch model.focus {
	case FocusUserView:
		model.selectedUserIndex = model.adjustVisibleSelection(model.selectedUserIndex, model.visibleUserIndexes(), change)
	case FocusPullRequestsView:
		model.adjustPullRequestSelection(change)
	}
}

func (model *Model) adjustPullRequestSelection(change int) {
	tab := model.activePullRequestTab
	selectedIndex := model.selectedPullRequestIndexes[tab]
	visibleIndexes := model.visiblePullRequestIndexes(tab)
	model.selectedPullRequestIndexes[tab] = model.adjustVisibleSelection(selectedIndex, visibleIndexes, change)
}

func (model *Model) adjustVisibleSelection(selectedIndex int, visibleIndexes []int, change int) int {
	if len(visibleIndexes) == 0 {
		return selectedIndex
	}

	visibleSelectionIndex := indexOfInt(visibleIndexes, selectedIndex)
	if visibleSelectionIndex < 0 {
		visibleSelectionIndex = 0
	}

	visibleSelectionIndex = clampIndex(visibleSelectionIndex+change, len(visibleIndexes))
	return visibleIndexes[visibleSelectionIndex]
}

func (tab PullRequestTab) Label() string {
	switch tab {
	case RequestedPullRequestsTab:
		return "Requested"
	default:
		return "My PRs"
	}
}

func itemAt(items []Item, selectedIndex int) (Item, bool) {
	if len(items) == 0 {
		return Item{}, false
	}

	index := clampIndex(selectedIndex, len(items))
	return items[index], true
}

func pullRequestRowAt(rows []PullRequestRow, selectedIndex int) (PullRequestRow, bool) {
	if len(rows) == 0 {
		return PullRequestRow{}, false
	}

	index := clampIndex(selectedIndex, len(rows))
	return rows[index], true
}

func clampIndex(index int, itemCount int) int {
	if itemCount == 0 {
		return 0
	}

	if index < 0 {
		return 0
	}

	maxIndex := itemCount - 1
	if index > maxIndex {
		return maxIndex
	}

	return index
}

func pageDelta(pageSize int) int {
	if pageSize <= 1 {
		return 1
	}

	return pageSize
}

func copyItems(items []Item) []Item {
	copiedItems := make([]Item, len(items))
	copy(copiedItems, items)
	return copiedItems
}

func copyPullRequestRows(rows []PullRequestRow) []PullRequestRow {
	copiedRows := make([]PullRequestRow, 0, len(rows))
	for _, row := range rows {
		copiedRow := PullRequestRow{Item: row.Item}
		if row.Summary != nil {
			summaryCopy := *row.Summary
			copiedRow.Summary = &summaryCopy
		}
		copiedRows = append(copiedRows, copiedRow)
	}
	return copiedRows
}

func pullRequestRowsFromItems(items []Item) []PullRequestRow {
	rows := make([]PullRequestRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, PullRequestRow{Item: item})
	}
	return rows
}

func pullRequestItems(rows []PullRequestRow) []Item {
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.Item)
	}
	return items
}

func (model *Model) pullRequestRows(tab PullRequestTab) []PullRequestRow {
	switch tab {
	case RequestedPullRequestsTab:
		return model.requestedPullRequests
	default:
		return model.myPullRequests
	}
}

func indexOfInt(items []int, expected int) int {
	for index, item := range items {
		if item == expected {
			return index
		}
	}

	return -1
}
