package models

type UserStats struct {
	UserID           string `json:"user_id"`
	Username         string `json:"username"`
	AssignedPRsCount int    `json:"assigned_prs_count"`
}

type PRStats struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	ReviewersCount  int    `json:"reviewers_count"`
	Status          string `json:"status"`
}

type StatsResponse struct {
	UsersStats []UserStats `json:"users_stats"`
	PRsStats   []PRStats   `json:"prs_stats"`
	TotalUsers int         `json:"total_users"`
	TotalPRs   int         `json:"total_prs"`
}
