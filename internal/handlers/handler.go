package handlers

import (
	"github.com/BrunoMalagoli/bsmart-challenge/internal/auth"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/db"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/websockets"
)

// Handler holds dependencies for all handlers
type Handler struct {
	DB         *db.DB
	JWTService *auth.JWTService
	Hub        *websockets.Hub
}

func NewHandler(database *db.DB, jwtService *auth.JWTService, hub *websockets.Hub) *Handler {
	return &Handler{
		DB:         database,
		JWTService: jwtService,
		Hub:        hub,
	}
}
