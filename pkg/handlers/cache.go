package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/routenames"
	"github.com/mikestefanello/pagoda/pkg/services"
)

type Cache struct {
	cache *services.CacheClient
}

// Cache handler removed - this is now a Slack-only application

func (h *Cache) Init(c *services.Container) error {
	h.cache = c.Cache
	return nil
}

func (h *Cache) Routes(g *echo.Group) {
	g.GET("/cache", h.Page).Name = routenames.Cache
	g.POST("/cache", h.Submit).Name = routenames.CacheSubmit
}

func (h *Cache) Page(ctx echo.Context) error {
	return echo.NewHTTPError(http.StatusNotFound, "Cache page not available")
}

func (h *Cache) Submit(ctx echo.Context) error {
	return echo.NewHTTPError(http.StatusNotFound, "Cache page not available")
}
