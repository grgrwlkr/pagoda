package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/backlite/ui"
	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/ent/admin"
	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/msg"
	"github.com/mikestefanello/pagoda/pkg/pager"
	"github.com/mikestefanello/pagoda/pkg/redirect"
	"github.com/mikestefanello/pagoda/pkg/routenames"
	"github.com/mikestefanello/pagoda/pkg/services"
)

type Admin struct {
	orm      *ent.Client
	admin    *admin.Handler
	backlite *ui.Handler
}

// Admin handler removed - this is now a Slack-only application

func (h *Admin) Init(c *services.Container) error {
	var err error
	h.orm = c.ORM
	h.admin = admin.NewHandler(h.orm, admin.HandlerConfig{
		ItemsPerPage: 25,
		PageQueryKey: pager.QueryKey,
		TimeFormat:   time.DateTime,
	})
	h.backlite, err = ui.NewHandler(ui.Config{
		DB:           c.Database,
		BasePath:     "/admin/tasks",
		ItemsPerPage: 25,
		ReleaseAfter: c.Config.Tasks.ReleaseAfter,
	})
	return err
}

func (h *Admin) Routes(g *echo.Group) {
	// Routes removed - this is now a Slack-only application
	// Admin panel is not needed for Slack messenger
	// All admin routes have been disabled
}

// middlewareEntityLoad is middleware to extract the entity ID and attempt to load the given entity.
func (h *Admin) middlewareEntityLoad(n admin.EntityType) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			id, err := strconv.Atoi(ctx.Param("id"))
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid entity ID")
			}

			entity, err := h.admin.Get(ctx, n, id)
			switch {
			case err == nil:
				ctx.Set(context.AdminEntityIDKey, id)
				ctx.Set(context.AdminEntityKey, map[string][]string(entity))
				return next(ctx)
			case ent.IsNotFound(err):
				return echo.NewHTTPError(http.StatusNotFound, "entity not found")
			default:
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}
		}
	}
}

// Admin entity handlers removed - this is now a Slack-only application
func (h *Admin) EntityList(n admin.EntityType) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "Admin panel not available")
	}
}

func (h *Admin) EntityAdd(n admin.EntityType) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "Admin panel not available")
	}
}

func (h *Admin) EntityAddSubmit(n admin.EntityType) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "Admin panel not available")
	}
}

func (h *Admin) EntityEdit(n admin.EntityType) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "Admin panel not available")
	}
}

func (h *Admin) EntityEditSubmit(n admin.EntityType) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "Admin panel not available")
	}
}

func (h *Admin) EntityDelete(n admin.EntityType) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "Admin panel not available")
	}
}

func (h *Admin) EntityDeleteSubmit(n admin.EntityType) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		id := ctx.Get(context.AdminEntityIDKey).(int)
		if err := h.admin.Delete(ctx, n, id); err != nil {
			msg.Error(ctx, err.Error())
			return h.EntityDelete(n)(ctx)
		}

		msg.Success(ctx, fmt.Sprintf("Successfully deleted %s (ID %d).", n.GetName(), id))

		return redirect.
			New(ctx).
			Route(routenames.AdminEntityList(n.GetName())).
			StatusCode(http.StatusFound).
			Go()
	}
}

func (h *Admin) Backlite(handler func(http.ResponseWriter, *http.Request) error) echo.HandlerFunc {
	return func(c echo.Context) error {
		if id := c.Param("id"); id != "" {
			c.Request().SetPathValue("task", id)
		}
		return handler(c.Response().Writer, c.Request())
	}
}
