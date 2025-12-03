package middleware

// ============================================================================
// Channel Middleware
// ============================================================================
// This file contains middleware for channel-related operations.
// ============================================================================

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/ent/channelmember"
	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/log"
)

const (
	// ChannelKey is the key used to store the channel in context.
	ChannelKey = "channel"
)

// LoadChannel loads a channel from the ID parameter and stores it in context.
func LoadChannel(orm *ent.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := log.Ctx(c)
			logger.Debug("=== MIDDLEWARE: LOAD CHANNEL START ===")

			// Try to get channel ID from different parameter names
			var id int
			var err error

			if param := c.Param("id"); param != "" {
				id, err = strconv.Atoi(param)
				logger.Debug("Channel ID from 'id' parameter", "channel_id", id)
			} else if param := c.Param("channel_id"); param != "" {
				id, err = strconv.Atoi(param)
				logger.Debug("Channel ID from 'channel_id' parameter", "channel_id", id)
			} else {
				logger.Warn("Channel ID parameter not found")
				return echo.NewHTTPError(http.StatusBadRequest, "channel ID is required")
			}

			if err != nil {
				logger.Warn("Invalid channel ID parameter", "error", err, "id_param", c.Param("id"), "channel_id_param", c.Param("channel_id"))
				return echo.NewHTTPError(http.StatusBadRequest, "invalid channel ID")
			}

			logger.Debug("Loading channel", "channel_id", id)
			channel, err := orm.Channel.Get(c.Request().Context(), id)
			if err != nil {
				if ent.IsNotFound(err) {
					logger.Warn("Channel not found", "channel_id", id)
					return echo.NewHTTPError(http.StatusNotFound, "channel not found")
				}
				logger.Error("Failed to load channel", "error", err, "channel_id", id)
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to load channel")
			}

			logger.Debug("Channel loaded successfully", "channel_id", channel.ID, "channel_name", channel.Name, "workspace_id", channel.WorkspaceID)
			c.Set(ChannelKey, channel)
			logger.Debug("=== MIDDLEWARE: LOAD CHANNEL END ===")
			return next(c)
		}
	}
}

// RequireChannelMember ensures the authenticated user is a member of the channel.
func RequireChannelMember(orm *ent.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := log.Ctx(c)
			logger.Debug("=== MIDDLEWARE: REQUIRE CHANNEL MEMBER START ===")

			user := c.Get(context.AuthenticatedUserKey)
			if user == nil {
				logger.Warn("User not authenticated in RequireChannelMember middleware")
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
			}

			channel := c.Get(ChannelKey)
			if channel == nil {
				logger.Error("Channel not loaded in context")
				return echo.NewHTTPError(http.StatusInternalServerError, "channel not loaded")
			}

			ch := channel.(*ent.Channel)
			userEntity := user.(*ent.User)
			logger.Debug("Checking channel membership", "channel_id", ch.ID, "user_id", userEntity.ID)

			exists, err := orm.ChannelMember.
				Query().
				Where(channelmember.ChannelIDEQ(ch.ID)).
				Where(channelmember.UserIDEQ(int(userEntity.ID))).
				Exist(c.Request().Context())

			if err != nil {
				logger.Error("Failed to check channel membership", "error", err, "channel_id", ch.ID, "user_id", userEntity.ID)
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify membership")
			}

			if !exists {
				logger.Warn("User is not a member of channel", "channel_id", ch.ID, "user_id", userEntity.ID)
				return echo.NewHTTPError(http.StatusForbidden, "you are not a member of this channel")
			}

			logger.Debug("Channel membership verified", "channel_id", ch.ID, "user_id", userEntity.ID)
			logger.Debug("=== MIDDLEWARE: REQUIRE CHANNEL MEMBER END ===")
			return next(c)
		}
	}
}

// ============================================================================
// CUSTOM CODE END
// ============================================================================
