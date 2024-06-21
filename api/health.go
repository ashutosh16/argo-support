package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type HealthService struct {
}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s *HealthService) RegisterHandlers(routes gin.IRoutes) {
	routes.GET("/health/full", s.Health)
}

func (s *HealthService) Health(c *gin.Context) {
	c.JSON(http.StatusOK, "Health Check OK")
}
