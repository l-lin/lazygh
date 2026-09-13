package githubcli

import (
	"errors"
	"strings"
)

const pullRequestDetailGraphQLQuery = `query($owner:String!,$name:String!,$number:Int!,$threadCursor:String,$commentCursor:String,$ids:[ID!]!){viewer{login}repository(owner:$owner,name:$name){pullRequest(number:$number){isMergeQueueEnabled isInMergeQueue viewerCanEnableAutoMerge mergeQueueEntry{id state position estimatedTimeToMerge} reviews(last:100){nodes{id state author{login}}} id reactionGroups{content viewerHasReacted users{totalCount}} comments(first:100,after:$commentCursor){pageInfo{hasNextPage endCursor}nodes{id viewerDidAuthor author{login} body createdAt url reactionGroups{content viewerHasReacted users{totalCount}}}} reviewThreads(first:100,after:$threadCursor){pageInfo{hasNextPage endCursor}nodes{id isResolved isOutdated viewerCanResolve viewerCanUnresolve path line originalLine startLine originalStartLine diffSide startDiffSide comments(first:100){pageInfo{hasNextPage endCursor}nodes{id viewerDidAuthor state author{login} body createdAt url diffHunk reactionGroups{content viewerHasReacted users{totalCount}}}}}}}}nodes(ids:$ids){... on PullRequestReviewComment{id reactionGroups{content viewerHasReacted users{totalCount}}}... on PullRequestReview{id reactionGroups{content viewerHasReacted users{totalCount}}}}}`
const pullRequestDetailGraphQLQueryWithoutReactionGroups = `query($owner:String!,$name:String!,$number:Int!,$threadCursor:String,$commentCursor:String){viewer{login}repository(owner:$owner,name:$name){pullRequest(number:$number){isMergeQueueEnabled isInMergeQueue viewerCanEnableAutoMerge mergeQueueEntry{id state position estimatedTimeToMerge} reviews(last:100){nodes{id state author{login}}} id reactionGroups{content viewerHasReacted users{totalCount}} comments(first:100,after:$commentCursor){pageInfo{hasNextPage endCursor}nodes{id viewerDidAuthor author{login} body createdAt url reactionGroups{content viewerHasReacted users{totalCount}}}} reviewThreads(first:100,after:$threadCursor){pageInfo{hasNextPage endCursor}nodes{id isResolved isOutdated viewerCanResolve viewerCanUnresolve path line originalLine startLine originalStartLine diffSide startDiffSide comments(first:100){pageInfo{hasNextPage endCursor}nodes{id viewerDidAuthor state author{login} body createdAt url diffHunk reactionGroups{content viewerHasReacted users{totalCount}}}}}}}}}`

type pullRequestDetailGraphQLData struct {
	MergeQueueMetadata      pullRequestMergeQueueMetadata
	ReactionTargets         pullRequestReactionTargets
	ReviewThreads           []PullRequestReviewThread
	ReactionGroupsByID      map[string][]ReactionGroup
	PendingReviewID         string
	PendingReviewStateKnown bool
}

type pullRequestDetailGraphQLPage struct {
	Viewer struct {
		Login string `json:"login"`
	} `json:"viewer"`
	Repository *struct {
		PullRequest *pullRequestDetailGraphQLPullRequest `json:"pullRequest"`
	} `json:"repository"`
	Nodes []*struct {
		ID             string          `json:"id"`
		ReactionGroups []ReactionGroup `json:"reactionGroups"`
	} `json:"nodes"`
}

type pullRequestDetailGraphQLPullRequest struct {
	IsMergeQueueEnabled      bool                        `json:"isMergeQueueEnabled"`
	IsInMergeQueue           bool                        `json:"isInMergeQueue"`
	ViewerCanEnableAutoMerge bool                        `json:"viewerCanEnableAutoMerge"`
	MergeQueueEntry          *PullRequestMergeQueueEntry `json:"mergeQueueEntry"`
	Reviews                  struct {
		Nodes []struct {
			ID     string `json:"id"`
			State  string `json:"state"`
			Author *struct {
				Login string `json:"login"`
			} `json:"author"`
		} `json:"nodes"`
	} `json:"reviews"`
	ID             string          `json:"id"`
	ReactionGroups []ReactionGroup `json:"reactionGroups"`
	Comments       struct {
		PageInfo struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
		Nodes []PullRequestComment `json:"nodes"`
	} `json:"comments"`
	ReviewThreads struct {
		PageInfo struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
		Nodes []struct {
			ID                 string `json:"id"`
			IsResolved         bool   `json:"isResolved"`
			IsOutdated         bool   `json:"isOutdated"`
			ViewerCanResolve   bool   `json:"viewerCanResolve"`
			ViewerCanUnresolve bool   `json:"viewerCanUnresolve"`
			Path               string `json:"path"`
			Line               int    `json:"line"`
			OriginalLine       int    `json:"originalLine"`
			StartLine          int    `json:"startLine"`
			OriginalStartLine  int    `json:"originalStartLine"`
			DiffSide           string `json:"diffSide"`
			StartDiffSide      string `json:"startDiffSide"`
			Comments           struct {
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []PullRequestComment `json:"nodes"`
			} `json:"comments"`
		} `json:"nodes"`
	} `json:"reviewThreads"`
}

func (client *PullRequestDetailService) loadPullRequestDetailGraphQLData(repository string, number int, detail PullRequestDetail, inlineComments []PullRequestInlineComment) (pullRequestDetailGraphQLData, error) {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return pullRequestDetailGraphQLData{}, err
	}

	owner, name, err := splitRepositoryOwnerAndName(trimmedRepository)
	if err != nil {
		return pullRequestDetailGraphQLData{}, err
	}

	reactionTargetIDs := uniqueReactionTargetIDs(append(
		pullRequestInlineCommentReactionTargetIDs(inlineComments),
		pullRequestReviewReactionTargetIDs(detail.Reviews)...,
	))
	query := pullRequestDetailGraphQLQueryWithoutReactionGroups
	if len(reactionTargetIDs) > 0 {
		query = pullRequestDetailGraphQLQuery
	}

	data := pullRequestDetailGraphQLData{ReactionGroupsByID: map[string][]ReactionGroup{}}
	threadCursor := ""
	commentCursor := ""
	for {
		request := GraphQLRequest{Query: query, Variables: []GraphQLVariable{
			typedGraphQLVariable("owner", owner),
			typedGraphQLVariable("name", name),
			typedGraphQLVariable("number", number),
		}}
		if threadCursor != "" {
			request.Variables = append(request.Variables, typedGraphQLVariable("threadCursor", threadCursor))
		}
		if commentCursor != "" {
			request.Variables = append(request.Variables, typedGraphQLVariable("commentCursor", commentCursor))
		}
		for _, id := range reactionTargetIDs {
			request.Variables = append(request.Variables, typedGraphQLVariable("ids[]", id))
		}

		result, err := client.queryGraphQL(request)
		if err != nil {
			return pullRequestDetailGraphQLData{}, err
		}
		page, err := parsePullRequestDetailGraphQLPage(result.Stdout)
		if err != nil {
			return pullRequestDetailGraphQLData{}, err
		}
		if page.Repository == nil || page.Repository.PullRequest == nil {
			return pullRequestDetailGraphQLData{}, ErrInvalidPullRequestDetailGraphQLResponse
		}

		pullRequest := page.Repository.PullRequest
		firstPage := threadCursor == "" && commentCursor == ""
		if firstPage && strings.TrimSpace(page.Viewer.Login) != "" {
			data.PendingReviewStateKnown = true
			for index := len(pullRequest.Reviews.Nodes) - 1; index >= 0; index-- {
				review := pullRequest.Reviews.Nodes[index]
				if !strings.EqualFold(strings.TrimSpace(review.State), "PENDING") || review.Author == nil || !strings.EqualFold(strings.TrimSpace(review.Author.Login), strings.TrimSpace(page.Viewer.Login)) {
					continue
				}
				data.PendingReviewID = strings.TrimSpace(review.ID)
				break
			}
		}
		if data.ReactionTargets.PullRequestID == "" {
			data.MergeQueueMetadata = pullRequestMergeQueueMetadata{
				IsMergeQueueEnabled:      pullRequest.IsMergeQueueEnabled,
				IsInMergeQueue:           pullRequest.IsInMergeQueue,
				ViewerCanEnableAutoMerge: pullRequest.ViewerCanEnableAutoMerge,
				MergeQueueEntry:          normalizePullRequestMergeQueueEntry(pullRequest.MergeQueueEntry),
			}
			data.ReactionTargets.PullRequestID = strings.TrimSpace(pullRequest.ID)
			data.ReactionTargets.ReactionGroups = normalizeReactionGroups(pullRequest.ReactionGroups)
		}
		if firstPage || commentCursor != "" {
			data.ReactionTargets.Comments = append(data.ReactionTargets.Comments, normalizePullRequestComments(pullRequest.Comments.Nodes)...)
		}
		if firstPage || threadCursor != "" {
			data.ReviewThreads = append(data.ReviewThreads, pullRequestDetailGraphQLReviewThreads(pullRequest.ReviewThreads.Nodes)...)
		}
		for _, node := range page.Nodes {
			if node == nil || strings.TrimSpace(node.ID) == "" {
				continue
			}
			data.ReactionGroupsByID[strings.TrimSpace(node.ID)] = normalizeReactionGroups(node.ReactionGroups)
		}

		if (firstPage || threadCursor != "") && len(pullRequest.ReviewThreads.Nodes) > 0 {
			for index := range pullRequest.ReviewThreads.Nodes {
				node := pullRequest.ReviewThreads.Nodes[index]
				if !node.Comments.PageInfo.HasNextPage {
					continue
				}
				threadID := strings.TrimSpace(node.ID)
				nextCursor := strings.TrimSpace(node.Comments.PageInfo.EndCursor)
				if threadID == "" || nextCursor == "" {
					return pullRequestDetailGraphQLData{}, ErrInvalidPullRequestReviewThreadsResponse
				}
				comments, err := newReviewThreadAssembler(client).loadCommentsAfter(threadID, nextCursor)
				if err != nil {
					return pullRequestDetailGraphQLData{}, err
				}
				data.ReviewThreads[len(data.ReviewThreads)-len(pullRequest.ReviewThreads.Nodes)+index].Comments = append(data.ReviewThreads[len(data.ReviewThreads)-len(pullRequest.ReviewThreads.Nodes)+index].Comments, comments...)
			}
		}

		threadHasNextPage := pullRequest.ReviewThreads.PageInfo.HasNextPage
		commentHasNextPage := pullRequest.Comments.PageInfo.HasNextPage
		if !threadHasNextPage && !commentHasNextPage {
			return data, nil
		}
		if threadHasNextPage {
			threadCursor = strings.TrimSpace(pullRequest.ReviewThreads.PageInfo.EndCursor)
			if threadCursor == "" {
				return pullRequestDetailGraphQLData{}, ErrInvalidPullRequestReviewThreadsResponse
			}
		} else {
			threadCursor = ""
		}
		if commentHasNextPage {
			commentCursor = strings.TrimSpace(pullRequest.Comments.PageInfo.EndCursor)
			if commentCursor == "" {
				return pullRequestDetailGraphQLData{}, ErrInvalidPullRequestReactionTargetsResponse
			}
		} else {
			commentCursor = ""
		}
	}
}

func parsePullRequestDetailGraphQLPage(stdout []byte) (pullRequestDetailGraphQLPage, error) {
	var response pullRequestDetailGraphQLPage
	if err := decodeEndpointGraphQLResponse(stdout, &response, ErrInvalidPullRequestDetailGraphQLResponse); err != nil {
		return pullRequestDetailGraphQLPage{}, err
	}
	return response, nil
}

func pullRequestDetailGraphQLReviewThreads(nodes []struct {
	ID                 string `json:"id"`
	IsResolved         bool   `json:"isResolved"`
	IsOutdated         bool   `json:"isOutdated"`
	ViewerCanResolve   bool   `json:"viewerCanResolve"`
	ViewerCanUnresolve bool   `json:"viewerCanUnresolve"`
	Path               string `json:"path"`
	Line               int    `json:"line"`
	OriginalLine       int    `json:"originalLine"`
	StartLine          int    `json:"startLine"`
	OriginalStartLine  int    `json:"originalStartLine"`
	DiffSide           string `json:"diffSide"`
	StartDiffSide      string `json:"startDiffSide"`
	Comments           struct {
		PageInfo struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
		Nodes []PullRequestComment `json:"nodes"`
	} `json:"comments"`
}) []PullRequestReviewThread {
	threads := make([]PullRequestReviewThread, 0, len(nodes))
	for _, node := range nodes {
		threads = append(threads, PullRequestReviewThread{
			ID:                 node.ID,
			IsResolved:         node.IsResolved,
			IsOutdated:         node.IsOutdated,
			ViewerCanResolve:   node.ViewerCanResolve,
			ViewerCanUnresolve: node.ViewerCanUnresolve,
			Path:               node.Path,
			Line:               node.Line,
			OriginalLine:       node.OriginalLine,
			StartLine:          node.StartLine,
			OriginalStartLine:  node.OriginalStartLine,
			DiffSide:           node.DiffSide,
			StartDiffSide:      node.StartDiffSide,
			Comments:           normalizePullRequestComments(node.Comments.Nodes),
		}.normalized())
	}
	return threads
}

var ErrInvalidPullRequestDetailGraphQLResponse = errors.New("invalid pull request detail GraphQL response")
