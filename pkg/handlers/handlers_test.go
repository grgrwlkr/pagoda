package handlers

import (
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler is a test handler for testing
type TestHandler struct{}

func (h *TestHandler) Init(c *services.Container) error { return nil }
func (h *TestHandler) Routes(g *echo.Group)             {}

func TestGetSetHandlers(t *testing.T) {
	handlers = []Handler{}
	assert.Empty(t, GetHandlers())
	// Pages handler removed - using test handler instead
	h := new(TestHandler)
	Register(h)
	got := GetHandlers()
	require.Len(t, got, 1)
	assert.Equal(t, h, got[0])
}

func TestFail(t *testing.T) {
	err := fail(errors.New("err message"), "log message")
	require.IsType(t, new(echo.HTTPError), err)
	he := err.(*echo.HTTPError)
	assert.Equal(t, http.StatusInternalServerError, he.Code)
	assert.Equal(t, "log message: err message", he.Message)
}
