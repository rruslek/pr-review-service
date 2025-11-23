package database

import (
	"database/sql"
	"pr-review-service/internal/models"
)

func (db *DB) GetTeam(teamName string) (*models.Team, error) {
	team := &models.Team{TeamName: teamName}

	rows, err := db.Query(`
		SELECT user_id, username, is_active 
		FROM users 
		WHERE team_name = $1 
		ORDER BY user_id
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var member models.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, err
		}
		team.Members = append(team.Members, member)
	}

	if len(team.Members) == 0 {
		return nil, sql.ErrNoRows
	}

	return team, nil
}

func (db *DB) TeamExists(teamName string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", teamName).Scan(&exists)
	return exists, err
}

func (db *DB) CreateTeam(team *models.Team) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("INSERT INTO teams (team_name) VALUES ($1) ON CONFLICT DO NOTHING", team.TeamName); err != nil {
		return err
	}

	for _, member := range team.Members {
		_, err := tx.Exec(`
			INSERT INTO users (user_id, username, team_name, is_active) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) 
			DO UPDATE SET username = $2, team_name = $3, is_active = $4
		`, member.UserID, member.Username, team.TeamName, member.IsActive)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) BulkDeactivateTeamUsers(teamName string) ([]string, error) {
	rows, err := db.Query(`
		UPDATE users 
		SET is_active = false 
		WHERE team_name = $1 AND is_active = true
		RETURNING user_id
	`, teamName)
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

func (db *DB) GetOpenPRsWithReviewers(deactivatedUserIDs []string) (map[string][]string, map[string]string, error) {
	if len(deactivatedUserIDs) == 0 {
		return make(map[string][]string), make(map[string]string), nil
	}

	query := `
		SELECT 
			pr.pull_request_id, 
			pr.author_id,
			array_agg(prr.reviewer_id) as reviewer_ids
		FROM pull_requests pr
		INNER JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE pr.status = 'OPEN' AND prr.reviewer_id = ANY($1)
		GROUP BY pr.pull_request_id, pr.author_id
	`

	rows, err := db.Query(query, deactivatedUserIDs)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	prReviewers := make(map[string][]string)
	prAuthors := make(map[string]string)

	for rows.Next() {
		var prID, authorID string
		var reviewerIDs []string
		if err := rows.Scan(&prID, &authorID, &reviewerIDs); err != nil {
			return nil, nil, err
		}
		prReviewers[prID] = reviewerIDs
		prAuthors[prID] = authorID
	}

	return prReviewers, prAuthors, nil
}

func (db *DB) GetActiveTeamMembersForReplacement(teamName string, excludeUserIDs []string) ([]string, error) {
	query := `
		SELECT user_id 
		FROM users 
		WHERE team_name = $1 AND is_active = true AND user_id != ALL($2)
		ORDER BY user_id
	`

	rows, err := db.Query(query, teamName, excludeUserIDs)
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

func (db *DB) BulkReassignReviewers(reassignments map[string]map[string]string) ([]string, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var reassignedPRs []string

	for prID, replacements := range reassignments {
		for oldReviewerID, newReviewerID := range replacements {
			_, err = tx.Exec(`
				DELETE FROM pr_reviewers 
				WHERE pull_request_id = $1 AND reviewer_id = $2
			`, prID, oldReviewerID)
			if err != nil {
				return nil, err
			}

			_, err = tx.Exec(`
				INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
				VALUES ($1, $2)
				ON CONFLICT (pull_request_id, reviewer_id) DO NOTHING
			`, prID, newReviewerID)
			if err != nil {
				return nil, err
			}
		}
		reassignedPRs = append(reassignedPRs, prID)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return reassignedPRs, nil
}

func (db *DB) GetTeamNameForUsers(userIDs []string) (map[string]string, error) {
	if len(userIDs) == 0 {
		return make(map[string]string), nil
	}

	query := `
		SELECT user_id, team_name 
		FROM users 
		WHERE user_id = ANY($1)
	`

	rows, err := db.Query(query, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userTeams := make(map[string]string)
	for rows.Next() {
		var userID, teamName string
		if err := rows.Scan(&userID, &teamName); err != nil {
			return nil, err
		}
		userTeams[userID] = teamName
	}

	return userTeams, nil
}
