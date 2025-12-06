package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/htmx"
	"github.com/mikestefanello/pagoda/pkg/log"
	"github.com/mikestefanello/pagoda/pkg/ui/pages"
)

type Error struct{}

func (e *Error) Page(err error, ctx echo.Context) {
	if ctx.Response().Committed || context.IsCanceledError(err) {
		return
	}

	// Determine the error status code.
	code := http.StatusInternalServerError
	errorMessage := "Internal Server Error"
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if msg, ok := he.Message.(string); ok {
			errorMessage = msg
		}
	}

	// Log the error.
	logger := log.Ctx(ctx)
	switch {
	case code >= 500:
		logger.Error(err.Error())
	case code >= 400:
		logger.Warn(err.Error())
	}

	// Set the status code.
	ctx.Response().WriteHeader(code)

	// For HTMX requests, return simple text/JSON instead of full error page
	// This prevents HTMX from replacing content with full error page
	if htmx.GetRequest(ctx).Enabled {
		// Return simple error message as text
		// HTMX will trigger htmx:response-error event which can be handled by form handlers
		ctx.Response().Header().Set("Content-Type", "text/plain")
		if _, writeErr := ctx.Response().Write([]byte(errorMessage)); writeErr != nil {
			log.Ctx(ctx).Error("failed to write error response",
				"error", writeErr,
			)
		}
		return
	}

	// Render the error page for non-HTMX requests.
	if err = pages.Error(ctx, code); err != nil {
		log.Ctx(ctx).Error("failed to render error page",
			"error", err,
		)
	}
}
