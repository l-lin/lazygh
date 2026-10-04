package githubcli

import "sync"

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
	return assembler.AssembleProgressively(repository, number, nil)
}

func (assembler PullRequestDiffAssembler) AssembleProgressively(repository string, number int, publish func(PullRequestDiff)) (PullRequestDiff, error) {
	if assembler.LoadUnifiedDiff == nil || assembler.LoadDiffFiles == nil {
		return PullRequestDiff{}, ErrInvalidPullRequestDiffFilesResponse
	}

	var unifiedDiff string
	var files []PullRequestDiffFile
	var diffErr, filesErr error
	var reads sync.WaitGroup
	reads.Go(func() { unifiedDiff, diffErr = assembler.LoadUnifiedDiff(repository, number) })
	reads.Go(func() { files, filesErr = assembler.LoadDiffFiles(repository, number) })
	reads.Wait()
	if diffErr != nil {
		return PullRequestDiff{}, diffErr
	}
	if filesErr != nil {
		return PullRequestDiff{}, filesErr
	}
	var err error
	if publish != nil {
		publish(PullRequestDiff{UnifiedDiff: unifiedDiff, Files: files})
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
