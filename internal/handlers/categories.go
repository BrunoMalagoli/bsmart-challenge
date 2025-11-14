package handlers

import (
	"net/http"
	"strconv"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/websockets"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListCategories(c *gin.Context) {
	// Get optional search parameter
	search := c.Query("search")

	categories, err := h.DB.ListCategories(c.Request.Context(), search)
	if err != nil {
		models.RespondError(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to list categories")
		return
	}

	response := models.CategoryListResponse{
		Categories: categories,
		Total:      len(categories),
	}

	models.RespondSuccess(c, http.StatusOK, response)
}

func (h *Handler) GetCategory(c *gin.Context) {
	// Parse category ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		models.RespondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid category ID")
		return
	}

	category, err := h.DB.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		models.RespondError(c, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}

	models.RespondSuccess(c, http.StatusOK, category)
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var req models.CategoryCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		models.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	category, err := h.DB.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		models.RespondError(c, http.StatusInternalServerError, "CREATE_ERROR", "Failed to create category")
		return
	}

	// Broadcast WebSocket event
	websockets.BroadcastEvent(h.Hub, websockets.EventCategoryCreated, category)

	models.RespondSuccess(c, http.StatusCreated, category)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	// Parse category ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		models.RespondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid category ID")
		return
	}

	var req models.CategoryUpdateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		models.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	category, err := h.DB.UpdateCategory(c.Request.Context(), id, &req)
	if err != nil {
		if err.Error() == "category not found" {
			models.RespondError(c, http.StatusNotFound, "NOT_FOUND", "Category not found")
			return
		}
		models.RespondError(c, http.StatusInternalServerError, "UPDATE_ERROR", "Failed to update category")
		return
	}

	// Broadcast WebSocket event
	websockets.BroadcastEvent(h.Hub, websockets.EventCategoryUpdated, category)

	models.RespondSuccess(c, http.StatusOK, category)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	// Parse category ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		models.RespondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid category ID")
		return
	}

	if err := h.DB.DeleteCategory(c.Request.Context(), id); err != nil {
		if err.Error() == "category not found" {
			models.RespondError(c, http.StatusNotFound, "NOT_FOUND", "Category not found")
			return
		}
		models.RespondError(c, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete category")
		return
	}

	// Broadcast WebSocket event
	websockets.BroadcastEvent(h.Hub, websockets.EventCategoryDeleted, gin.H{"id": id})

	models.RespondSuccess(c, http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
