package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"pr-review-service/internal/models"
	"pr-review-service/internal/service"
)

type Handlers struct {
	service *service.Service
}

func NewHandlers(svc *service.Service) *Handlers {
	return &Handlers{service: svc}
}

func (h *Handlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func (h *Handlers) writeError(w http.ResponseWriter, status int, code, message string) {
	h.writeJSON(w, status, models.ErrorResponse{
		Error: models.ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

func (h *Handlers) handleServiceError(w http.ResponseWriter, err error) {
	switch err {
	case service.ErrTeamExists:
		h.writeError(w, http.StatusBadRequest, "TEAM_EXISTS", "team_name already exists")
	case service.ErrTeamNotFound, service.ErrUserNotFound, service.ErrPRNotFound:
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
	case service.ErrPRExists:
		h.writeError(w, http.StatusConflict, "PR_EXISTS", "PR id already exists")
	case service.ErrPRMerged:
		h.writeError(w, http.StatusConflict, "PR_MERGED", "cannot reassign on merged PR")
	case service.ErrNotAssigned:
		h.writeError(w, http.StatusConflict, "NOT_ASSIGNED", "reviewer is not assigned to this PR")
	case service.ErrNoCandidate:
		h.writeError(w, http.StatusConflict, "NO_CANDIDATE", "no active replacement candidate in team")
	default:
		log.Printf("Unexpected error: %v", err)
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}
