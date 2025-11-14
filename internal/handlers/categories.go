package handlers

import (
	"net/http"
	"strconv"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/websockets"
	"github.com/gin-gonic/gin"
)

// ListCategories godoc
// @Summary      Listar categorías
// @Description  Obtiene la lista completa de categorías con búsqueda opcional por nombre
// @Tags         categorías
// @Accept       json
// @Produce      json
// @Param        search  query     string  false  "Término de búsqueda en nombre de categoría"
// @Success      200  {object}  models.ApiResponse{data=models.CategoryListResponse}  "Lista de categorías"
// @Failure      401  {object}  models.ApiResponse{error=models.ApiError}  "No autenticado"
// @Failure      500  {object}  models.ApiResponse{error=models.ApiError}  "Error interno del servidor"
// @Security     BearerAuth
// @Router       /categories [get]
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

// GetCategory godoc
// @Summary      Obtener detalle de categoría
// @Description  Obtiene la información completa de una categoría específica
// @Tags         categorías
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID de la categoría"
// @Success      200  {object}  models.ApiResponse{data=models.Category}  "Detalle de la categoría"
// @Failure      400  {object}  models.ApiResponse{error=models.ApiError}  "ID inválido"
// @Failure      401  {object}  models.ApiResponse{error=models.ApiError}  "No autenticado"
// @Failure      404  {object}  models.ApiResponse{error=models.ApiError}  "Categoría no encontrada"
// @Failure      500  {object}  models.ApiResponse{error=models.ApiError}  "Error interno del servidor"
// @Security     BearerAuth
// @Router       /categories/{id} [get]
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

// CreateCategory godoc
// @Summary      Crear nueva categoría
// @Description  Crea una nueva categoría en el sistema. Requiere rol de administrador. Emite evento WebSocket 'category:created'.
// @Tags         categorías
// @Accept       json
// @Produce      json
// @Param        request  body      models.CategoryCreateRequest  true  "Datos de la categoría"
// @Success      201  {object}  models.ApiResponse{data=models.Category}  "Categoría creada exitosamente"
// @Failure      400  {object}  models.ApiResponse{error=models.ApiError}  "Datos de entrada inválidos"
// @Failure      401  {object}  models.ApiResponse{error=models.ApiError}  "No autenticado"
// @Failure      403  {object}  models.ApiResponse{error=models.ApiError}  "Requiere rol de administrador"
// @Failure      500  {object}  models.ApiResponse{error=models.ApiError}  "Error interno del servidor"
// @Security     BearerAuth
// @Router       /categories [post]
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

// UpdateCategory godoc
// @Summary      Actualizar categoría
// @Description  Actualiza una categoría existente. Requiere rol de administrador. Los campos no enviados no se modifican. Emite evento WebSocket 'category:updated'.
// @Tags         categorías
// @Accept       json
// @Produce      json
// @Param        id       path      int                           true  "ID de la categoría"
// @Param        request  body      models.CategoryUpdateRequest  true  "Datos a actualizar (campos opcionales)"
// @Success      200  {object}  models.ApiResponse{data=models.Category}  "Categoría actualizada exitosamente"
// @Failure      400  {object}  models.ApiResponse{error=models.ApiError}  "ID inválido o datos de entrada inválidos"
// @Failure      401  {object}  models.ApiResponse{error=models.ApiError}  "No autenticado"
// @Failure      403  {object}  models.ApiResponse{error=models.ApiError}  "Requiere rol de administrador"
// @Failure      404  {object}  models.ApiResponse{error=models.ApiError}  "Categoría no encontrada"
// @Failure      500  {object}  models.ApiResponse{error=models.ApiError}  "Error interno del servidor"
// @Security     BearerAuth
// @Router       /categories/{id} [put]
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

// DeleteCategory godoc
// @Summary      Eliminar categoría
// @Description  Elimina una categoría del sistema. Requiere rol de administrador. Emite evento WebSocket 'category:deleted'.
// @Tags         categorías
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID de la categoría"
// @Success      200  {object}  models.ApiResponse{data=object}  "Categoría eliminada exitosamente"
// @Failure      400  {object}  models.ApiResponse{error=models.ApiError}  "ID inválido"
// @Failure      401  {object}  models.ApiResponse{error=models.ApiError}  "No autenticado"
// @Failure      403  {object}  models.ApiResponse{error=models.ApiError}  "Requiere rol de administrador"
// @Failure      404  {object}  models.ApiResponse{error=models.ApiError}  "Categoría no encontrada"
// @Failure      500  {object}  models.ApiResponse{error=models.ApiError}  "Error interno del servidor"
// @Security     BearerAuth
// @Router       /categories/{id} [delete]
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
