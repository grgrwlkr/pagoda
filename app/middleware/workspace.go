package middleware

// ============================================================================
// CUSTOM CODE START - Workspace Middleware
// ============================================================================
// This file contains middleware for workspace-related operations.
//
// File location: app/middleware/
// This is YOUR code, not part of Pagoda core.
// ============================================================================

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/ent/workspacemember"
	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/log"
)

const (
	// WorkspaceKey is the key used to store the workspace in context.
	WorkspaceKey = "workspace"
)

// LoadWorkspace loads a workspace from the ID parameter and stores it in context.
func LoadWorkspace(orm *ent.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
			}

			workspace, err := orm.Workspace.Get(c.Request().Context(), id)
			if err != nil {
				if ent.IsNotFound(err) {
					return echo.NewHTTPError(http.StatusNotFound, "workspace not found")
				}
				log.Ctx(c).Error("failed to load workspace", "error", err, "workspace_id", id)
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to load workspace")
			}

			c.Set(WorkspaceKey, workspace)
			return next(c)
		}
	}
}

// RequireWorkspaceMember ensures the authenticated user is a member of the workspace.
func RequireWorkspaceMember(orm *ent.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := c.Get(context.AuthenticatedUserKey)
			if user == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
			}

			workspace := c.Get(WorkspaceKey)
			if workspace == nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "workspace not loaded")
			}

			ws := workspace.(*ent.Workspace)
			userEntity := user.(*ent.User)

			exists, err := orm.WorkspaceMember.
				Query().
				Where(workspacemember.WorkspaceIDEQ(ws.ID)).
				Where(workspacemember.UserIDEQ(int(userEntity.ID))).
				Exist(c.Request().Context())

			if err != nil {
				log.Ctx(c).Error("failed to check workspace membership", "error", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify membership")
			}

			if !exists {
				return echo.NewHTTPError(http.StatusForbidden, "you are not a member of this workspace")
			}

			return next(c)
		}
	}
}

// ============================================================================
// CUSTOM CODE END
// ============================================================================
