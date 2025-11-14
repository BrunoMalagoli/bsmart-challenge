package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
	"github.com/jackc/pgx/v5"
)

func (db *DB) CreateCategory(ctx context.Context, req *models.CategoryCreateRequest) (*models.Category, error) {
	query := `
		INSERT INTO categories (name, description, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, name, description, created_at, updated_at
	`

	var category models.Category
	err := db.QueryRow(ctx, query, req.Name, req.Description).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &category, nil
}

func (db *DB) GetCategoryByID(ctx context.Context, id int) (*models.Category, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	var category models.Category
	err := db.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

// ListCategories retrieves all categories
func (db *DB) ListCategories(ctx context.Context, search string) ([]models.Category, error) {
	var query string
	var args []interface{}

	if search != "" {
		query = `
			SELECT id, name, description, created_at, updated_at
			FROM categories
			WHERE name ILIKE $1
			ORDER BY name
		`
		args = append(args, "%"+search+"%")
	} else {
		query = `
			SELECT id, name, description, created_at, updated_at
			FROM categories
			ORDER BY name
		`
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}

	categories, err := ScanRows(rows, func(row pgx.Row) (models.Category, error) {
		var c models.Category
		err := row.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
		return c, err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan categories: %w", err)
	}

	return categories, nil
}

func (db *DB) UpdateCategory(ctx context.Context, id int, req *models.CategoryUpdateRequest) (*models.Category, error) {
	// Build dynamic UPDATE query
	updates := []string{}
	args := []interface{}{}
	argCount := 0

	if req.Name != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("name = $%d", argCount))
		args = append(args, *req.Name)
	}

	if req.Description != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("description = $%d", argCount))
		args = append(args, *req.Description)
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// Always update updated_at
	argCount++
	updates = append(updates, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, "NOW()")

	// Add category ID as last argument
	argCount++
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE categories
		SET %s
		WHERE id = $%d
		RETURNING id, name, description, created_at, updated_at
	`, strings.Join(updates, ", "), argCount)

	var category models.Category
	err := db.QueryRow(ctx, query, args...).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return &category, nil
}

func (db *DB) DeleteCategory(ctx context.Context, id int) error {
	query := `DELETE FROM categories WHERE id = $1`

	result, err := db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

// SearchCategories searches categories by name with pagination and sorting
func (db *DB) SearchCategories(ctx context.Context, searchTerm string, pagination *PaginationParams, filter *FilterParams) ([]models.Category, int, error) {
	pagination.Validate()
	filter.Validate()

	// Build WHERE clause
	var whereClause string
	var args []interface{}

	if searchTerm != "" {
		whereClause = "WHERE name ILIKE $1"
		args = append(args, "%"+searchTerm+"%")
	}

	// Build ORDER BY clause
	allowedFields := map[string]string{
		"name":       "name",
		"created_at": "created_at",
	}
	orderByClause := filter.BuildOrderByClause(allowedFields)
	if orderByClause == "" {
		orderByClause = "ORDER BY name ASC" // Default sorting
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM categories %s", whereClause)
	total, err := db.CountRows(ctx, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count categories: %w", err)
	}

	// Query categories with pagination
	argCount := len(args)
	argCount++
	limitArg := argCount
	argCount++
	offsetArg := argCount

	query := fmt.Sprintf(`
		SELECT id, name, description, created_at, updated_at
		FROM categories
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderByClause, limitArg, offsetArg)

	args = append(args, pagination.Limit, pagination.Offset())

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search categories: %w", err)
	}

	categories, err := ScanRows(rows, func(row pgx.Row) (models.Category, error) {
		var c models.Category
		err := row.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
		return c, err
	})

	if err != nil {
		return nil, 0, fmt.Errorf("failed to scan categories: %w", err)
	}

	return categories, total, nil
}
