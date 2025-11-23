package service

import (
	"math/rand"
	"pr-review-service/internal/database"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

func selectRandomReviewers(candidates []string, n int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	if len(candidates) <= n {
		return candidates
	}

	shuffled := make([]string, len(candidates))
	copy(shuffled, candidates)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:n]
}

func (s *Service) HealthCheck() error {
	return s.db.Ping()
}
