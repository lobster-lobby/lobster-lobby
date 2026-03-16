package services

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/lobster-lobby/lobster-lobby/models"
	"github.com/lobster-lobby/lobster-lobby/repository"
)

type ReputationService struct {
	reputationRepo *repository.ReputationRepository
	userRepo       *repository.UserRepository
}

func NewReputationService(reputationRepo *repository.ReputationRepository, userRepo *repository.UserRepository) *ReputationService {
	return &ReputationService{reputationRepo: reputationRepo, userRepo: userRepo}
}

func (s *ReputationService) AwardPoints(ctx context.Context, userID bson.ObjectID, action, entityID, entityType string) error {
	points, ok := models.PointValues[action]
	if !ok {
		return fmt.Errorf("unknown reputation action: %s", action)
	}

	event := &models.ReputationEvent{
		UserID:     userID,
		Action:     action,
		Points:     points,
		EntityID:   entityID,
		EntityType: entityType,
	}

	if err := s.reputationRepo.LogEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to log reputation event: %w", err)
	}

	return s.RecalculateScore(ctx, userID)
}

func (s *ReputationService) GetReputation(ctx context.Context, userID bson.ObjectID) (*models.ReputationScore, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return &user.Reputation, nil
}

func (s *ReputationService) GetHistory(ctx context.Context, userID bson.ObjectID, page int) ([]models.ReputationEvent, int64, error) {
	return s.reputationRepo.ListByUser(ctx, userID, page, 20)
}

func (s *ReputationService) RecalculateScore(ctx context.Context, userID bson.ObjectID) error {
	score, contributions, err := s.reputationRepo.SumByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to sum reputation: %w", err)
	}

	rep := models.ReputationScore{
		Score:         score,
		Contributions: contributions,
		Tier:          models.TierForScore(score),
	}

	return s.userRepo.UpdateReputation(ctx, userID, rep)
}
