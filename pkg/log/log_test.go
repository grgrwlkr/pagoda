package log

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCtxSet(t *testing.T) {
	// Create Echo context directly to avoid circular import with pkg/tests
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	logger := Ctx(ctx)
	assert.NotNil(t, logger)

	logger = logger.With("a", "b")
	Set(ctx, logger)

	got := Ctx(ctx)
	assert.Equal(t, got, logger)
}
