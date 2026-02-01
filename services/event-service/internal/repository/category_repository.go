package repository

import (
	"context"
	"event-service/internal/models"
	"event-service/pkg/db/postgres"

	"go.uber.org/zap"
)

type CategoryRepository interface {
	List(ctx context.Context) ([]models.Category, error)
}

type categoryRepository struct {
	db     *postgres.DB
	logger *zap.SugaredLogger
}

func NewCategoryRepository(db *postgres.DB, logger *zap.SugaredLogger) CategoryRepository {
	return &categoryRepository{db: db, logger: logger}
}

func (r *categoryRepository) List(ctx context.Context) ([]models.Category, error) {
	query := `
		SELECT id, parent_id, name, slug, COALESCE(icon_url, ''), COALESCE(color_code, '')
		FROM categories
		ORDER BY COALESCE(parent_id, id), id
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.Errorw("Failed to list categories", "error", err)
		return nil, err
	}
	defer rows.Close()

	var list []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Name, &c.Slug, &c.IconURL, &c.ColorCode); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}
