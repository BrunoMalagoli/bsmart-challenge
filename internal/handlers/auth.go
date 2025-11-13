package handlers

import (
	"net/http"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/auth"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
	"github.com/gin-gonic/gin"
)

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	var req models.UserRegisterRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		models.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	// Check if email already exists
	exists, err := h.DB.EmailExists(c.Request.Context(), req.Email)
	if err != nil {
		models.RespondError(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to check email existence")
		return
	}

	if exists {
		models.RespondError(c, http.StatusConflict, "EMAIL_EXISTS", "Email already registered")
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		models.RespondError(c, http.StatusInternalServerError, "HASH_ERROR", "Failed to hash password")
		return
	}

	// Get default "client" role
	role, err := h.DB.GetRoleByName(c.Request.Context(), "client")
	if err != nil {
		models.RespondError(c, http.StatusInternalServerError, "ROLE_ERROR", "Failed to get default role")
		return
	}

	// Create user
	user, err := h.DB.CreateUser(c.Request.Context(), req.Email, passwordHash, &role.ID)
	if err != nil {
		models.RespondError(c, http.StatusInternalServerError, "CREATE_ERROR", "Failed to create user")
		return
	}

	// Load user with role
	user, err = h.DB.GetUserByID(c.Request.Context(), user.ID)
	if err != nil {
		models.RespondError(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to load user")
		return
	}

	// Generate JWT token
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	token, err := h.JWTService.GenerateToken(user.ID, user.Email, roleName)
	if err != nil {
		models.RespondError(c, http.StatusInternalServerError, "TOKEN_ERROR", "Failed to generate token")
		return
	}

	// Return response
	response := models.UserLoginResponse{
		User:  user.ToUserResponse(),
		Token: token,
	}

	models.RespondSuccess(c, http.StatusCreated, response)
}

// Login handles user login
func (h *Handler) Login(c *gin.Context) {
	var req models.UserLoginRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		models.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	// Get user by email
	user, err := h.DB.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		models.RespondError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		return
	}

	// Check password
	if err := auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		models.RespondError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		return
	}

	// Generate JWT token
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	token, err := h.JWTService.GenerateToken(user.ID, user.Email, roleName)
	if err != nil {
		models.RespondError(c, http.StatusInternalServerError, "TOKEN_ERROR", "Failed to generate token")
		return
	}

	// Return response
	response := models.UserLoginResponse{
		User:  user.ToUserResponse(),
		Token: token,
	}

	models.RespondSuccess(c, http.StatusOK, response)
}
