package githubcli

type pullRequestDetailDTO PullRequestDetail

type pullRequestInlineCommentDTO PullRequestInlineComment

type pullRequestSearchResultDTO PullRequest

type pullRequestDiffFileDTO PullRequestDiffFile

func parsePullRequestDetailResponse(stdout []byte) (PullRequestDetail, error) {
	var dto pullRequestDetailDTO
	if err := decodeEndpointJSONResponse(stdout, &dto, ErrInvalidPullRequestDetailResponse); err != nil {
		return PullRequestDetail{}, err
	}
	return PullRequestDetail(dto).normalized(), nil
}

func parsePullRequestInlineCommentsResponse(stdout []byte) ([]PullRequestInlineComment, error) {
	var dtos []pullRequestInlineCommentDTO
	if err := decodeEndpointPaginatedOrFlatJSONResponse(stdout, &dtos, ErrInvalidPullRequestInlineCommentResponse); err != nil {
		return nil, err
	}
	return mapPullRequestInlineCommentDTOs(dtos), nil
}

func parsePullRequestSearchResultsResponse(stdout []byte) ([]PullRequest, error) {
	var dtos []pullRequestSearchResultDTO
	if err := decodeEndpointJSONResponse(stdout, &dtos, ErrInvalidPullRequestResponse); err != nil {
		return nil, err
	}
	return mapPullRequestSearchResultDTOs(dtos), nil
}

func parsePullRequestDiffFilesResponse(stdout []byte) ([]PullRequestDiffFile, error) {
	var dtos []pullRequestDiffFileDTO
	if err := decodeEndpointPaginatedOrFlatJSONResponse(stdout, &dtos, ErrInvalidPullRequestDiffFilesResponse); err != nil {
		return nil, err
	}
	return mapPullRequestDiffFileDTOs(dtos), nil
}

func mapPullRequestInlineCommentDTOs(dtos []pullRequestInlineCommentDTO) []PullRequestInlineComment {
	if len(dtos) == 0 {
		return nil
	}
	comments := make([]PullRequestInlineComment, 0, len(dtos))
	for _, dto := range dtos {
		comments = append(comments, PullRequestInlineComment(dto).normalized())
	}
	return comments
}

func mapPullRequestSearchResultDTOs(dtos []pullRequestSearchResultDTO) []PullRequest {
	if len(dtos) == 0 {
		return nil
	}
	pullRequests := make([]PullRequest, 0, len(dtos))
	for _, dto := range dtos {
		pullRequests = append(pullRequests, PullRequest(dto).normalized())
	}
	return pullRequests
}

func mapPullRequestDiffFileDTOs(dtos []pullRequestDiffFileDTO) []PullRequestDiffFile {
	if len(dtos) == 0 {
		return nil
	}
	files := make([]PullRequestDiffFile, 0, len(dtos))
	for _, dto := range dtos {
		files = append(files, PullRequestDiffFile(dto).normalized())
	}
	return files
}
