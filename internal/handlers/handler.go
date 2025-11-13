package handlers

import (
	"github.com/BrunoMalagoli/bsmart-challenge/internal/auth"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/db"
)

// Handler holds dependencies for all handlers
type Handler struct {
	DB         *db.DB
	JWTService *auth.JWTService
}

func NewHandler(database *db.DB, jwtService *auth.JWTService) *Handler {
	return &Handler{
		DB:         database,
		JWTService: jwtService,
	}
}
