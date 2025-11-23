package handlers

import (
	"encoding/json"
	"net/http"
	"pr-review-service/internal/models"
)

// POST /team/add
func (h *Handlers) AddTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var team models.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.service.CreateTeam(&team); err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{"team": team})
}

// GET /team/get
func (h *Handlers) GetTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "team_name parameter is required")
		return
	}

	team, err := h.service.GetTeam(teamName)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, team)
}

// POST /team/bulkDeactivate
func (h *Handlers) BulkDeactivateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.BulkDeactivateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.TeamName == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "team_name is required")
		return
	}

	result, err := h.service.BulkDeactivateTeamUsers(req.TeamName)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}
