package models

type BulkDeactivateRequest struct {
	TeamName string `json:"team_name"`
}

type BulkDeactivateResponse struct {
	TeamName         string   `json:"team_name"`
	DeactivatedUsers []string `json:"deactivated_users"`
	ReassignedPRs    []string `json:"reassigned_prs"`
	DeactivatedCount int      `json:"deactivated_count"`
	ReassignedCount  int      `json:"reassigned_count"`
}
