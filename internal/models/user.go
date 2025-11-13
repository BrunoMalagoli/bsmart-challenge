package models

import "time"

type Role struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name" binding:"required"`
}

type User struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email" binding:"required,email"`
	PasswordHash string    `json:"-" db:"password_hash"` // Never expose password in JSON
	RoleID       *int      `json:"role_id,omitempty" db:"role_id"`
	Role         *Role     `json:"role,omitempty" db:"-"` // Joined role data
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type UserRegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserLoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

type UserResponse struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Role      *Role     `json:"role,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ToUserResponse converts a User to UserResponse (removes sensitive data)
func (u *User) ToUserResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}
