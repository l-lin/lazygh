package githubcli

import (
	"fmt"
	"strings"
)

const pullRequestReviewThreadsQuery = `query($owner:String!,$name:String!,$number:Int!,$cursor:String){repository(owner:$owner,name:$name){pullRequest(number:$number){reviewThreads(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{id isResolved isOutdated viewerCanResolve viewerCanUnresolve path line originalLine startLine originalStartLine diffSide startDiffSide comments(first:100){pageInfo{hasNextPage endCursor}nodes{id viewerDidAuthor state author{login} body createdAt url diffHunk reactionGroups{content viewerHasReacted users{totalCount}}}}}}}}}`
const pullRequestReviewThreadCommentsQuery = `query($threadID:ID!,$cursor:String!){node(id:$threadID){... on PullRequestReviewThread{comments(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{id viewerDidAuthor state author{login} body createdAt url diffHunk reactionGroups{content viewerHasReacted users{totalCount}}}}}}}`

var ErrInvalidPullRequestReviewThreadsResponse = fmt.Errorf("invalid pull request review threads response")

type PullRequestReviewThread struct {
	ID                 string               `json:"id"`
	IsResolved         bool                 `json:"isResolved"`
	IsOutdated         bool                 `json:"isOutdated"`
	ViewerCanResolve   bool                 `json:"viewerCanResolve"`
	ViewerCanUnresolve bool                 `json:"viewerCanUnresolve"`
	Path               string               `json:"path"`
	Line               int                  `json:"line"`
	OriginalLine       int                  `json:"originalLine"`
	StartLine          int                  `json:"startLine"`
	OriginalStartLine  int                  `json:"originalStartLine"`
	DiffSide           string               `json:"diffSide"`
	StartDiffSide      string               `json:"startDiffSide"`
	Comments           []PullRequestComment `json:"-"`
}

type pullRequestReviewThreadsPage struct {
	Threads     []pullRequestReviewThreadPageNode
	HasNextPage bool
	EndCursor   string
}

type pullRequestReviewThreadPageNode struct {
	Thread              PullRequestReviewThread
	CommentsHasNextPage bool
	CommentsEndCursor   string
}

type pullRequestReviewThreadCommentsPage struct {
	Comments    []PullRequestComment
	HasNextPage bool
	EndCursor   string
}

func (client *PullRequestDetailService) listPullRequestReviewThreads(repository string, number int) ([]PullRequestReviewThread, error) {
	return newReviewThreadAssembler(client).List(repository, number)
}

func (client *PullRequestDetailService) pullRequestReviewThreadsPage(owner string, name string, number int, cursor string) (pullRequestReviewThreadsPage, error) {
	request := GraphQLRequest{Query: pullRequestReviewThreadsQuery, Variables: []GraphQLVariable{typedGraphQLVariable("owner", strings.TrimSpace(owner)), typedGraphQLVariable("name", strings.TrimSpace(name)), typedGraphQLVariable("number", number)}}
	if strings.TrimSpace(cursor) != "" {
		request.Variables = append(request.Variables, typedGraphQLVariable("cursor", strings.TrimSpace(cursor)))
	}

	result, err := client.queryGraphQL(request)
	if err != nil {
		return pullRequestReviewThreadsPage{}, err
	}

	return parsePullRequestReviewThreadsPage(result.Stdout)
}

func (client *PullRequestDetailService) pullRequestReviewThreadCommentsPage(threadID string, cursor string) (pullRequestReviewThreadCommentsPage, error) {
	trimmedThreadID := strings.TrimSpace(threadID)
	trimmedCursor := strings.TrimSpace(cursor)
	if trimmedThreadID == "" || trimmedCursor == "" {
		return pullRequestReviewThreadCommentsPage{}, ErrInvalidPullRequestReviewThreadsResponse
	}

	result, err := client.queryGraphQL(GraphQLRequest{Query: pullRequestReviewThreadCommentsQuery, Variables: []GraphQLVariable{typedGraphQLVariable("threadID", trimmedThreadID), typedGraphQLVariable("cursor", trimmedCursor)}})
	if err != nil {
		return pullRequestReviewThreadCommentsPage{}, err
	}

	return parsePullRequestReviewThreadCommentsPage(result.Stdout)
}

func parsePullRequestReviewThreadsPage(stdout []byte) (pullRequestReviewThreadsPage, error) {
	var response struct {
		Repository *struct {
			PullRequest *struct {
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
			} `json:"pullRequest"`
		} `json:"repository"`
	}
	if err := decodeEndpointGraphQLResponse(stdout, &response, ErrInvalidPullRequestReviewThreadsResponse); err != nil {
		return pullRequestReviewThreadsPage{}, err
	}
	if response.Repository == nil || response.Repository.PullRequest == nil {
		return pullRequestReviewThreadsPage{}, ErrInvalidPullRequestReviewThreadsResponse
	}

	return mapPullRequestReviewThreadsPageDTO(response.Repository.PullRequest.ReviewThreads), nil
}

func mapPullRequestReviewThreadsPageDTO(reviewThreads struct {
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
}) pullRequestReviewThreadsPage {
	page := pullRequestReviewThreadsPage{
		Threads:     make([]pullRequestReviewThreadPageNode, 0, len(reviewThreads.Nodes)),
		HasNextPage: reviewThreads.PageInfo.HasNextPage,
		EndCursor:   strings.TrimSpace(reviewThreads.PageInfo.EndCursor),
	}
	for _, node := range reviewThreads.Nodes {
		page.Threads = append(page.Threads, pullRequestReviewThreadPageNode{
			Thread: PullRequestReviewThread{
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
			}.normalized(),
			CommentsHasNextPage: node.Comments.PageInfo.HasNextPage,
			CommentsEndCursor:   strings.TrimSpace(node.Comments.PageInfo.EndCursor),
		})
	}
	return page
}

func parsePullRequestReviewThreadCommentsPage(stdout []byte) (pullRequestReviewThreadCommentsPage, error) {
	var response struct {
		Node *struct {
			Comments struct {
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []PullRequestComment `json:"nodes"`
			} `json:"comments"`
		} `json:"node"`
	}
	if err := decodeEndpointGraphQLResponse(stdout, &response, ErrInvalidPullRequestReviewThreadsResponse); err != nil {
		return pullRequestReviewThreadCommentsPage{}, err
	}
	if response.Node == nil {
		return pullRequestReviewThreadCommentsPage{}, ErrInvalidPullRequestReviewThreadsResponse
	}

	return mapPullRequestReviewThreadCommentsPageDTO(response.Node.Comments), nil
}

func mapPullRequestReviewThreadCommentsPageDTO(comments struct {
	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
	Nodes []PullRequestComment `json:"nodes"`
}) pullRequestReviewThreadCommentsPage {
	return pullRequestReviewThreadCommentsPage{
		Comments:    normalizePullRequestComments(comments.Nodes),
		HasNextPage: comments.PageInfo.HasNextPage,
		EndCursor:   strings.TrimSpace(comments.PageInfo.EndCursor),
	}
}

func normalizePullRequestComments(comments []PullRequestComment) []PullRequestComment {
	normalizedComments := make([]PullRequestComment, 0, len(comments))
	for _, comment := range comments {
		normalizedComments = append(normalizedComments, comment.normalized())
	}
	return normalizedComments
}

func (thread PullRequestReviewThread) normalized() PullRequestReviewThread {
	thread.ID = strings.TrimSpace(thread.ID)
	thread.Path = strings.TrimSpace(thread.Path)
	thread.DiffSide = strings.ToUpper(strings.TrimSpace(thread.DiffSide))
	thread.StartDiffSide = strings.ToUpper(strings.TrimSpace(thread.StartDiffSide))
	thread.Comments = normalizePullRequestComments(thread.Comments)
	return thread
}
