package models

import "time"

type Category struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name" binding:"required"`
	Description *string   `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CategoryCreateRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

type CategoryUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// CategoryListResponse represents a list of categories
type CategoryListResponse struct {
	Categories []Category `json:"categories"`
	Total      int        `json:"total"`
}
