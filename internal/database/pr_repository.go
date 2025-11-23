package database

import (
	"database/sql"
	"pr-review-service/internal/models"
	"time"
)

func (db *DB) GetPullRequest(prID string) (*models.PullRequest, error) {
	pr := &models.PullRequest{}
	var createdAt, mergedAt sql.NullTime

	err := db.QueryRow(`
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&createdAt,
		&mergedAt,
	)
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		pr.CreatedAt = &createdAt.Time
	}
	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	rows, err := db.Query(`
		SELECT reviewer_id 
		FROM pr_reviewers 
		WHERE pull_request_id = $1
		ORDER BY reviewer_id
	`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	return pr, nil
}

func (db *DB) PRExists(prID string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)", prID).Scan(&exists)
	return exists, err
}

func (db *DB) CreatePullRequest(prID, prName, authorID string, reviewers []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
		VALUES ($1, $2, $3, 'OPEN')
	`, prID, prName, authorID)
	if err != nil {
		return err
	}

	for _, reviewerID := range reviewers {
		_, err = tx.Exec(`
			INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
			VALUES ($1, $2)
		`, prID, reviewerID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) MergePullRequest(prID string) error {
	now := time.Now()
	_, err := db.Exec(`
		UPDATE pull_requests 
		SET status = 'MERGED', merged_at = COALESCE(merged_at, $1)
		WHERE pull_request_id = $2
	`, now, prID)
	return err
}

func (db *DB) IsReviewerAssigned(prID, userID string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2)
	`, prID, userID).Scan(&exists)
	return exists, err
}

func (db *DB) ReassignReviewer(prID, oldReviewerID, newReviewerID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		DELETE FROM pr_reviewers 
		WHERE pull_request_id = $1 AND reviewer_id = $2
	`, prID, oldReviewerID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
		VALUES ($1, $2)
	`, prID, newReviewerID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
