package repository

import (
	"context"
	"errors"
	"fmt"
	"profile-service/internal/models"
	"profile-service/pkg/db/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// ProfileRepository — интерфейс для работы с БД профилей
type ProfileRepository interface {
	Create(ctx context.Context, profile *models.Profile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
}

type profileRepository struct {
	db     *postgres.DB
	logger *zap.SugaredLogger
}

func NewProfileRepository(db *postgres.DB, logger *zap.SugaredLogger) ProfileRepository {
	return &profileRepository{
		db:     db,
		logger: logger,
	}
}

// Create — сохранение нового профиля
func (r *profileRepository) Create(ctx context.Context, profile *models.Profile) error {
	query := `
		INSERT INTO profile.profiles (user_id, first_name, last_name, avatar_url, bio, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (user_id) DO NOTHING
	`

	err := r.db.Exec(ctx, query,
		profile.UserID,
		profile.FirstName,
		profile.LastName,
		profile.AvatarURL,
		profile.Bio,
	)

	if err != nil {
		r.logger.Errorw("Failed to insert profile", "user_id", profile.UserID, "error", err)
		return fmt.Errorf("failed to create profile: %w", err)
	}

	return nil
}

// GetByUserID — получение профиля по ID пользователя
func (r *profileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	query := `
		SELECT user_id, first_name, last_name, avatar_url, bio, created_at, updated_at
		FROM profile.profiles
		WHERE user_id = $1
	`

	profile := &models.Profile{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&profile.UserID,
		&profile.FirstName,
		&profile.LastName,
		&profile.AvatarURL,
		&profile.Bio,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("profile not found")
		}
		r.logger.Errorw("Database error on GetByUserID", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to query profile: %w", err)
	}

	return profile, nil
}
