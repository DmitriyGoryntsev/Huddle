package models

import "time"

type Category struct {
	ID        int       `json:"id" db:"id"`
	ParentID  *int      `json:"parent_id,omitempty" db:"parent_id"` // null, если это главная категория
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"` // Человекочитаемый ID (например, 'sports')
	IconURL   string    `json:"icon_url" db:"icon_url"`
	ColorCode string    `json:"color_code" db:"color_code"` // Для выделения на карте
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Поле для вложенности (опционально, если будешь отдавать дерево)
	Subcategories []Category `json:"subcategories,omitempty"`
}
