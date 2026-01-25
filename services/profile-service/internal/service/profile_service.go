package service

import (
	"context"
	"fmt"
	"profile-service/internal/models"
	"profile-service/internal/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ProfileService — бизнес-логика
type ProfileService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
	CreateProfile(ctx context.Context, profileData models.UserRegistered) error
}

type profileService struct {
	profileRepo repository.ProfileRepository
	logger      *zap.SugaredLogger
}

func NewProfileService(
	profileRepo repository.ProfileRepository,
	logger *zap.SugaredLogger,
) ProfileService {
	return &profileService{
		profileRepo: profileRepo,
		logger:      logger,
	}
}

// GetProfile — просто получение данных
func (s *profileService) GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	profile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Errorw("Failed to get profile", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	return profile, nil
}

// CreateProfile — создание начального профиля
func (s *profileService) CreateProfile(ctx context.Context, profileData models.UserRegistered) error {
	// TODO PROFILE ALREADY EXIST

	profile := &models.Profile{
		UserID:    profileData.UserID,
		FirstName: profileData.FirstName,
		LastName:  profileData.LastName,
		AvatarURL: "",
		Bio:       "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.profileRepo.Create(ctx, profile); err != nil {
		s.logger.Errorw("Failed to create profile from event",
			"user_id", profileData.UserID,
			"error", err,
		)
		return fmt.Errorf("failed to create profile: %w", err)
	}

	s.logger.Infow("Profile created successfully", "user_id", profileData.UserID)
	return nil
}
