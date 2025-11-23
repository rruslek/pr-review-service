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

func (s *Service) BulkDeactivateTeamUsers(teamName string) (*models.BulkDeactivateResponse, error) {
	deactivatedUserIDs, err := s.db.BulkDeactivateTeamUsers(teamName)
	if err != nil {
		return nil, err
	}

	if len(deactivatedUserIDs) == 0 {
		return &models.BulkDeactivateResponse{
			TeamName:         teamName,
			DeactivatedUsers: []string{},
			ReassignedPRs:    []string{},
			DeactivatedCount: 0,
			ReassignedCount:  0,
		}, nil
	}

	userTeams, err := s.db.GetTeamNameForUsers(deactivatedUserIDs)
	if err != nil {
		return nil, err
	}

	prReviewers, prAuthors, err := s.db.GetOpenPRsWithReviewers(deactivatedUserIDs)
	if err != nil {
		return nil, err
	}

	reassignments := make(map[string]map[string]string)

	for prID, reviewers := range prReviewers {
		authorID := prAuthors[prID]
		reassignments[prID] = make(map[string]string)

		for _, reviewerID := range reviewers {
			isDeactivated := false
			for _, deactivatedID := range deactivatedUserIDs {
				if reviewerID == deactivatedID {
					isDeactivated = true
					break
				}
			}

			if !isDeactivated {
				continue
			}

			reviewerTeam, ok := userTeams[reviewerID]
			if !ok {
				reviewerTeam = teamName
			}

			excludeList := append([]string{reviewerID, authorID}, deactivatedUserIDs...)
			candidates, err := s.db.GetActiveTeamMembersForReplacement(reviewerTeam, excludeList)
			if err != nil {
				return nil, err
			}

			if len(candidates) == 0 {
				authorTeam, err := s.getAuthorTeam(authorID)
				if err == nil {
					excludeList = append([]string{reviewerID, authorID}, deactivatedUserIDs...)
					candidates, err = s.db.GetActiveTeamMembersForReplacement(authorTeam, excludeList)
					if err != nil {
						return nil, err
					}
				}
			}

			if len(candidates) > 0 {
				newReviewerID := selectRandomReviewers(candidates, 1)[0]
				reassignments[prID][reviewerID] = newReviewerID
			}
		}
	}

	reassignedPRs, err := s.db.BulkReassignReviewers(reassignments)
	if err != nil {
		return nil, err
	}

	return &models.BulkDeactivateResponse{
		TeamName:         teamName,
		DeactivatedUsers: deactivatedUserIDs,
		ReassignedPRs:    reassignedPRs,
		DeactivatedCount: len(deactivatedUserIDs),
		ReassignedCount:  len(reassignedPRs),
	}, nil
}

func (s *Service) getAuthorTeam(authorID string) (string, error) {
	user, err := s.db.GetUser(authorID)
	if err != nil {
		return "", err
	}
	return user.TeamName, nil
}
