package tui

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

type SeedData struct {
	Users                 []Item
	MyPullRequests        []Item
	RequestedPullRequests []Item
}

type Model struct {
	focus                      Focus
	lastSideFocus              Focus
	users                      []Item
	myPullRequests             []Item
	requestedPullRequests      []Item
	selectedUserIndex          int
	activePullRequestTab       PullRequestTab
	selectedPullRequestIndexes map[PullRequestTab]int
	searchActive               bool
	searchTarget               Focus
	searchTargetPullRequestTab PullRequestTab
	searchDraft                string
	userSearchQuery            string
	detailSearchQuery          string
	pullRequestSearchQueries   map[PullRequestTab]string
}

func NewModel(seed SeedData) *Model {
	return &Model{
		focus:                 FocusUserView,
		lastSideFocus:         FocusUserView,
		users:                 copyItems(seed.Users),
		myPullRequests:        copyItems(seed.MyPullRequests),
		requestedPullRequests: copyItems(seed.RequestedPullRequests),
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
	switch tab {
	case RequestedPullRequestsTab:
		return copyItems(model.requestedPullRequests)
	default:
		return copyItems(model.myPullRequests)
	}
}

func (model *Model) SetPullRequests(tab PullRequestTab, pullRequests []Item) {
	switch tab {
	case RequestedPullRequestsTab:
		model.requestedPullRequests = copyItems(pullRequests)
		model.selectedPullRequestIndexes[RequestedPullRequestsTab] = clampIndex(model.selectedPullRequestIndexes[RequestedPullRequestsTab], len(model.requestedPullRequests))
	default:
		model.myPullRequests = copyItems(pullRequests)
		model.selectedPullRequestIndexes[MyPullRequestsTab] = clampIndex(model.selectedPullRequestIndexes[MyPullRequestsTab], len(model.myPullRequests))
	}

	model.clampSearchSelectionForPullRequestTab(tab)
}

func (model *Model) CurrentPullRequests() []Item {
	return model.PullRequests(model.activePullRequestTab)
}

func (model *Model) DetailContent() string {
	item, ok := model.detailItem()
	if !ok {
		return ""
	}

	return item.Detail
}

func (model *Model) NextSideView() {
	if model.focus == FocusDetailView {
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
		items := model.CurrentPullRequests()
		selectedIndex := model.selectedPullRequestIndexes[model.activePullRequestTab]
		return itemAt(items, selectedIndex)
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

func indexOfInt(items []int, expected int) int {
	for index, item := range items {
		if item == expected {
			return index
		}
	}

	return -1
}
