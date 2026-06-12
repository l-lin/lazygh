package githubcli

import (
	"errors"
	"reflect"
	"testing"
)

func TestParsePullRequestMergeQueueMetadata_GivenWhitespaceAndQueueEntry_WhenParsing_ThenItReturnsNormalizedMetadata(t *testing.T) {
	subject := []byte(`{"data":{"repository":{"pullRequest":{"isMergeQueueEnabled":true,"isInMergeQueue":true,"mergeQueueEntry":{"id":" MQE_1 ","state":" QUEUED ","position":2,"estimatedTimeToMerge":15}}}}}`)

	actual, actualErr := parsePullRequestMergeQueueMetadata(subject)

	then_noError(t, actualErr)
	expected := pullRequestMergeQueueMetadata{
		IsMergeQueueEnabled: true,
		IsInMergeQueue:      true,
		MergeQueueEntry: &PullRequestMergeQueueEntry{
			ID:                   "MQE_1",
			State:                "QUEUED",
			Position:             2,
			EstimatedTimeToMerge: 15,
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected merge queue metadata %+v, actual %+v", expected, actual)
	}
}

func TestEnqueuePullRequest_GivenPullRequestID_WhenSubmitting_ThenItRunsTheQueueMutation(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"enqueuePullRequest":{"mergeQueueEntry":{"id":"MQE_1","state":"QUEUED","position":1}}}}`)}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.EnqueuePullRequest(" PR_kwDOA ")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"api", "graphql", "-f", "query=" + enqueuePullRequestMutation, "-F", "pullRequestId=PR_kwDOA"})
}

func TestDequeuePullRequest_GivenPullRequestID_WhenSubmitting_ThenItRunsTheQueueMutation(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"dequeuePullRequest":{"mergeQueueEntry":{"id":"MQE_1","state":"DEQUEUED","position":1}}}}`)}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.DequeuePullRequest(" PR_kwDOA ")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"api", "graphql", "-f", "query=" + dequeuePullRequestMutation, "-F", "pullRequestId=PR_kwDOA"})
}

func TestEnqueuePullRequest_GivenMissingQueueEntryID_WhenSubmitting_ThenItReturnsTheInvalidResponseError(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"enqueuePullRequest":{"mergeQueueEntry":{"id":"   ","state":"QUEUED","position":1}}}}`)}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.EnqueuePullRequest("PR_kwDOA")

	if !errors.Is(actualErr, ErrInvalidPullRequestMergeQueueMutationResponse) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidPullRequestMergeQueueMutationResponse, actualErr)
	}
}
