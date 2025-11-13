package db

import (
	"context"
	"fmt"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
	"github.com/jackc/pgx/v5"
)

func (db *DB) CreateUser(ctx context.Context, email, passwordHash string, roleID *int) (*models.User, error) {
	query := `
		INSERT INTO users (email, password_hash, role_id, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, email, password_hash, role_id, created_at
	`

	var user models.User
	err := db.QueryRow(ctx, query, email, passwordHash, roleID).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.RoleID,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

func (db *DB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.role_id, u.created_at,
		       r.id, r.name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1
	`

	var user models.User
	var role models.Role
	var roleID *int
	var roleName *string

	err := db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.RoleID,
		&user.CreatedAt,
		&roleID,
		&roleName,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Set role if it exists
	if roleID != nil && roleName != nil {
		role.ID = *roleID
		role.Name = *roleName
		user.Role = &role
	}

	return &user, nil
}

func (db *DB) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.role_id, u.created_at,
		       r.id, r.name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`

	var user models.User
	var role models.Role
	var roleID *int
	var roleName *string

	err := db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.RoleID,
		&user.CreatedAt,
		&roleID,
		&roleName,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Set role if it exists
	if roleID != nil && roleName != nil {
		role.ID = *roleID
		role.Name = *roleName
		user.Role = &role
	}

	return &user, nil
}

func (db *DB) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	query := `SELECT id, name FROM roles WHERE name = $1`

	var role models.Role
	err := db.QueryRow(ctx, query, name).Scan(&role.ID, &role.Name)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

func (db *DB) CreateRole(ctx context.Context, name string) (*models.Role, error) {
	query := `
		INSERT INTO roles (name)
		VALUES ($1)
		RETURNING id, name
	`

	var role models.Role
	err := db.QueryRow(ctx, query, name).Scan(&role.ID, &role.Name)

	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return &role, nil
}

func (db *DB) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := db.QueryRow(ctx, query, email).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}
