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
