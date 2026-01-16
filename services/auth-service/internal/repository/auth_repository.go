package repository

import (
	"auth-service/internal/models"
	"auth-service/pkg/db/postgres"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// Tx — интерфейс для работы с транзакциями
type Tx interface {
	Exec(ctx context.Context, query string, args ...interface{}) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// UserRepository — интерфейс для работы с пользователями
type UserRepository interface {
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error

	// МЕТОДЫ ДЛЯ ТРАНЗАКЦИЙ
	BeginTx(ctx context.Context) (Tx, error)
	CreateTx(ctx context.Context, tx Tx, user *models.User) error
}

// userRepository — реализация
type userRepository struct {
	db     *postgres.DB
	logger *zap.SugaredLogger
}

// NewUserRepository — конструктор
func NewUserRepository(db *postgres.DB, logger *zap.SugaredLogger) UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// pgxTx — обёртка над pgx.Tx, чтобы соответствовать интерфейсу Tx
type pgxTx struct {
	tx pgx.Tx
}

func (t *pgxTx) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := t.tx.Exec(ctx, query, args...)
	return err
}

func (t *pgxTx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *pgxTx) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

// BeginTx — начало транзакции
func (r *userRepository) BeginTx(ctx context.Context) (Tx, error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		r.logger.Errorw("Failed to begin transaction", "error", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &pgxTx{tx: tx}, nil
}

// CreateTx — создание пользователя в транзакции
func (r *userRepository) CreateTx(ctx context.Context, tx Tx, user *models.User) error {
	pgxTx, ok := tx.(*pgxTx)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}

	query := `
		INSERT INTO auth.users (
			id, email, password_hash, role, status, is_verified, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	log := r.logger.With("user_id", user.ID, "email", user.Email)
	if err := pgxTx.Exec(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Role, user.Status, user.IsVerified,
		user.CreatedAt, user.UpdatedAt,
	); err != nil {
		log.Errorw("Failed to create user in transaction", "error", err)
		return fmt.Errorf("failed to insert user in transaction: %w", err)
	}

	log.Infow("User created in transaction")
	return nil
}

// Create — создание пользователя
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO auth.users (
			id, email, password_hash, role, status, is_verified, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	log := r.logger.With("user_id", user.ID, "email", user.Email)
	if err := r.db.Exec(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.Role, user.Status, user.IsVerified,
		user.CreatedAt, user.UpdatedAt,
	); err != nil {
		log.Errorw("Failed to create user", "error", err)
		return fmt.Errorf("failed to insert user: %w", err)
	}

	log.Infow("User created in DB")
	return nil
}

// GetByID — по ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, status, is_verified, last_login_at, created_at, updated_at
		FROM auth.users WHERE id = $1
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Status,
		&user.IsVerified, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Infow("User not found by ID", "user_id", id)
			return nil, errors.New("user not found")
		}
		r.logger.Errorw("DB error on GetByID", "user_id", id, "error", err)
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return user, nil
}

// GetByEmail — по email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, status, is_verified, last_login_at, created_at, updated_at
		FROM auth.users WHERE email = $1
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Status,
		&user.IsVerified, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Infow("User not found by email", "email", email)
			return nil, errors.New("user not found")
		}
		r.logger.Errorw("DB error on GetByEmail", "email", email, "error", err)
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return user, nil
}

// ExistsByEmail — проверка существования
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM auth.users WHERE email = $1)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, email).Scan(&exists); err != nil {
		r.logger.Errorw("Failed to check email existence", "email", email, "error", err)
		return false, fmt.Errorf("query failed: %w", err)
	}
	return exists, nil
}

// UpdateLastLogin — обновление времени входа
func (r *userRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE auth.users SET last_login_at = NOW(), updated_at = NOW() WHERE id = $1`
	if err := r.db.Exec(ctx, query, id); err != nil {
		r.logger.Warnw("Failed to update last_login_at", "user_id", id, "error", err)
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}
