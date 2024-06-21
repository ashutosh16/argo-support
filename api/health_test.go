package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthService_Health(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test router
	router := gin.New()

	// Create an instance of the service and register the handler
	service := NewHealthService()
	service.RegisterHandlers(router)

	// Create a test request to the health endpoint
	req, _ := http.NewRequest(http.MethodGet, "/health/full", nil)
	resp := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(resp, req)

	// Check the status code and response
	assert.Equal(t, http.StatusOK, resp.Code, "Expected HTTP status code 200")
	assert.Equal(t, "\"Health Check OK\"", resp.Body.String(), "Expected response body 'Health Check OK'")
}
