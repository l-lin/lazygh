package tui

import (
	"net/http"
	"sync"
)

type persistentCacheStore struct {
	pullRequestCache persistentPullRequestCache
}

type sessionStore struct {
	connectedUserLoadStarted       bool
	connectedUserStatusOperationID uint64
	connectedUserLogin             string
	connectedUserName              string
}

func newSessionStore() *sessionStore {
	return &sessionStore{}
}

type pullRequestListStore struct {
	persistence                       *persistentCacheStore
	myPullRequestsLoadStarted         bool
	requestedPullRequestsLoadStarted  bool
	myPullRequestsLoading             bool
	requestedPullRequestsLoading      bool
	myPullRequestsCount               int
	myPullRequestsCountKnown          bool
	requestedPullRequestsCount        int
	requestedPullRequestsCountKnown   bool
	additionalPullRequestsLoadStarted map[PullRequestTab]bool
	additionalPullRequestsLoading     map[PullRequestTab]bool
	additionalPullRequestsCounts      map[PullRequestTab]pullRequestCountState
	pullRequestFreshness              map[string]pullRequestFreshnessState
	pullRequestFreshnessTracking      bool
	pullRequestTabMembership          map[PullRequestTab]map[string]struct{}
	pullRequestTabMembershipKnown     map[PullRequestTab]bool
	pullRequestLoadGenerations        map[PullRequestTab]uint64
	pullRequestStatusOperationIDs     map[PullRequestTab]uint64
	pullRequestRefreshErrorTab        PullRequestTab
	pullRequestRefreshErrorKnown      bool
}

func newPullRequestListStore(persistence *persistentCacheStore) *pullRequestListStore {
	return &pullRequestListStore{
		persistence:                       persistence,
		additionalPullRequestsLoadStarted: map[PullRequestTab]bool{},
		additionalPullRequestsLoading:     map[PullRequestTab]bool{},
		additionalPullRequestsCounts:      map[PullRequestTab]pullRequestCountState{},
		pullRequestFreshness:              map[string]pullRequestFreshnessState{},
		pullRequestTabMembership:          map[PullRequestTab]map[string]struct{}{},
		pullRequestTabMembershipKnown:     map[PullRequestTab]bool{},
		pullRequestLoadGenerations:        map[PullRequestTab]uint64{},
		pullRequestStatusOperationIDs:     map[PullRequestTab]uint64{},
	}
}

type notificationStore struct {
	persistence                       *persistentCacheStore
	notificationsLoadStarted          bool
	notificationsLoading              bool
	notificationsLoadingDetailMessage string
	notificationsStatusOperationID    uint64
	notificationDoneStore             notificationDoneStore
}

func newNotificationStore(persistence *persistentCacheStore) *notificationStore {
	return &notificationStore{
		persistence:           persistence,
		notificationDoneStore: noopNotificationDoneStore{},
	}
}

type detailStore struct {
	persistence                          *persistentCacheStore
	pullRequestDetailCache               map[string]pullRequestDetailResult
	pullRequestDetailLoadInFlight        map[string]bool
	pullRequestDetailStatusOperationIDs  map[string]uint64
	pullRequestDetailDocumentCache       map[pullRequestDetailDocumentCacheKey]detailDocument
	pullRequestConversationDocumentCache map[pullRequestDetailDocumentCacheKey]browserConversationDocument
	pullRequestChangesRenderedRowsCache  map[pullRequestDetailDocumentCacheKey][]reviewDiffRenderedRow
	issueDetailCache                     map[string]issueDetailResult
	issueDetailLoadInFlight              map[string]bool
	issueDetailStatusOperationIDs        map[string]uint64
	releaseDetailCache                   map[string]releaseDetailResult
	releaseDetailLoadInFlight            map[string]bool
	releaseDetailStatusOperationIDs      map[string]uint64
	browserCollapsedSectionStates        map[string]bool
}

func newDetailStore(persistence *persistentCacheStore) *detailStore {
	return &detailStore{
		persistence:                          persistence,
		pullRequestDetailCache:               map[string]pullRequestDetailResult{},
		pullRequestDetailLoadInFlight:        map[string]bool{},
		pullRequestDetailStatusOperationIDs:  map[string]uint64{},
		pullRequestDetailDocumentCache:       map[pullRequestDetailDocumentCacheKey]detailDocument{},
		pullRequestConversationDocumentCache: map[pullRequestDetailDocumentCacheKey]browserConversationDocument{},
		pullRequestChangesRenderedRowsCache:  map[pullRequestDetailDocumentCacheKey][]reviewDiffRenderedRow{},
		issueDetailCache:                     map[string]issueDetailResult{},
		issueDetailLoadInFlight:              map[string]bool{},
		issueDetailStatusOperationIDs:        map[string]uint64{},
		releaseDetailCache:                   map[string]releaseDetailResult{},
		releaseDetailLoadInFlight:            map[string]bool{},
		releaseDetailStatusOperationIDs:      map[string]uint64{},
		browserCollapsedSectionStates:        map[string]bool{},
	}
}

type reviewStore struct {
	persistence                       *persistentCacheStore
	pullRequestDiffCache              map[string]pullRequestDiffResult
	pullRequestDiffLoadInFlight       map[string]bool
	pullRequestDiffStatusOperationIDs map[string]uint64
	commitDiffCache                   map[string]commitDiffResult
	commitDiffLoadInFlight            map[string]bool
	storyReviewCache                  map[string]storyReviewResult
	reviewDiffRenderCache             map[reviewDiffRenderCacheKey]reviewDiffRenderCacheEntry
	pendingPullRequestReviewCache     map[string]pendingPullRequestReviewState
}

func newReviewStore(persistence *persistentCacheStore) *reviewStore {
	return &reviewStore{
		persistence:                       persistence,
		pullRequestDiffCache:              map[string]pullRequestDiffResult{},
		pullRequestDiffLoadInFlight:       map[string]bool{},
		pullRequestDiffStatusOperationIDs: map[string]uint64{},
		commitDiffCache:                   map[string]commitDiffResult{},
		commitDiffLoadInFlight:            map[string]bool{},
		storyReviewCache:                  map[string]storyReviewResult{},
		reviewDiffRenderCache:             map[reviewDiffRenderCacheKey]reviewDiffRenderCacheEntry{},
		pendingPullRequestReviewCache:     map[string]pendingPullRequestReviewState{},
	}
}

type buildStore struct {
	pullRequestBuildRunLoad  *pullRequestBuildRunLoadState
	pullRequestBuildRunPopup *pullRequestBuildRunPopupState
}

func newBuildStore() *buildStore {
	return &buildStore{}
}

type statusStore struct {
	storyReviewLoading         bool
	storyReviewLoadingMessage  string
	feedbackMessage            string
	ghCommandLoadingMessage    string
	nextStatusLineOperationID  uint64
	statusLineOperationStarted bool
	statusLineOperation        statusLineOperationState
}

func newStatusStore() *statusStore {
	return &statusStore{}
}

type optimisticMutationCoordinator struct {
	optimisticMutationSequence int
}

func newOptimisticMutationCoordinator() *optimisticMutationCoordinator {
	return &optimisticMutationCoordinator{}
}

type imageLoadCoordinator struct {
	detailImageStore            detailImageStore
	detailImageManager          detailImageManager
	detailImageHTMLLoadInFlight map[string]bool
	detailImageHTMLLoadFailed   map[string]bool
	detailImageLoadInFlight     map[string]bool
	detailImageLoadFailed       map[string]bool
	githubAuthToken             string
	githubAuthTokenLoaded       bool
	detailImageAuthTokenMu      sync.Mutex
	imageHTTPClient             *http.Client
}

func newImageLoadCoordinator(imageStore detailImageStore, imageManager detailImageManager) *imageLoadCoordinator {
	return &imageLoadCoordinator{
		detailImageStore:            imageStore,
		detailImageManager:          imageManager,
		detailImageHTMLLoadInFlight: map[string]bool{},
		detailImageHTMLLoadFailed:   map[string]bool{},
		detailImageLoadInFlight:     map[string]bool{},
		detailImageLoadFailed:       map[string]bool{},
		imageHTTPClient:             http.DefaultClient,
	}
}
