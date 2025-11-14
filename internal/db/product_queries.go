package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
	"github.com/jackc/pgx/v5"
)

// CreateProduct creates a new product with categories
func (db *DB) CreateProduct(ctx context.Context, req *models.ProductCreateRequest) (*models.Product, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO products (name, description, price, stock, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, name, description, price, stock, created_at, updated_at
	`

	var product models.Product
	err = tx.QueryRow(ctx, query, req.Name, req.Description, req.Price, req.Stock).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Insert product-category relationships
	if len(req.CategoryIDs) > 0 {
		if err := db.addProductCategories(ctx, tx, product.ID, req.CategoryIDs); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Load categories
	categories, _ := db.GetProductCategories(ctx, product.ID)
	product.Categories = categories

	return &product, nil
}

func (db *DB) GetProductByID(ctx context.Context, id int) (*models.Product, error) {
	query := `
		SELECT id, name, description, price, stock, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product models.Product
	err := db.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Load categories
	categories, _ := db.GetProductCategories(ctx, id)
	product.Categories = categories

	return &product, nil
}

// ListProducts retrieves a paginated list of products with filtering and sorting
func (db *DB) ListProducts(ctx context.Context, pagination *PaginationParams, filter *FilterParams) ([]models.Product, int, error) {
	pagination.Validate()
	filter.Validate()

	// Build WHERE clause for search and category filter
	var whereClause string
	var fromClause string
	var args []interface{}
	argCount := 0

	if filter.CategoryID != nil {
		fromClause = "FROM products p INNER JOIN product_category pc ON p.id = pc.product_id"
		argCount++
		whereClause = fmt.Sprintf("WHERE pc.category_id = $%d", argCount)
		args = append(args, *filter.CategoryID)
	} else {
		fromClause = "FROM products p"
	}

	if filter.Search != "" {
		argCount++
		if whereClause != "" {
			whereClause += fmt.Sprintf(" AND to_tsvector('simple', p.name) @@ plainto_tsquery('simple', $%d)", argCount)
		} else {
			whereClause = fmt.Sprintf("WHERE to_tsvector('simple', p.name) @@ plainto_tsquery('simple', $%d)", argCount)
		}
		args = append(args, filter.Search)
	}

	// Build ORDER BY clause
	allowedFields := map[string]string{
		"name":       "p.name",
		"price":      "p.price",
		"stock":      "p.stock",
		"created_at": "p.created_at",
	}
	orderByClause := filter.BuildOrderByClause(allowedFields)
	if orderByClause == "" {
		orderByClause = "ORDER BY p.created_at DESC" // Default sorting
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(DISTINCT p.id) %s %s", fromClause, whereClause)
	total, err := db.CountRows(ctx, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Query products with pagination
	argCount++
	limitArg := argCount
	argCount++
	offsetArg := argCount

	query := fmt.Sprintf(`
		SELECT DISTINCT p.id, p.name, p.description, p.price, p.stock, p.created_at, p.updated_at
		%s
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, fromClause, whereClause, orderByClause, limitArg, offsetArg)

	args = append(args, pagination.Limit, pagination.Offset())

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query products: %w", err)
	}

	products, err := ScanRows(rows, func(row pgx.Row) (models.Product, error) {
		var product models.Product
		err := row.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		return product, err
	})

	if err != nil {
		return nil, 0, fmt.Errorf("failed to scan products: %w", err)
	}

	// Load categories for each product
	for i := range products {
		categories, _ := db.GetProductCategories(ctx, products[i].ID)
		products[i].Categories = categories
	}

	return products, total, nil
}

func (db *DB) UpdateProduct(ctx context.Context, id int, req *models.ProductUpdateRequest) (*models.Product, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

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

	if req.Price != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("price = $%d", argCount))
		args = append(args, *req.Price)
	}

	if req.Stock != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("stock = $%d", argCount))
		args = append(args, *req.Stock)
	}

	if len(updates) == 0 && len(req.CategoryIDs) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// Always update updated_at
	argCount++
	updates = append(updates, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, time.Now())

	// Add product ID as last argument
	argCount++
	args = append(args, id)

	if len(updates) > 0 {
		query := fmt.Sprintf(`
			UPDATE products
			SET %s
			WHERE id = $%d
			RETURNING id, name, description, price, stock, created_at, updated_at
		`, strings.Join(updates, ", "), argCount)

		var product models.Product
		err = tx.QueryRow(ctx, query, args...).Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.CreatedAt,
			&product.UpdatedAt,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, fmt.Errorf("product not found")
			}
			return nil, fmt.Errorf("failed to update product: %w", err)
		}
	}

	// Update categories if provided
	if len(req.CategoryIDs) > 0 {
		// Delete existing categories
		_, err = tx.Exec(ctx, "DELETE FROM product_category WHERE product_id = $1", id)
		if err != nil {
			return nil, fmt.Errorf("failed to delete product categories: %w", err)
		}

		// Add new categories
		if err := db.addProductCategories(ctx, tx, id, req.CategoryIDs); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return updated product
	return db.GetProductByID(ctx, id)
}

func (db *DB) DeleteProduct(ctx context.Context, id int) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

func (db *DB) GetProductHistory(ctx context.Context, productID int, start, end *time.Time) ([]models.ProductHistory, error) {
	query := `
		SELECT id, product_id, price, stock, changed_at
		FROM product_history
		WHERE product_id = $1
	`

	args := []interface{}{productID}
	argCount := 1

	if start != nil {
		argCount++
		query += fmt.Sprintf(" AND changed_at >= $%d", argCount)
		args = append(args, *start)
	}

	if end != nil {
		argCount++
		query += fmt.Sprintf(" AND changed_at <= $%d", argCount)
		args = append(args, *end)
	}

	query += " ORDER BY changed_at DESC"

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query product history: %w", err)
	}

	history, err := ScanRows(rows, func(row pgx.Row) (models.ProductHistory, error) {
		var h models.ProductHistory
		err := row.Scan(&h.ID, &h.ProductID, &h.Price, &h.Stock, &h.ChangedAt)
		return h, err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan product history: %w", err)
	}

	return history, nil
}

func (db *DB) GetProductCategories(ctx context.Context, productID int) ([]models.Category, error) {
	query := `
		SELECT c.id, c.name, c.description, c.created_at, c.updated_at
		FROM categories c
		INNER JOIN product_category pc ON c.id = pc.category_id
		WHERE pc.product_id = $1
		ORDER BY c.name
	`

	rows, err := db.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to query product categories: %w", err)
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

// addProductCategories is a helper to add product-category relationships
func (db *DB) addProductCategories(ctx context.Context, tx pgx.Tx, productID int, categoryIDs []int) error {
	for _, categoryID := range categoryIDs {
		query := `INSERT INTO product_category (product_id, category_id) VALUES ($1, $2)`
		_, err := tx.Exec(ctx, query, productID, categoryID)
		if err != nil {
			return fmt.Errorf("failed to add product category: %w", err)
		}
	}
	return nil
}
