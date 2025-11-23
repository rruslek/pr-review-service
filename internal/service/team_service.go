package service

import (
	"errors"
	"pr-review-service/internal/models"
)

var (
	ErrTeamExists   = errors.New("team already exists")
	ErrTeamNotFound = errors.New("team not found")
)

func (s *Service) CreateTeam(team *models.Team) error {
	exists, err := s.db.TeamExists(team.TeamName)
	if err != nil {
		return err
	}
	if exists {
		return ErrTeamExists
	}

	return s.db.CreateTeam(team)
}

func (s *Service) GetTeam(teamName string) (*models.Team, error) {
	team, err := s.db.GetTeam(teamName)
	if err != nil {
		return nil, ErrTeamNotFound
	}
	return team, nil
}
