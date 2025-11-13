package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps the pgxpool.Pool to provide helper methods
type DB struct {
	*pgxpool.Pool
}

func NewDB(pool *pgxpool.Pool) *DB {
	return &DB{Pool: pool}
}

type PaginationParams struct {
	Page  int
	Limit int
}

func (p *PaginationParams) Offset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.Limit
}

// Validate validates and sets defaults for pagination params
func (p *PaginationParams) Validate() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 100 // Max limit
	}
}

type FilterParams struct {
	SortBy    string // Field to sort by
	SortOrder string // "asc" or "desc"
	Search    string // Search term
}

func (f *FilterParams) Validate() {
	if f.SortOrder != "asc" && f.SortOrder != "desc" {
		f.SortOrder = "asc"
	}
}

// BuildOrderByClause builds an ORDER BY SQL clause
func (f *FilterParams) BuildOrderByClause(allowedFields map[string]string) string {
	if f.SortBy == "" {
		return ""
	}

	// Check if the field is allowed
	dbField, ok := allowedFields[f.SortBy]
	if !ok {
		return ""
	}

	order := "ASC"
	if strings.ToLower(f.SortOrder) == "desc" {
		order = "DESC"
	}

	return fmt.Sprintf("ORDER BY %s %s", dbField, order)
}

func ScanRow[T any](row pgx.Row, dest *T, scanFunc func(pgx.Row, *T) error) error {
	return scanFunc(row, dest)
}

func ScanRows[T any](rows pgx.Rows, scanFunc func(pgx.Row) (T, error)) ([]T, error) {
	defer rows.Close()

	var results []T
	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (db *DB) CountRows(ctx context.Context, query string, args ...interface{}) (int, error) {
	var count int
	err := db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

func CalculateTotalPages(total, limit int) int {
	if limit == 0 {
		return 0
	}
	pages := total / limit
	if total%limit > 0 {
		pages++
	}
	return pages
}
