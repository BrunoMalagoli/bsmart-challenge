package models

import "time"

type Product struct {
	ID          int        `json:"id" db:"id"`
	Name        string     `json:"name" db:"name" binding:"required"`
	Description *string    `json:"description,omitempty" db:"description"`
	Price       float64    `json:"price" db:"price" binding:"required,gte=0"`
	Stock       int        `json:"stock" db:"stock" binding:"required,gte=0"`
	Categories  []Category `json:"categories,omitempty" db:"-"` // Joined category data
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// ProductHistory represents a historical record of product changes
type ProductHistory struct {
	ID        int       `json:"id" db:"id"`
	ProductID int       `json:"product_id" db:"product_id"`
	Price     float64   `json:"price" db:"price"`
	Stock     int       `json:"stock" db:"stock"`
	ChangedAt time.Time `json:"changed_at" db:"changed_at"`
}

// ProductCreateRequest represents the request to create a product
type ProductCreateRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Price       float64 `json:"price" binding:"required,gte=0"`
	Stock       int     `json:"stock" binding:"required,gte=0"`
	CategoryIDs []int   `json:"category_ids" binding:"required,min=1"` // At least one category
}

type ProductUpdateRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price" binding:"omitempty,gte=0"`
	Stock       *int     `json:"stock" binding:"omitempty,gte=0"`
	CategoryIDs []int    `json:"category_ids" binding:"omitempty,min=1"`
}

type ProductListResponse struct {
	Products   []Product `json:"products"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}

// ProductHistoryResponse represents a list of product history records
type ProductHistoryResponse struct {
	History []ProductHistory `json:"history"`
	Total   int              `json:"total"`
}
