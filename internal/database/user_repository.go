package database

import "pr-review-service/internal/models"

func (db *DB) GetUser(userID string) (*models.User, error) {
	user := &models.User{}
	err := db.QueryRow(`
		SELECT user_id, username, team_name, is_active 
		FROM users 
		WHERE user_id = $1
	`, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *DB) SetUserActive(userID string, isActive bool) error {
	_, err := db.Exec("UPDATE users SET is_active = $1 WHERE user_id = $2", isActive, userID)
	return err
}

func (db *DB) GetActiveTeamMembers(teamName string, excludeUserID string) ([]string, error) {
	rows, err := db.Query(`
		SELECT user_id 
		FROM users 
		WHERE team_name = $1 AND is_active = true AND user_id != $2
		ORDER BY user_id
	`, teamName, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (db *DB) GetUserPullRequests(userID string) ([]models.PullRequestShort, error) {
	rows, err := db.Query(`
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		INNER JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.reviewer_id = $1
		ORDER BY pr.pull_request_id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []models.PullRequestShort
	for rows.Next() {
		var pr models.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	return prs, nil
}
