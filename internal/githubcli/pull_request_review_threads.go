package githubcli

import (
	"encoding/json"
	"fmt"
	"strings"
)

const pullRequestReviewThreadsQuery = `query($owner:String!,$name:String!,$number:Int!,$cursor:String){repository(owner:$owner,name:$name){pullRequest(number:$number){reviewThreads(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{id isResolved isOutdated viewerCanResolve viewerCanUnresolve path line originalLine startLine originalStartLine diffSide startDiffSide comments(first:100){pageInfo{hasNextPage endCursor}nodes{author{login} body createdAt url diffHunk}}}}}}}`
const pullRequestReviewThreadCommentsQuery = `query($threadID:ID!,$cursor:String!){node(id:$threadID){... on PullRequestReviewThread{comments(first:100,after:$cursor){pageInfo{hasNextPage endCursor}nodes{author{login} body createdAt url diffHunk}}}}}`

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

type pullRequestReviewThreadsResponse struct {
	Data struct {
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
	} `json:"data"`
}

type pullRequestReviewThreadCommentsPage struct {
	Comments    []PullRequestComment
	HasNextPage bool
	EndCursor   string
}

type pullRequestReviewThreadCommentsResponse struct {
	Data *struct {
		Node *struct {
			Comments struct {
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []PullRequestComment `json:"nodes"`
			} `json:"comments"`
		} `json:"node"`
	} `json:"data"`
}

func (client *Client) listPullRequestReviewThreads(repository string, number int) ([]PullRequestReviewThread, error) {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return nil, err
	}

	owner, name, err := splitRepositoryOwnerAndName(trimmedRepository)
	if err != nil {
		return nil, err
	}

	threads := make([]PullRequestReviewThread, 0)
	cursor := ""
	for {
		page, err := client.pullRequestReviewThreadsPage(owner, name, number, cursor)
		if err != nil {
			return nil, err
		}

		for _, threadNode := range page.Threads {
			thread := threadNode.Thread
			if threadNode.CommentsHasNextPage {
				if strings.TrimSpace(thread.ID) == "" || strings.TrimSpace(threadNode.CommentsEndCursor) == "" {
					return nil, ErrInvalidPullRequestReviewThreadsResponse
				}
				remainingComments, err := client.pullRequestReviewThreadCommentsAfter(thread.ID, threadNode.CommentsEndCursor)
				if err != nil {
					return nil, err
				}
				thread.Comments = append(thread.Comments, remainingComments...)
			}
			threads = append(threads, thread.normalized())
		}

		if !page.HasNextPage {
			return threads, nil
		}
		if strings.TrimSpace(page.EndCursor) == "" {
			return nil, ErrInvalidPullRequestReviewThreadsResponse
		}
		cursor = page.EndCursor
	}
}

func (client *Client) pullRequestReviewThreadsPage(owner string, name string, number int, cursor string) (pullRequestReviewThreadsPage, error) {
	args := []string{
		"api",
		"graphql",
		"-f",
		"query=" + pullRequestReviewThreadsQuery,
		"-F",
		"owner=" + strings.TrimSpace(owner),
		"-F",
		"name=" + strings.TrimSpace(name),
		"-F",
		fmt.Sprintf("number=%d", number),
	}
	if strings.TrimSpace(cursor) != "" {
		args = append(args, "-F", "cursor="+strings.TrimSpace(cursor))
	}

	result, err := client.runGH("gh api graphql", args...)
	if err != nil {
		return pullRequestReviewThreadsPage{}, err
	}

	return parsePullRequestReviewThreadsPage(result.Stdout)
}

func (client *Client) pullRequestReviewThreadCommentsAfter(threadID string, cursor string) ([]PullRequestComment, error) {
	comments := make([]PullRequestComment, 0)
	nextCursor := strings.TrimSpace(cursor)
	for nextCursor != "" {
		page, err := client.pullRequestReviewThreadCommentsPage(threadID, nextCursor)
		if err != nil {
			return nil, err
		}
		comments = append(comments, page.Comments...)
		if !page.HasNextPage {
			return comments, nil
		}
		if strings.TrimSpace(page.EndCursor) == "" {
			return nil, ErrInvalidPullRequestReviewThreadsResponse
		}
		nextCursor = page.EndCursor
	}
	return comments, nil
}

func (client *Client) pullRequestReviewThreadCommentsPage(threadID string, cursor string) (pullRequestReviewThreadCommentsPage, error) {
	trimmedThreadID := strings.TrimSpace(threadID)
	trimmedCursor := strings.TrimSpace(cursor)
	if trimmedThreadID == "" || trimmedCursor == "" {
		return pullRequestReviewThreadCommentsPage{}, ErrInvalidPullRequestReviewThreadsResponse
	}

	result, err := client.runGH(
		"gh api graphql",
		"api",
		"graphql",
		"-f",
		"query="+pullRequestReviewThreadCommentsQuery,
		"-F",
		"threadID="+trimmedThreadID,
		"-F",
		"cursor="+trimmedCursor,
	)
	if err != nil {
		return pullRequestReviewThreadCommentsPage{}, err
	}

	return parsePullRequestReviewThreadCommentsPage(result.Stdout)
}

func parsePullRequestReviewThreadsPage(stdout []byte) (pullRequestReviewThreadsPage, error) {
	var response pullRequestReviewThreadsResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return pullRequestReviewThreadsPage{}, fmt.Errorf("%w: %v", ErrInvalidPullRequestReviewThreadsResponse, err)
	}
	if response.Data.Repository == nil || response.Data.Repository.PullRequest == nil {
		return pullRequestReviewThreadsPage{}, ErrInvalidPullRequestReviewThreadsResponse
	}

	reviewThreads := response.Data.Repository.PullRequest.ReviewThreads
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
			},
			CommentsHasNextPage: node.Comments.PageInfo.HasNextPage,
			CommentsEndCursor:   strings.TrimSpace(node.Comments.PageInfo.EndCursor),
		})
	}

	return page, nil
}

func parsePullRequestReviewThreadCommentsPage(stdout []byte) (pullRequestReviewThreadCommentsPage, error) {
	var response pullRequestReviewThreadCommentsResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return pullRequestReviewThreadCommentsPage{}, fmt.Errorf("%w: %v", ErrInvalidPullRequestReviewThreadsResponse, err)
	}
	if response.Data == nil || response.Data.Node == nil {
		return pullRequestReviewThreadCommentsPage{}, ErrInvalidPullRequestReviewThreadsResponse
	}

	comments := response.Data.Node.Comments
	return pullRequestReviewThreadCommentsPage{
		Comments:    normalizePullRequestComments(comments.Nodes),
		HasNextPage: comments.PageInfo.HasNextPage,
		EndCursor:   strings.TrimSpace(comments.PageInfo.EndCursor),
	}, nil
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
