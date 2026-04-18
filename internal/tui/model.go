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
	}
}

func DefaultSeedData() SeedData {
	return SeedData{
		Users: []Item{{
			Title:  "connected-user",
			Detail: "Dummy connected user. Real `gh api user` data arrives in TODO 03.",
		}},
		MyPullRequests: []Item{
			{Title: "lazygh#12 Add list rendering", Detail: "Dummy PR body for list rendering. Real PR descriptions arrive in TODO 04."},
			{Title: "lazygh#18 Tighten keyboard flow", Detail: "Dummy PR body for keyboard flow. Press `enter` to focus the detail pane."},
			{Title: "lazygh#24 Improve borders", Detail: "Dummy PR body for border styling. Yes, the border colors are somehow a feature now."},
		},
		RequestedPullRequests: []Item{
			{Title: "core/api#91 Review auth cleanup", Detail: "Dummy requested review body. Real review requests arrive in TODO 05."},
			{Title: "infra/cli#44 Review release notes", Detail: "Dummy requested review body. Use `[` and `]` in the PR view to switch tabs."},
		},
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

func (model *Model) PullRequests(tab PullRequestTab) []Item {
	switch tab {
	case RequestedPullRequestsTab:
		return copyItems(model.requestedPullRequests)
	default:
		return copyItems(model.myPullRequests)
	}
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
	switch model.currentSideFocus() {
	case FocusPullRequestsView:
		model.setSideFocus(FocusUserView)
	default:
		model.setSideFocus(FocusPullRequestsView)
	}
}

func (model *Model) PreviousSideView() {
	model.NextSideView()
}

func (model *Model) OpenDetail() {
	if model.focus == FocusDetailView {
		return
	}

	model.lastSideFocus = model.currentSideFocus()
	model.focus = FocusDetailView
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
		model.selectedUserIndex = clampIndex(model.selectedUserIndex+1, len(model.users))
	case FocusPullRequestsView:
		model.adjustPullRequestSelection(1)
	}
}

func (model *Model) MoveSelectionUp() {
	switch model.focus {
	case FocusUserView:
		model.selectedUserIndex = clampIndex(model.selectedUserIndex-1, len(model.users))
	case FocusPullRequestsView:
		model.adjustPullRequestSelection(-1)
	}
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

func (model *Model) adjustPullRequestSelection(change int) {
	items := model.CurrentPullRequests()
	selectedIndex := model.selectedPullRequestIndexes[model.activePullRequestTab]
	model.selectedPullRequestIndexes[model.activePullRequestTab] = clampIndex(selectedIndex+change, len(items))
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

func copyItems(items []Item) []Item {
	copiedItems := make([]Item, len(items))
	copy(copiedItems, items)
	return copiedItems
}
