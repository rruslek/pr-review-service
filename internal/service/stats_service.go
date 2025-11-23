package service

import "pr-review-service/internal/models"

func (s *Service) GetStats() (*models.StatsResponse, error) {
	usersStats, err := s.db.GetUserStats()
	if err != nil {
		return nil, err
	}

	prsStats, err := s.db.GetPRStats()
	if err != nil {
		return nil, err
	}

	totalUsers, err := s.db.GetTotalUsersCount()
	if err != nil {
		return nil, err
	}

	totalPRs, err := s.db.GetTotalPRsCount()
	if err != nil {
		return nil, err
	}

	return &models.StatsResponse{
		UsersStats: usersStats,
		PRsStats:   prsStats,
		TotalUsers: totalUsers,
		TotalPRs:   totalPRs,
	}, nil
}
