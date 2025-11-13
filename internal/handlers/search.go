package handlers

import (
	"net/http"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/db"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Search(c *gin.Context) {
	// Get search type (product or category)
	searchType := c.Query("type")
	if searchType == "" {
		searchType = "product" // Default to product search
	}

	query := c.Query("q")
	if query == "" {
		models.RespondError(c, http.StatusBadRequest, "MISSING_QUERY", "Search query parameter 'q' is required")
		return
	}

	switch searchType {
	case "product":
		h.searchProducts(c, query)
	case "category":
		h.searchCategories(c, query)
	default:
		models.RespondError(c, http.StatusBadRequest, "INVALID_TYPE", "Search type must be 'product' or 'category'")
	}
}

func (h *Handler) searchProducts(c *gin.Context, query string) {
	pagination := &db.PaginationParams{
		Page:  1,
		Limit: 20, // Default limit for search
	}

	filter := &db.FilterParams{
		Search:    query,
		SortBy:    "name",
		SortOrder: "asc",
	}

	products, total, err := h.DB.ListProducts(c.Request.Context(), pagination, filter)
	if err != nil {
		models.RespondError(c, http.StatusInternalServerError, "SEARCH_ERROR", "Failed to search products")
		return
	}

	response := models.ProductListResponse{
		Products:   products,
		Total:      total,
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		TotalPages: db.CalculateTotalPages(total, pagination.Limit),
	}

	models.RespondSuccess(c, http.StatusOK, response)
}

func (h *Handler) searchCategories(c *gin.Context, query string) {
	categories, err := h.DB.SearchCategories(c.Request.Context(), query)
	if err != nil {
		models.RespondError(c, http.StatusInternalServerError, "SEARCH_ERROR", "Failed to search categories")
		return
	}

	response := models.CategoryListResponse{
		Categories: categories,
		Total:      len(categories),
	}

	models.RespondSuccess(c, http.StatusOK, response)
}
