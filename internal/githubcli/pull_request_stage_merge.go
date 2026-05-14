package githubcli

import "strconv"

func (client *PullRequestMutationService) MarkPullRequestReadyForReview(repository string, number int) error {
	return client.runPullRequestReady(repository, number, false)
}

func (client *PullRequestMutationService) ConvertPullRequestToDraft(repository string, number int) error {
	return client.runPullRequestReady(repository, number, true)
}

func (client *PullRequestMutationService) ClosePullRequest(repository string, number int) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	if _, err := client.execute(rawCommand("pr", "close", strconv.Itoa(number), "-R", trimmedRepository)); err != nil {
		return err
	}

	return nil
}

func (client *PullRequestMutationService) ReopenPullRequest(repository string, number int) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	if _, err := client.execute(rawCommand("pr", "reopen", strconv.Itoa(number), "-R", trimmedRepository)); err != nil {
		return err
	}

	return nil
}

func (client *PullRequestMutationService) runPullRequestReady(repository string, number int, undo bool) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	arguments := []string{"pr", "ready", strconv.Itoa(number), "-R", trimmedRepository}
	if undo {
		arguments = append(arguments, "--undo")
	}
	if _, err := client.execute(rawCommand(arguments...)); err != nil {
		return err
	}

	return nil
}

func (client *PullRequestMutationService) SquashMergePullRequest(repository string, number int) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	if _, err := client.execute(rawCommand("pr", "merge", strconv.Itoa(number), "-R", trimmedRepository, "--squash")); err != nil {
		return err
	}

	return nil
}

func (client *PullRequestMutationService) UpdatePullRequestBranch(repository string, number int) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	if _, err := client.execute(rawCommand("pr", "update-branch", strconv.Itoa(number), "-R", trimmedRepository)); err != nil {
		return err
	}

	return nil
}
