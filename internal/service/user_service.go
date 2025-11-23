package service

import (
	"errors"
	"pr-review-service/internal/models"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

func (s *Service) SetUserActive(userID string, isActive bool) (*models.User, error) {
	user, err := s.db.GetUser(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if err := s.db.SetUserActive(userID, isActive); err != nil {
		return nil, err
	}

	user.IsActive = isActive
	return user, nil
}

func (s *Service) GetUser(userID string) (*models.User, error) {
	user, err := s.db.GetUser(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *Service) GetUserPullRequests(userID string) ([]models.PullRequestShort, error) {
	_, err := s.db.GetUser(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return s.db.GetUserPullRequests(userID)
}
