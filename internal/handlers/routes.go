package handlers

import (
	"net/http"
)

func (h *Handlers) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "static/index.html")
		} else {
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/team/add", h.AddTeam)
	mux.HandleFunc("/team/get", h.GetTeam)
	mux.HandleFunc("/users/setIsActive", h.SetUserActive)
	mux.HandleFunc("/users/getReview", h.GetUserReview)
	mux.HandleFunc("/pullRequest/create", h.CreatePullRequest)
	mux.HandleFunc("/pullRequest/merge", h.MergePullRequest)
	mux.HandleFunc("/pullRequest/reassign", h.ReassignReviewer)
	mux.HandleFunc("/health", h.HealthCheck)
}
