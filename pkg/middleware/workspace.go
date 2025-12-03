package middleware

// ============================================================================
// Workspace Middleware
// ============================================================================
// This file contains middleware for workspace-related operations.
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
			logger := log.Ctx(c)
			logger.Debug("=== MIDDLEWARE: LOAD WORKSPACE START ===")

			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				logger.Warn("Invalid workspace ID parameter", "error", err, "id_param", c.Param("id"))
				return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
			}
			logger.Debug("Loading workspace", "workspace_id", id)

			workspace, err := orm.Workspace.Get(c.Request().Context(), id)
			if err != nil {
				if ent.IsNotFound(err) {
					logger.Warn("Workspace not found", "workspace_id", id)
					return echo.NewHTTPError(http.StatusNotFound, "workspace not found")
				}
				logger.Error("Failed to load workspace", "error", err, "workspace_id", id)
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to load workspace")
			}

			logger.Debug("Workspace loaded successfully", "workspace_id", workspace.ID, "workspace_name", workspace.Name)
			c.Set(WorkspaceKey, workspace)
			logger.Debug("=== MIDDLEWARE: LOAD WORKSPACE END ===")
			return next(c)
		}
	}
}

// RequireWorkspaceMember ensures the authenticated user is a member of the workspace.
func RequireWorkspaceMember(orm *ent.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := log.Ctx(c)
			logger.Debug("=== MIDDLEWARE: REQUIRE WORKSPACE MEMBER START ===")

			user := c.Get(context.AuthenticatedUserKey)
			if user == nil {
				logger.Warn("User not authenticated in RequireWorkspaceMember middleware")
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
			}

			workspace := c.Get(WorkspaceKey)
			if workspace == nil {
				logger.Error("Workspace not loaded in context")
				return echo.NewHTTPError(http.StatusInternalServerError, "workspace not loaded")
			}

			ws := workspace.(*ent.Workspace)
			userEntity := user.(*ent.User)
			logger.Debug("Checking workspace membership", "workspace_id", ws.ID, "user_id", userEntity.ID)

			exists, err := orm.WorkspaceMember.
				Query().
				Where(workspacemember.WorkspaceIDEQ(ws.ID)).
				Where(workspacemember.UserIDEQ(int(userEntity.ID))).
				Exist(c.Request().Context())

			if err != nil {
				logger.Error("Failed to check workspace membership", "error", err, "workspace_id", ws.ID, "user_id", userEntity.ID)
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify membership")
			}

			if !exists {
				logger.Warn("User is not a member of workspace", "workspace_id", ws.ID, "user_id", userEntity.ID)
				return echo.NewHTTPError(http.StatusForbidden, "you are not a member of this workspace")
			}

			logger.Debug("Workspace membership verified", "workspace_id", ws.ID, "user_id", userEntity.ID)
			logger.Debug("=== MIDDLEWARE: REQUIRE WORKSPACE MEMBER END ===")
			return next(c)
		}
	}
}

// ============================================================================
// CUSTOM CODE END
// ============================================================================
