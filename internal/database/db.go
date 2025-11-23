package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func NewDB(connStr string) (*DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &DB{db}
	if err := database.initSchema(); err != nil {
		return nil, err
	}

	return database, nil
}

func (db *DB) initSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS teams (
			team_name VARCHAR(255) PRIMARY KEY
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			user_id VARCHAR(255) PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			team_name VARCHAR(255) NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
			is_active BOOLEAN NOT NULL DEFAULT true
		)`,
		`CREATE TABLE IF NOT EXISTS pull_requests (
			pull_request_id VARCHAR(255) PRIMARY KEY,
			pull_request_name VARCHAR(255) NOT NULL,
			author_id VARCHAR(255) NOT NULL REFERENCES users(user_id),
			status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			merged_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS pr_reviewers (
			pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
			reviewer_id VARCHAR(255) NOT NULL REFERENCES users(user_id),
			PRIMARY KEY (pull_request_id, reviewer_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_team ON users(team_name)`,
		`CREATE INDEX IF NOT EXISTS idx_pr_reviewers ON pr_reviewers(pull_request_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pr_reviewers_user ON pr_reviewers(reviewer_id)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}
