package database

import "pr-review-service/internal/models"

func (db *DB) GetUserStats() ([]models.UserStats, error) {
	rows, err := db.Query(`
		SELECT 
			u.user_id,
			u.username,
			COALESCE(COUNT(prr.reviewer_id), 0) as assigned_prs_count
		FROM users u
		LEFT JOIN pr_reviewers prr ON u.user_id = prr.reviewer_id
		GROUP BY u.user_id, u.username
		ORDER BY assigned_prs_count DESC, u.user_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.UserStats
	for rows.Next() {
		var s models.UserStats
		if err := rows.Scan(&s.UserID, &s.Username, &s.AssignedPRsCount); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}

func (db *DB) GetPRStats() ([]models.PRStats, error) {
	rows, err := db.Query(`
		SELECT 
			pr.pull_request_id,
			pr.pull_request_name,
			COALESCE(COUNT(prr.reviewer_id), 0) as reviewers_count,
			pr.status
		FROM pull_requests pr
		LEFT JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		GROUP BY pr.pull_request_id, pr.pull_request_name, pr.status
		ORDER BY reviewers_count DESC, pr.pull_request_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.PRStats
	for rows.Next() {
		var s models.PRStats
		if err := rows.Scan(&s.PullRequestID, &s.PullRequestName, &s.ReviewersCount, &s.Status); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}

func (db *DB) GetTotalUsersCount() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (db *DB) GetTotalPRsCount() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM pull_requests").Scan(&count)
	return count, err
}
