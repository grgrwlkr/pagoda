package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/services"
)

type Contact struct {
	mail *services.MailClient
}

// Contact handler removed - this is now a Slack-only application

func (h *Contact) Init(c *services.Container) error {
	h.mail = c.Mail
	return nil
}

func (h *Contact) Routes(g *echo.Group) {
	// Routes removed - this is now a Slack-only application
}

func (h *Contact) Page(ctx echo.Context) error {
	return echo.NewHTTPError(http.StatusNotFound, "Contact page not available")
}

func (h *Contact) Submit(ctx echo.Context) error {
	return echo.NewHTTPError(http.StatusNotFound, "Contact page not available")
}
