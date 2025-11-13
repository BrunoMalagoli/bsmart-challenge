package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/db"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
	"github.com/gin-gonic/gin"
)

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

	models.RespondSuccess(c, http.StatusCreated, product)
}

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

	models.RespondSuccess(c, http.StatusOK, product)
}

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

	models.RespondSuccess(c, http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

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
