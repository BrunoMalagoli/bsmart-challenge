package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/db"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/websockets"
	"github.com/gin-gonic/gin"
)

// ListProducts godoc
// @Summary      Listar productos con paginación
// @Description  Obtiene una lista paginada de productos con opciones de filtrado, ordenamiento y búsqueda full-text
// @Tags         productos
// @Accept       json
// @Produce      json
// @Param        page        query     int     false  "Número de página"  default(1)
// @Param        limit       query     int     false  "Productos por página (máximo 100)"  default(10)
// @Param        sort_by     query     string  false  "Campo para ordenar"  default(created_at)  Enums(name, price, stock, created_at)
// @Param        sort_order  query     string  false  "Dirección de ordenamiento"  default(desc)  Enums(asc, desc)
// @Param        search      query     string  false  "Término de búsqueda full-text en nombres de productos"
// @Success      200  {object}  models.ApiResponse{data=models.ProductListResponse}  "Lista de productos"
// @Failure      401  {object}  models.ApiResponse{error=models.ApiError}  "No autenticado"
// @Failure      500  {object}  models.ApiResponse{error=models.ApiError}  "Error interno del servidor"
// @Security     BearerAuth
// @Router       /products [get]
func (h *Handler) ListProducts(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	pagination := &db.PaginationParams{
		Page:  page,
		Limit: limit,
	}

	// Parse filter parameters
	filter := &db.FilterParams{
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
		Search:    c.Query("search"),
	}

	// Get products from database
	products, total, err := h.DB.ListProducts(c.Request.Context(), pagination, filter)
	if err != nil {
		models.RespondError(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to list products")
		return
	}

	// Calculate total pages
	totalPages := db.CalculateTotalPages(total, pagination.Limit)

	// Build response
	response := models.ProductListResponse{
		Products:   products,
		Total:      total,
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		TotalPages: totalPages,
	}

	models.RespondSuccess(c, http.StatusOK, response)
}

// GetProduct godoc
// @Summary      Obtener detalle de producto
// @Description  Obtiene la información completa de un producto específico incluyendo sus categorías
// @Tags         productos
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID del producto"
// @Success      200  {object}  models.ApiResponse{data=models.Product}  "Detalle del producto"
// @Failure      400  {object}  models.ApiResponse{error=models.ApiError}  "ID inválido"
// @Failure      401  {object}  models.ApiResponse{error=models.ApiError}  "No autenticado"
// @Failure      404  {object}  models.ApiResponse{error=models.ApiError}  "Producto no encontrado"
// @Failure      500  {object}  models.ApiResponse{error=models.ApiError}  "Error interno del servidor"
// @Security     BearerAuth
// @Router       /products/{id} [get]
func (h *Handler) GetProduct(c *gin.Context) {
	// Parse product ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		models.RespondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	// Get product from database
	product, err := h.DB.GetProductByID(c.Request.Context(), id)
	if err != nil {
		models.RespondError(c, http.StatusNotFound, "NOT_FOUND", "Product not found")
		return
	}

	models.RespondSuccess(c, http.StatusOK, product)
}

// CreateProduct godoc
// @Summary      Crear nuevo producto
// @Description  Crea un nuevo producto en el sistema. Requiere rol de administrador. Emite evento WebSocket 'product:created'.
// @Tags         productos
// @Accept       json
// @Produce      json
// @Param        request  body      models.ProductCreateRequest  true  "Datos del producto"
// @Success      201  {object}  models.ApiResponse{data=models.Product}  "Producto creado exitosamente"
// @Failure      400  {object}  models.ApiResponse{error=models.ApiError}  "Datos de entrada inválidos"
// @Failure      401  {object}  models.ApiResponse{error=models.ApiError}  "No autenticado"
// @Failure      403  {object}  models.ApiResponse{error=models.ApiError}  "Requiere rol de administrador"
// @Failure      500  {object}  models.ApiResponse{error=models.ApiError}  "Error interno del servidor"
// @Security     BearerAuth
// @Router       /products [post]
func (h *Handler) CreateProduct(c *gin.Context) {
	var req models.ProductCreateRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		models.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	// Create product
	product, err := h.DB.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		models.RespondError(c, http.StatusInternalServerError, "CREATE_ERROR", "Failed to create product")
		return
	}

	// Broadcast WebSocket event
	websockets.BroadcastEvent(h.Hub, websockets.EventProductCreated, product)

	models.RespondSuccess(c, http.StatusCreated, product)
}

// UpdateProduct godoc
// @Summary      Actualizar producto
// @Description  Actualiza un producto existente. Requiere rol de administrador. Los campos no enviados no se modifican. Emite evento WebSocket 'product:updated'.
// @Tags         productos
// @Accept       json
// @Produce      json
// @Param        id       path      int                          true  "ID del producto"
// @Param        request  body      models.ProductUpdateRequest  true  "Datos a actualizar (campos opcionales)"
// @Success      200  {object}  models.ApiResponse{data=models.Product}  "Producto actualizado exitosamente"
// @Failure      400  {object}  models.ApiResponse{error=models.ApiError}  "ID inválido o datos de entrada inválidos"
// @Failure      401  {object}  models.ApiResponse{error=models.ApiError}  "No autenticado"
// @Failure      403  {object}  models.ApiResponse{error=models.ApiError}  "Requiere rol de administrador"
// @Failure      404  {object}  models.ApiResponse{error=models.ApiError}  "Producto no encontrado"
// @Failure      500  {object}  models.ApiResponse{error=models.ApiError}  "Error interno del servidor"
// @Security     BearerAuth
// @Router       /products/{id} [put]
func (h *Handler) UpdateProduct(c *gin.Context) {
	// Parse product ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		models.RespondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	var req models.ProductUpdateRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		models.RespondError(c, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	// Update product
	product, err := h.DB.UpdateProduct(c.Request.Context(), id, &req)
	if err != nil {
		if err.Error() == "product not found" {
			models.RespondError(c, http.StatusNotFound, "NOT_FOUND", "Product not found")
			return
		}
		models.RespondError(c, http.StatusInternalServerError, "UPDATE_ERROR", "Failed to update product")
		return
	}

	// Broadcast WebSocket event
	websockets.BroadcastEvent(h.Hub, websockets.EventProductUpdated, product)

	models.RespondSuccess(c, http.StatusOK, product)
}

// DeleteProduct godoc
// @Summary      Eliminar producto
// @Description  Elimina un producto del sistema. Requiere rol de administrador. Emite evento WebSocket 'product:deleted'.
// @Tags         productos
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID del producto"
// @Success      200  {object}  models.ApiResponse{data=object}  "Producto eliminado exitosamente"
// @Failure      400  {object}  models.ApiResponse{error=models.ApiError}  "ID inválido"
// @Failure      401  {object}  models.ApiResponse{error=models.ApiError}  "No autenticado"
// @Failure      403  {object}  models.ApiResponse{error=models.ApiError}  "Requiere rol de administrador"
// @Failure      404  {object}  models.ApiResponse{error=models.ApiError}  "Producto no encontrado"
// @Failure      500  {object}  models.ApiResponse{error=models.ApiError}  "Error interno del servidor"
// @Security     BearerAuth
// @Router       /products/{id} [delete]
func (h *Handler) DeleteProduct(c *gin.Context) {
	// Parse product ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		models.RespondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	// Delete product
	if err := h.DB.DeleteProduct(c.Request.Context(), id); err != nil {
		if err.Error() == "product not found" {
			models.RespondError(c, http.StatusNotFound, "NOT_FOUND", "Product not found")
			return
		}
		models.RespondError(c, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete product")
		return
	}

	// Broadcast WebSocket event
	websockets.BroadcastEvent(h.Hub, websockets.EventProductDeleted, gin.H{"id": id})

	models.RespondSuccess(c, http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// GetProductHistory godoc
// @Summary      Obtener historial de producto
// @Description  Obtiene el historial de cambios de precio y stock de un producto. Opcionalmente filtra por rango de fechas (formato RFC3339).
// @Tags         productos
// @Accept       json
// @Produce      json
// @Param        id     path      int     true   "ID del producto"
// @Param        start  query     string  false  "Fecha de inicio (RFC3339: 2024-01-01T00:00:00Z)"
// @Param        end    query     string  false  "Fecha de fin (RFC3339: 2024-12-31T23:59:59Z)"
// @Success      200  {object}  models.ApiResponse{data=models.ProductHistoryResponse}  "Historial del producto"
// @Failure      400  {object}  models.ApiResponse{error=models.ApiError}  "ID inválido o formato de fecha incorrecto"
// @Failure      401  {object}  models.ApiResponse{error=models.ApiError}  "No autenticado"
// @Failure      500  {object}  models.ApiResponse{error=models.ApiError}  "Error interno del servidor"
// @Security     BearerAuth
// @Router       /products/{id}/history [get]
func (h *Handler) GetProductHistory(c *gin.Context) {
	// Parse product ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		models.RespondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	// Parse date range parameters (optional)
	var startDate, endDate *time.Time

	if startStr := c.Query("start"); startStr != "" {
		start, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			models.RespondError(c, http.StatusBadRequest, "INVALID_DATE", "Invalid start date format (use RFC3339)")
			return
		}
		startDate = &start
	}

	if endStr := c.Query("end"); endStr != "" {
		end, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			models.RespondError(c, http.StatusBadRequest, "INVALID_DATE", "Invalid end date format (use RFC3339)")
			return
		}
		endDate = &end
	}

	// Get product history
	history, err := h.DB.GetProductHistory(c.Request.Context(), id, startDate, endDate)
	if err != nil {
		models.RespondError(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to get product history")
		return
	}

	response := models.ProductHistoryResponse{
		History: history,
		Total:   len(history),
	}

	models.RespondSuccess(c, http.StatusOK, response)
}
