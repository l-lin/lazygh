package githubcli

type PullRequestDiffAssembler struct {
	LoadUnifiedDiff   func(repository string, number int) (string, error)
	LoadDiffFiles     func(repository string, number int) ([]PullRequestDiffFile, error)
	ListReviewThreads func(repository string, number int) ([]PullRequestReviewThread, error)
}

func newPullRequestDiffAssembler(service *PullRequestDetailService) PullRequestDiffAssembler {
	return PullRequestDiffAssembler{
		LoadUnifiedDiff:   service.getPullRequestUnifiedDiff,
		LoadDiffFiles:     service.listPullRequestDiffFiles,
		ListReviewThreads: service.listPullRequestReviewThreads,
	}
}

func (assembler PullRequestDiffAssembler) Assemble(repository string, number int) (PullRequestDiff, error) {
	if assembler.LoadUnifiedDiff == nil || assembler.LoadDiffFiles == nil {
		return PullRequestDiff{}, ErrInvalidPullRequestDiffFilesResponse
	}

	unifiedDiff, err := assembler.LoadUnifiedDiff(repository, number)
	if err != nil {
		return PullRequestDiff{}, err
	}
	files, err := assembler.LoadDiffFiles(repository, number)
	if err != nil {
		return PullRequestDiff{}, err
	}
	threads := []PullRequestReviewThread(nil)
	if assembler.ListReviewThreads != nil {
		threads, err = assembler.ListReviewThreads(repository, number)
		if err != nil {
			return PullRequestDiff{}, err
		}
	}
	return PullRequestDiff{UnifiedDiff: unifiedDiff, Files: files, Threads: threads}, nil
}
