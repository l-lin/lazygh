package cache

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	sqliteIncrementalAutoVacuumMode        = 2
	staleMergedClosedPullRequestDiffTTL    = 24 * time.Hour
	staleMergedClosedPullRequestDetailTTL  = 7 * 24 * time.Hour
	staleMergedClosedPullRequestSummaryTTL = 30 * 24 * time.Hour
	sqliteTimestampLayout                  = "2006-01-02 15:04:05"
)

type pullRequestPruneCandidate struct {
	repository string
	number     int
	summary    githubcli.PullRequest
	hasDetail  bool
	hasDiff    bool
	updatedAt  time.Time
}

func (store *Store) ensureIncrementalAutoVacuum() error {
	mode, err := store.autoVacuumMode()
	if err != nil {
		return err
	}
	if mode == sqliteIncrementalAutoVacuumMode {
		return nil
	}
	if _, err := store.db.Exec(`PRAGMA auto_vacuum = INCREMENTAL`); err != nil {
		return err
	}
	_, err = store.db.Exec(`VACUUM`)
	return err
}

func (store *Store) autoVacuumMode() (int, error) {
	var actual int
	if err := store.db.QueryRow(`PRAGMA auto_vacuum`).Scan(&actual); err != nil {
		return 0, err
	}
	return actual, nil
}

func (store *Store) pruneStaleMergedPullRequests() error {
	if store == nil || store.db == nil {
		return nil
	}

	referencedKeys, err := store.referencedPullRequestKeys()
	if err != nil {
		return err
	}
	candidates, err := store.pullRequestPruneCandidates()
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	transaction, err := store.db.Begin()
	if err != nil {
		return err
	}
	defer rollbackOnFailure(transaction)

	changed := false
	for _, candidate := range candidates {
		if referencedKeys[pullRequestCacheKey(candidate.repository, candidate.number)] {
			continue
		}
		if !pullRequestStatePrunable(candidate.summary.State) {
			continue
		}

		age := now.Sub(candidate.updatedAt)
		if age >= staleMergedClosedPullRequestDiffTTL && candidate.hasDiff {
			if _, err := transaction.Exec(`
				UPDATE pull_requests
				SET diff_json = NULL,
					diff_updated_at = ''
				WHERE repository = ? AND number = ?
			`, candidate.repository, candidate.number); err != nil {
				return err
			}
			candidate.hasDiff = false
			changed = true
		}
		if age >= staleMergedClosedPullRequestDetailTTL && candidate.hasDetail {
			if _, err := transaction.Exec(`
				UPDATE pull_requests
				SET detail_json = NULL,
					detail_updated_at = ''
				WHERE repository = ? AND number = ?
			`, candidate.repository, candidate.number); err != nil {
				return err
			}
			candidate.hasDetail = false
			changed = true
		}
		if age >= staleMergedClosedPullRequestSummaryTTL && !candidate.hasDetail && !candidate.hasDiff {
			if _, err := transaction.Exec(`
				DELETE FROM pull_requests
				WHERE repository = ? AND number = ?
			`, candidate.repository, candidate.number); err != nil {
				return err
			}
			changed = true
		}
	}
	if !changed {
		return nil
	}
	if err := transaction.Commit(); err != nil {
		return err
	}
	store.runIncrementalVacuumBestEffort()
	return nil
}

func (store *Store) referencedPullRequestKeys() (map[string]bool, error) {
	referencedKeys := map[string]bool{}

	listRows, err := store.db.Query(`SELECT payload_json FROM pull_request_lists`)
	if err != nil {
		return nil, err
	}
	defer listRows.Close()
	for listRows.Next() {
		var payload string
		if err := listRows.Scan(&payload); err != nil {
			return nil, err
		}
		var pullRequests []githubcli.PullRequest
		if err := json.Unmarshal([]byte(payload), &pullRequests); err != nil {
			return nil, err
		}
		for _, pullRequest := range pullRequests {
			if key := pullRequestCacheKey(pullRequest.Repository.NameWithOwner, pullRequest.Number); key != "" {
				referencedKeys[key] = true
			}
		}
	}
	if err := listRows.Err(); err != nil {
		return nil, err
	}

	notificationRows, err := store.db.Query(`SELECT payload_json FROM notification_lists`)
	if err != nil {
		return nil, err
	}
	defer notificationRows.Close()
	for notificationRows.Next() {
		var payload string
		if err := notificationRows.Scan(&payload); err != nil {
			return nil, err
		}
		var notifications []githubcli.Notification
		if err := json.Unmarshal([]byte(payload), &notifications); err != nil {
			return nil, err
		}
		for _, notification := range notifications {
			summary, ok := notification.PullRequestSummary()
			if !ok {
				continue
			}
			if key := pullRequestCacheKey(summary.Repository.NameWithOwner, summary.Number); key != "" {
				referencedKeys[key] = true
			}
		}
	}
	if err := notificationRows.Err(); err != nil {
		return nil, err
	}

	return referencedKeys, nil
}

func (store *Store) pullRequestPruneCandidates() ([]pullRequestPruneCandidate, error) {
	rows, err := store.db.Query(`
		SELECT repository, number, summary_json, detail_json, diff_json, updated_at
		FROM pull_requests
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candidates := make([]pullRequestPruneCandidate, 0)
	for rows.Next() {
		var repository string
		var number int
		var summaryJSON sql.NullString
		var detailJSON sql.NullString
		var diffJSON sql.NullString
		var updatedAt string
		if err := rows.Scan(&repository, &number, &summaryJSON, &detailJSON, &diffJSON, &updatedAt); err != nil {
			return nil, err
		}
		if !summaryJSON.Valid || strings.TrimSpace(summaryJSON.String) == "" {
			continue
		}
		parsedUpdatedAt, ok := parseSQLiteTimestamp(updatedAt)
		if !ok {
			continue
		}
		var summary githubcli.PullRequest
		if err := json.Unmarshal([]byte(summaryJSON.String), &summary); err != nil {
			return nil, err
		}
		candidates = append(candidates, pullRequestPruneCandidate{
			repository: strings.TrimSpace(repository),
			number:     number,
			summary:    summary,
			hasDetail:  detailJSON.Valid && strings.TrimSpace(detailJSON.String) != "",
			hasDiff:    diffJSON.Valid && strings.TrimSpace(diffJSON.String) != "",
			updatedAt:  parsedUpdatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return candidates, nil
}

func pullRequestStatePrunable(state string) bool {
	switch strings.ToUpper(strings.TrimSpace(state)) {
	case "MERGED", "CLOSED":
		return true
	default:
		return false
	}
}

func parseSQLiteTimestamp(value string) (time.Time, bool) {
	parsed, err := time.ParseInLocation(sqliteTimestampLayout, strings.TrimSpace(value), time.UTC)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func pullRequestCacheKey(repository string, number int) string {
	trimmedRepository, normalizedNumber, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return ""
	}
	return trimmedRepository + "#" + strconv.Itoa(normalizedNumber)
}

func (store *Store) runIncrementalVacuumBestEffort() {
	if store == nil || store.db == nil {
		return
	}
	_, _ = store.db.Exec(`PRAGMA incremental_vacuum`)
}
