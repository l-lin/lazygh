package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type searchRepeatDirection int

const (
	searchRepeatForward searchRepeatDirection = iota
	searchRepeatBackward
)

type clipboardWriteSelectionTarget int

const (
	clipboardWriteSelectionNone clipboardWriteSelectionTarget = iota
	clipboardWriteSelectionDetail
	clipboardWriteSelectionBuildPopup
)

type MsgFeedbackSet struct {
	Target  Focus
	Message string
}

type MsgActionsPopupActionErrorHandled struct {
	Err error
}

type MsgModalEditorOpened struct {
	State *modalEditorState
}

type MsgModalEditorClosed struct{}

type MsgModalEditorSubmitRequested struct{}

type MsgModalEditorSubmitFinished struct {
	Err     error
	Success modalEditorSubmitSuccess
}

type MsgModalEditorExternalEditRequested struct{}

type MsgModalEditorExternalEditFinished struct {
	Text string
	Err  error
}

type MsgPullRequestBuildRunLoadRequested struct {
	Target pullRequestBuildRunTarget
}

type MsgPullRequestBuildRunJobLogLoadRequested struct {
	Summary githubdomain.PullRequest
	Check   githubdomain.PullRequestStatusCheck
}

type MsgPullRequestBuildRunPopupOpened struct {
	Content pullRequestBuildRunPopupContent
}

type MsgPullRequestBuildRunPopupClosed struct{}

type MsgAdvanceDetailTab struct {
	Delta int
}

type MsgAdvancePullRequestTab struct {
	Delta int
}

type MsgOpenDetailRequested struct{}

type MsgCloseDetailRequested struct{}

type MsgStartReviewFileTreeSearch struct {
	Query string
}

type MsgSubmitReviewFileTreeSearch struct{}

type MsgCancelReviewFileTreeSearch struct{}

type MsgOpenPullRequestInBrowserView struct {
	Summary githubdomain.PullRequest
}

type MsgOpenPullRequestInDetailFullscreen struct {
	SideFocus Focus
}

type MsgExitReviewMode struct{}

type MsgToggleHelp struct{}

type MsgCloseHelp struct{}

type MsgAdjustFocusedPane struct {
	Delta int
}

type MsgLineNavigationRequested struct {
	Delta int
}

type pageNavigationKind int

const (
	pageNavigationKindHalfDown pageNavigationKind = iota
	pageNavigationKindHalfUp
	pageNavigationKindFullDown
	pageNavigationKindFullUp
)

type MsgPageNavigationRequested struct {
	Kind pageNavigationKind
}

type MsgSideListViewportRequested struct {
	Placement viewportPlacement
}

type detailViewportOperation int

const (
	detailViewportOperationScrollDown detailViewportOperation = iota
	detailViewportOperationScrollUp
	detailViewportOperationRecenter
	detailViewportOperationPlaceTop
	detailViewportOperationPlaceBottom
)

type MsgDetailViewportRequested struct {
	Operation detailViewportOperation
}

type MsgOpenBrowserURLRequested struct {
	URL            string
	SuccessMessage string
	FailureMessage string
	Target         Focus
}

type MsgOpenBrowserURLFinished struct {
	SuccessMessage string
	FailureMessage string
	Target         Focus
	Err            error
}

type MsgClipboardWriteFinished struct {
	SuccessMessage  string
	FailureMessage  string
	Target          Focus
	Err             error
	Selection       detailSelectionRange
	SelectionTarget clipboardWriteSelectionTarget
}

type MsgReadPullRequestURLFromClipboardRequested struct{}

type MsgOpenPullRequestByURLSubmitRequested struct {
	URL string
}

type MsgPullRequestURLReadFromClipboard struct {
	URL string
	Err error
}

type MsgOpenLinkUnderCursorRequested struct{}

type MsgOpenPullRequestBuildRunPopupLinkRequested struct{}

type MsgCopySelectedDetailTextRequested struct{}

type MsgCopyPullRequestURLRequested struct{}

type MsgCopyPullRequestBuildRunPopupContentRequested struct{}

type MsgOpenNotificationInBrowserRequested struct{}

type MsgRefreshNotificationsRequested struct{}

type MsgRepeatActionsPopupSearch struct {
	Direction searchRepeatDirection
}

type MsgRepeatSideSearch struct {
	Focus     Focus
	Direction searchRepeatDirection
}

type MsgRepeatPullRequestSearch struct {
	Direction searchRepeatDirection
}

type MsgRepeatReviewFileTreeSearch struct {
	Direction searchRepeatDirection
}

type MsgToggleReviewTreeRowVisibility struct{}

type MsgSetAllReviewTreeFolds struct {
	Collapsed bool
}

type MsgMoveReviewSelection struct {
	Delta int
}

type MsgMoveReviewSelectionToTop struct{}

type MsgMoveReviewSelectionToBottom struct{}

type MsgMoveReviewFile struct {
	Delta int
}

type MsgMoveReviewComment struct {
	Direction reviewNavigationDirection
}

type MsgSearchWordUnderCursor struct {
	Reverse bool
}

type MsgRepeatDetailSearchRequested struct {
	Direction searchRepeatDirection
}

type MsgDetailSearchWordResolved struct {
	Query   string
	Reverse bool
}

type MsgToggleInlineConversationVisibility struct{}

type MsgSetAllDetailFolds struct {
	Collapsed bool
}

func (MsgFeedbackSet) isMsg()                                  {}
func (MsgActionsPopupActionErrorHandled) isMsg()               {}
func (MsgModalEditorOpened) isMsg()                            {}
func (MsgModalEditorClosed) isMsg()                            {}
func (MsgModalEditorSubmitRequested) isMsg()                   {}
func (MsgModalEditorSubmitFinished) isMsg()                    {}
func (MsgModalEditorExternalEditRequested) isMsg()             {}
func (MsgModalEditorExternalEditFinished) isMsg()              {}
func (MsgPullRequestBuildRunLoadRequested) isMsg()             {}
func (MsgPullRequestBuildRunJobLogLoadRequested) isMsg()       {}
func (MsgPullRequestBuildRunPopupOpened) isMsg()               {}
func (MsgPullRequestBuildRunPopupClosed) isMsg()               {}
func (MsgAdvanceDetailTab) isMsg()                             {}
func (MsgAdvancePullRequestTab) isMsg()                        {}
func (MsgOpenDetailRequested) isMsg()                          {}
func (MsgCloseDetailRequested) isMsg()                         {}
func (MsgStartReviewFileTreeSearch) isMsg()                    {}
func (MsgSubmitReviewFileTreeSearch) isMsg()                   {}
func (MsgCancelReviewFileTreeSearch) isMsg()                   {}
func (MsgOpenPullRequestInBrowserView) isMsg()                 {}
func (MsgOpenPullRequestInDetailFullscreen) isMsg()            {}
func (MsgExitReviewMode) isMsg()                               {}
func (MsgToggleHelp) isMsg()                                   {}
func (MsgCloseHelp) isMsg()                                    {}
func (MsgAdjustFocusedPane) isMsg()                            {}
func (MsgLineNavigationRequested) isMsg()                      {}
func (MsgPageNavigationRequested) isMsg()                      {}
func (MsgSideListViewportRequested) isMsg()                    {}
func (MsgDetailViewportRequested) isMsg()                      {}
func (MsgOpenBrowserURLRequested) isMsg()                      {}
func (MsgOpenBrowserURLFinished) isMsg()                       {}
func (MsgClipboardWriteFinished) isMsg()                       {}
func (MsgReadPullRequestURLFromClipboardRequested) isMsg()     {}
func (MsgOpenPullRequestByURLSubmitRequested) isMsg()          {}
func (MsgPullRequestURLReadFromClipboard) isMsg()              {}
func (MsgOpenLinkUnderCursorRequested) isMsg()                 {}
func (MsgOpenPullRequestBuildRunPopupLinkRequested) isMsg()    {}
func (MsgCopySelectedDetailTextRequested) isMsg()              {}
func (MsgCopyPullRequestURLRequested) isMsg()                  {}
func (MsgCopyPullRequestBuildRunPopupContentRequested) isMsg() {}
func (MsgOpenNotificationInBrowserRequested) isMsg()           {}
func (MsgRefreshNotificationsRequested) isMsg()                {}
func (MsgRepeatActionsPopupSearch) isMsg()                     {}
func (MsgRepeatSideSearch) isMsg()                             {}
func (MsgRepeatPullRequestSearch) isMsg()                      {}
func (MsgRepeatReviewFileTreeSearch) isMsg()                   {}
func (MsgToggleReviewTreeRowVisibility) isMsg()                {}
func (MsgSetAllReviewTreeFolds) isMsg()                        {}
func (MsgMoveReviewSelection) isMsg()                          {}
func (MsgMoveReviewSelectionToTop) isMsg()                     {}
func (MsgMoveReviewSelectionToBottom) isMsg()                  {}
func (MsgMoveReviewFile) isMsg()                               {}
func (MsgMoveReviewComment) isMsg()                            {}
func (MsgSearchWordUnderCursor) isMsg()                        {}
func (MsgRepeatDetailSearchRequested) isMsg()                  {}
func (MsgDetailSearchWordResolved) isMsg()                     {}
func (MsgToggleInlineConversationVisibility) isMsg()           {}
func (MsgSetAllDetailFolds) isMsg()                            {}
