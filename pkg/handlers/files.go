package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/routenames"
	"github.com/mikestefanello/pagoda/pkg/services"
	"github.com/spf13/afero"
)

type Files struct {
	files afero.Fs
}

// Files handler removed - this is now a Slack-only application

func (h *Files) Init(c *services.Container) error {
	h.files = c.Files
	return nil
}

func (h *Files) Routes(g *echo.Group) {
	g.GET("/files", h.Page).Name = routenames.Files
	g.POST("/files", h.Submit).Name = routenames.FilesSubmit
}

func (h *Files) Page(ctx echo.Context) error {
	return echo.NewHTTPError(http.StatusNotFound, "Files page not available")
}

func (h *Files) Submit(ctx echo.Context) error {
	return echo.NewHTTPError(http.StatusNotFound, "Files page not available")
}
