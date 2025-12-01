package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/services"
)

var handlers []Handler

// Handler handles one or more HTTP routes
type Handler interface {
	// Routes allows for self-registration of HTTP routes on the router
	Routes(g *echo.Group)

	// Init provides the service container to initialize
	Init(*services.Container) error
}

// ModeHandler allows handlers to specify which app mode they should be registered for
type ModeHandler interface {
	Handler
	// ShouldRegister returns true if the handler should be registered for the given app mode
	// App modes: "pagoda" (default), "slack"
	ShouldRegister(appMode string) bool
}

// Register registers a handler
func Register(h Handler) {
	handlers = append(handlers, h)
}

// GetHandlers returns all handlers
func GetHandlers() []Handler {
	return handlers
}

// fail is a helper to fail a request by returning a 500 error and logging the error
func fail(err error, log string) error {
	// The error handler will handle logging
	return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("%s: %v", log, err))
}
