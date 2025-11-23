package service

import (
	"errors"
	"pr-review-service/internal/models"
)

var (
	ErrPRExists    = errors.New("PR already exists")
	ErrPRNotFound  = errors.New("PR not found")
	ErrPRMerged    = errors.New("PR is merged")
	ErrNotAssigned = errors.New("reviewer is not assigned")
	ErrNoCandidate = errors.New("no active replacement candidate")
)

func (s *Service) CreatePullRequest(prID, prName, authorID string) (*models.PullRequest, error) {
	exists, err := s.db.PRExists(prID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrPRExists
	}

	author, err := s.db.GetUser(authorID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	candidates, err := s.db.GetActiveTeamMembers(author.TeamName, authorID)
	if err != nil {
		return nil, err
	}

	reviewers := selectRandomReviewers(candidates, 2)

	if err := s.db.CreatePullRequest(prID, prName, authorID, reviewers); err != nil {
		return nil, err
	}

	return s.db.GetPullRequest(prID)
}

func (s *Service) MergePullRequest(prID string) (*models.PullRequest, error) {
	_, err := s.db.GetPullRequest(prID)
	if err != nil {
		return nil, ErrPRNotFound
	}

	if err := s.db.MergePullRequest(prID); err != nil {
		return nil, err
	}

	return s.db.GetPullRequest(prID)
}

func (s *Service) ReassignReviewer(prID, oldReviewerID string) (*models.PullRequest, string, error) {
	pr, err := s.db.GetPullRequest(prID)
	if err != nil {
		return nil, "", ErrPRNotFound
	}

	if pr.Status == "MERGED" {
		return nil, "", ErrPRMerged
	}

	isAssigned, err := s.db.IsReviewerAssigned(prID, oldReviewerID)
	if err != nil {
		return nil, "", err
	}
	if !isAssigned {
		return nil, "", ErrNotAssigned
	}

	oldReviewer, err := s.db.GetUser(oldReviewerID)
	if err != nil {
		return nil, "", ErrUserNotFound
	}

	candidates, err := s.db.GetActiveTeamMembers(oldReviewer.TeamName, oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	var filteredCandidates []string
	for _, candidate := range candidates {
		if candidate != pr.AuthorID {
			filteredCandidates = append(filteredCandidates, candidate)
		}
	}

	if len(filteredCandidates) == 0 {
		return nil, "", ErrNoCandidate
	}

	newReviewerID := selectRandomReviewers(filteredCandidates, 1)[0]

	if err := s.db.ReassignReviewer(prID, oldReviewerID, newReviewerID); err != nil {
		return nil, "", err
	}

	updatedPR, err := s.db.GetPullRequest(prID)
	if err != nil {
		return nil, "", err
	}

	return updatedPR, newReviewerID, nil
}
