package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/routenames"
	"github.com/mikestefanello/pagoda/pkg/services"
)

type Task struct {
	// tasks removed - this is now a Slack-only application
}

// Task handler removed - this is now a Slack-only application

func (h *Task) Init(c *services.Container) error {
	return nil
}

func (h *Task) Routes(g *echo.Group) {
	g.GET("/task", h.Page).Name = routenames.Task
	g.POST("/task", h.Submit).Name = routenames.TaskSubmit
}

func (h *Task) Page(ctx echo.Context) error {
	return echo.NewHTTPError(http.StatusNotFound, "Task page not available")
}

func (h *Task) Submit(ctx echo.Context) error {
	return echo.NewHTTPError(http.StatusNotFound, "Task page not available")
}
