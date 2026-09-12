package cache

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	appconfig "github.com/l-lin/lazygh/internal/config"
	githubcli "github.com/l-lin/lazygh/internal/github"
)

type PullRequestFreshness struct {
	Repository    string
	Number        int
	Seen          bool
	SeenUpdatedAt string
}

func reconcilePullRequestSearchMemberships(transaction *sql.Tx, search appconfig.PullRequestSearch, pullRequests []githubcli.PullRequest) error {
	searchKey := pullRequestSearchKey(search)
	if _, actualErr := transaction.Exec(`DELETE FROM pull_request_search_memberships WHERE search_key = ?`, searchKey); actualErr != nil {
		return actualErr
	}

	for _, pullRequest := range pullRequests {
		repository, number, actualErr := normalizePullRequestIdentity(pullRequest.Repository.NameWithOwner, pullRequest.Number)
		if actualErr != nil {
			continue
		}
		if _, actualErr = transaction.Exec(`
			INSERT OR IGNORE INTO pull_request_search_memberships (search_key, repository, number)
			VALUES (?, ?, ?)
		`, searchKey, repository, number); actualErr != nil {
			return actualErr
		}
	}
	return nil
}

func (store *Store) PullRequestFreshness() ([]PullRequestFreshness, error) {
	if store == nil || store.db == nil {
		return nil, nil
	}

	rows, actualErr := store.db.Query(`
		SELECT repository, number, seen, seen_updated_at
		FROM pull_request_freshness
		ORDER BY repository, number
	`)
	if actualErr != nil {
		return nil, actualErr
	}
	defer rows.Close()

	freshness := make([]PullRequestFreshness, 0)
	for rows.Next() {
		var entry PullRequestFreshness
		var seen int
		if actualErr := rows.Scan(&entry.Repository, &entry.Number, &seen, &entry.SeenUpdatedAt); actualErr != nil {
			return nil, actualErr
		}
		entry.Seen = seen != 0
		entry.Repository = strings.TrimSpace(entry.Repository)
		entry.SeenUpdatedAt = strings.TrimSpace(entry.SeenUpdatedAt)
		freshness = append(freshness, entry)
	}
	if actualErr := rows.Err(); actualErr != nil {
		return nil, actualErr
	}
	return freshness, nil
}

func (store *Store) MarkPullRequestSeen(repository string, number int, updatedAt string) error {
	if store == nil || store.db == nil {
		return nil
	}

	trimmedRepository, normalizedNumber, actualErr := normalizePullRequestIdentity(repository, number)
	if actualErr != nil {
		return actualErr
	}
	trimmedUpdatedAt := strings.TrimSpace(updatedAt)

	transaction, actualErr := store.db.Begin()
	if actualErr != nil {
		return actualErr
	}
	defer rollbackOnFailure(transaction)

	var existingUpdatedAt string
	actualErr = transaction.QueryRow(`
		SELECT seen_updated_at
		FROM pull_request_freshness
		WHERE repository = ? AND number = ?
	`, trimmedRepository, normalizedNumber).Scan(&existingUpdatedAt)
	switch {
	case errors.Is(actualErr, sql.ErrNoRows):
		_, actualErr = transaction.Exec(`
			INSERT INTO pull_request_freshness (repository, number, seen, seen_updated_at)
			VALUES (?, ?, 1, ?)
		`, trimmedRepository, normalizedNumber, trimmedUpdatedAt)
	case actualErr != nil:
		return actualErr
	default:
		if comparePullRequestFreshnessVersions(trimmedUpdatedAt, existingUpdatedAt) > 0 {
			existingUpdatedAt = trimmedUpdatedAt
		}
		_, actualErr = transaction.Exec(`
			UPDATE pull_request_freshness
			SET seen = 1, seen_updated_at = ?
			WHERE repository = ? AND number = ?
		`, existingUpdatedAt, trimmedRepository, normalizedNumber)
	}
	if actualErr != nil {
		return actualErr
	}
	return transaction.Commit()
}

func comparePullRequestFreshnessVersions(left string, right string) int {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == right {
		return 0
	}

	leftTime, leftErr := time.Parse(time.RFC3339Nano, left)
	rightTime, rightErr := time.Parse(time.RFC3339Nano, right)
	if leftErr == nil && rightErr == nil {
		switch {
		case leftTime.Before(rightTime):
			return -1
		case leftTime.After(rightTime):
			return 1
		default:
			return 0
		}
	}
	return strings.Compare(left, right)
}

func (store *Store) ReconcilePullRequestSearches(searches []appconfig.PullRequestSearch) error {
	if store == nil || store.db == nil {
		return nil
	}

	searchKeys := make([]string, 0, len(searches))
	for _, search := range searches {
		searchKeys = append(searchKeys, pullRequestSearchKey(search))
	}

	transaction, actualErr := store.db.Begin()
	if actualErr != nil {
		return actualErr
	}
	defer rollbackOnFailure(transaction)

	if len(searchKeys) == 0 {
		if _, actualErr = transaction.Exec(`DELETE FROM pull_request_search_memberships`); actualErr != nil {
			return actualErr
		}
	} else {
		placeholders := strings.TrimRight(strings.Repeat("?,", len(searchKeys)), ",")
		arguments := make([]any, 0, len(searchKeys))
		for _, searchKey := range searchKeys {
			arguments = append(arguments, searchKey)
		}
		if _, actualErr = transaction.Exec(fmt.Sprintf(`DELETE FROM pull_request_search_memberships WHERE search_key NOT IN (%s)`, placeholders), arguments...); actualErr != nil {
			return actualErr
		}
	}

	for _, search := range searches {
		var payloadJSON string
		actualErr = transaction.QueryRow(`
			SELECT payload_json
			FROM pull_request_lists
			WHERE search_key = ?
		`, pullRequestSearchKey(search)).Scan(&payloadJSON)
		if errors.Is(actualErr, sql.ErrNoRows) {
			if actualErr = reconcilePullRequestSearchMemberships(transaction, search, nil); actualErr != nil {
				return actualErr
			}
			continue
		}
		if actualErr != nil {
			return actualErr
		}

		var pullRequests []githubcli.PullRequest
		if actualErr = json.Unmarshal([]byte(payloadJSON), &pullRequests); actualErr != nil {
			return fmt.Errorf("decode cached pull requests: %w", actualErr)
		}
		if actualErr = reconcilePullRequestSearchMemberships(transaction, search, pullRequests); actualErr != nil {
			return actualErr
		}
	}

	if actualErr = deleteUnreferencedPullRequestFreshness(transaction); actualErr != nil {
		return actualErr
	}
	return transaction.Commit()
}

func deleteUnreferencedPullRequestFreshness(transaction *sql.Tx) error {
	_, actualErr := transaction.Exec(`
		DELETE FROM pull_request_freshness
		WHERE NOT EXISTS (
			SELECT 1
			FROM pull_request_search_memberships membership
			WHERE membership.repository = pull_request_freshness.repository
			  AND membership.number = pull_request_freshness.number
		)
	`)
	return actualErr
}
