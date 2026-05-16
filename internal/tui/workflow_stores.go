package tui

import (
	"net/http"
	"sync"
)

type persistentCacheStore struct {
	pullRequestCache persistentPullRequestCache
}

type sessionStore struct {
	connectedUserLoadStarted bool
	connectedUserLogin       string
	connectedUserName        string
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
}

func newPullRequestListStore(persistence *persistentCacheStore) *pullRequestListStore {
	return &pullRequestListStore{
		persistence:                       persistence,
		additionalPullRequestsLoadStarted: map[PullRequestTab]bool{},
		additionalPullRequestsLoading:     map[PullRequestTab]bool{},
		additionalPullRequestsCounts:      map[PullRequestTab]pullRequestCountState{},
	}
}

type notificationStore struct {
	persistence                       *persistentCacheStore
	notificationsLoadStarted          bool
	notificationsLoading              bool
	notificationsLoadingDetailMessage string
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
	pullRequestDetailDocumentCache       map[pullRequestDetailDocumentCacheKey]detailDocument
	pullRequestConversationDocumentCache map[pullRequestDetailDocumentCacheKey]browserConversationDocument
	pullRequestChangesRenderedRowsCache  map[pullRequestDetailDocumentCacheKey][]reviewDiffRenderedRow
	issueDetailCache                     map[string]issueDetailResult
	issueDetailLoadInFlight              map[string]bool
	releaseDetailCache                   map[string]releaseDetailResult
	releaseDetailLoadInFlight            map[string]bool
	browserCollapsedSectionStates        map[string]bool
}

func newDetailStore(persistence *persistentCacheStore) *detailStore {
	return &detailStore{
		persistence:                          persistence,
		pullRequestDetailCache:               map[string]pullRequestDetailResult{},
		pullRequestDetailLoadInFlight:        map[string]bool{},
		pullRequestDetailDocumentCache:       map[pullRequestDetailDocumentCacheKey]detailDocument{},
		pullRequestConversationDocumentCache: map[pullRequestDetailDocumentCacheKey]browserConversationDocument{},
		pullRequestChangesRenderedRowsCache:  map[pullRequestDetailDocumentCacheKey][]reviewDiffRenderedRow{},
		issueDetailCache:                     map[string]issueDetailResult{},
		issueDetailLoadInFlight:              map[string]bool{},
		releaseDetailCache:                   map[string]releaseDetailResult{},
		releaseDetailLoadInFlight:            map[string]bool{},
		browserCollapsedSectionStates:        map[string]bool{},
	}
}

type reviewStore struct {
	persistence                   *persistentCacheStore
	pullRequestDiffCache          map[string]pullRequestDiffResult
	pullRequestDiffLoadInFlight   map[string]bool
	reviewDiffRenderCache         map[reviewDiffRenderCacheKey]reviewDiffRenderCacheEntry
	pendingPullRequestReviewCache map[string]pendingPullRequestReviewState
}

func newReviewStore(persistence *persistentCacheStore) *reviewStore {
	return &reviewStore{
		persistence:                   persistence,
		pullRequestDiffCache:          map[string]pullRequestDiffResult{},
		pullRequestDiffLoadInFlight:   map[string]bool{},
		reviewDiffRenderCache:         map[reviewDiffRenderCacheKey]reviewDiffRenderCacheEntry{},
		pendingPullRequestReviewCache: map[string]pendingPullRequestReviewState{},
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
	storyReviewLoading bool
	feedbackMessage    string
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
