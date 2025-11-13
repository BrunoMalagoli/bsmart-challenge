package models

import "github.com/gin-gonic/gin"

// ApiResponse is the standard response format for all API endpoints
type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ApiError   `json:"error,omitempty"`
}

// ApiError represents an error in the API response
type ApiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func SuccessResponse(data interface{}) ApiResponse {
	return ApiResponse{
		Success: true,
		Data:    data,
	}
}

func ErrorResponse(code, message string) ApiResponse {
	return ApiResponse{
		Success: false,
		Error: &ApiError{
			Code:    code,
			Message: message,
		},
	}
}

// ErrorResponseWithDetails creates an error API response with additional details
func ErrorResponseWithDetails(code, message, details string) ApiResponse {
	return ApiResponse{
		Success: false,
		Error: &ApiError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// RespondJSON sends a JSON response with the given status code and data
func RespondJSON(c *gin.Context, statusCode int, response ApiResponse) {
	c.JSON(statusCode, response)
}

func RespondSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, SuccessResponse(data))
}

func RespondError(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, ErrorResponse(code, message))
}

func RespondErrorWithDetails(c *gin.Context, statusCode int, code, message, details string) {
	c.JSON(statusCode, ErrorResponseWithDetails(code, message, details))
}
