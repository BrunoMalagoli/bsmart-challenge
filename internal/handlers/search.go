package handlers

import (
	"net/http"
	"strconv"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/db"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
	"github.com/gin-gonic/gin"
)

// Search godoc
// @Summary      Búsqueda universal
// @Description  Busca productos o categorías según el parámetro 'type'. Soporta paginación, ordenamiento y filtros. Para productos, permite filtrar por category_id.
// @Tags         búsqueda
// @Accept       json
// @Produce      json
// @Param        type         query     string  false  "Tipo de búsqueda"  default(product)  Enums(product, category)
// @Param        q            query     string  false  "Término de búsqueda"
// @Param        page         query     int     false  "Número de página"  default(1)
// @Param        limit        query     int     false  "Items por página"  default(20)
// @Param        sort_by      query     string  false  "Campo para ordenar (name, price, stock, created_at para products)"  default(name)
// @Param        sort_order   query     string  false  "Dirección de ordenamiento"  default(asc)  Enums(asc, desc)
// @Param        category_id  query     int     false  "Filtrar productos por ID de categoría (solo para type=product)"
// @Success      200  {object}  models.ApiResponse{data=models.ProductListResponse}  "Resultados de búsqueda (ProductListResponse o CategoryListResponse según type)"
// @Failure      400  {object}  models.ApiResponse{error=models.ApiError}  "Parámetros inválidos"
// @Failure      401  {object}  models.ApiResponse{error=models.ApiError}  "No autenticado"
// @Failure      500  {object}  models.ApiResponse{error=models.ApiError}  "Error interno del servidor"
// @Security     BearerAuth
// @Router       /search [get]
func (h *Handler) Search(c *gin.Context) {
	searchType := c.Query("type")
	if searchType == "" {
		searchType = "product" // Default to product search
	}

	query := c.Query("q") // Optional search query

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
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	pagination := &db.PaginationParams{
		Page:  page,
		Limit: limit,
	}

	// Parse filter parameters
	filter := &db.FilterParams{
		Search:    query,
		SortBy:    c.DefaultQuery("sort_by", "name"),
		SortOrder: c.DefaultQuery("sort_order", "asc"),
	}

	// Parse category_id filter (optional)
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			models.RespondError(c, http.StatusBadRequest, "INVALID_CATEGORY_ID", "Invalid category_id parameter")
			return
		}
		filter.CategoryID = &categoryID
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
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	pagination := &db.PaginationParams{
		Page:  page,
		Limit: limit,
	}

	// Parse filter parameters
	filter := &db.FilterParams{
		SortBy:    c.DefaultQuery("sort_by", "name"),
		SortOrder: c.DefaultQuery("sort_order", "asc"),
	}

	categories, total, err := h.DB.SearchCategories(c.Request.Context(), query, pagination, filter)
	if err != nil {
		models.RespondError(c, http.StatusInternalServerError, "SEARCH_ERROR", "Failed to search categories")
		return
	}

	response := models.CategoryListResponse{
		Categories: categories,
		Total:      total,
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		TotalPages: db.CalculateTotalPages(total, pagination.Limit),
	}

	models.RespondSuccess(c, http.StatusOK, response)
}
