package middleware

import (
	"fmt"
	"net/http"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
	"github.com/gin-gonic/gin"
)

// ErrorHandler is a middleware that catches panics and formats errors consistently
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {

				fmt.Printf("Panic recovered: %v\n", err)

				// Return a generic error response
				models.RespondError(
					c,
					http.StatusInternalServerError,
					"INTERNAL_ERROR",
					"An unexpected error occurred",
				)

				c.Abort()
			}
		}()

		c.Next()

		// Check if there were any errors during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			statusCode := c.Writer.Status()
			if statusCode == http.StatusOK {
				statusCode = http.StatusInternalServerError
			}

			models.RespondError(
				c,
				statusCode,
				"REQUEST_ERROR",
				err.Error(),
			)
		}
	}
}

// Logger is a simple request logger middleware
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log request
		fmt.Printf("[%s] %s %s\n", c.Request.Method, c.Request.URL.Path, c.ClientIP())

		c.Next()

		// Log response status
		fmt.Printf("[%s] %s - Status: %d\n", c.Request.Method, c.Request.URL.Path, c.Writer.Status())
	}
}
