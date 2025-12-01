package middleware

// ============================================================================
// CUSTOM CODE START - Channel Middleware
// ============================================================================
// This file contains middleware for channel-related operations.
//
// File location: app/middleware/
// This is YOUR code, not part of Pagoda core.
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
			// Try to get channel ID from different parameter names
			var id int
			var err error

			if param := c.Param("id"); param != "" {
				id, err = strconv.Atoi(param)
			} else if param := c.Param("channel_id"); param != "" {
				id, err = strconv.Atoi(param)
			} else {
				return echo.NewHTTPError(http.StatusBadRequest, "channel ID is required")
			}

			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid channel ID")
			}

			channel, err := orm.Channel.Get(c.Request().Context(), id)
			if err != nil {
				if ent.IsNotFound(err) {
					return echo.NewHTTPError(http.StatusNotFound, "channel not found")
				}
				log.Ctx(c).Error("failed to load channel", "error", err, "channel_id", id)
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to load channel")
			}

			c.Set(ChannelKey, channel)
			return next(c)
		}
	}
}

// RequireChannelMember ensures the authenticated user is a member of the channel.
func RequireChannelMember(orm *ent.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := c.Get(context.AuthenticatedUserKey)
			if user == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
			}

			channel := c.Get(ChannelKey)
			if channel == nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "channel not loaded")
			}

			ch := channel.(*ent.Channel)
			userEntity := user.(*ent.User)

			exists, err := orm.ChannelMember.
				Query().
				Where(channelmember.ChannelIDEQ(ch.ID)).
				Where(channelmember.UserIDEQ(int(userEntity.ID))).
				Exist(c.Request().Context())

			if err != nil {
				log.Ctx(c).Error("failed to check channel membership", "error", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify membership")
			}

			if !exists {
				return echo.NewHTTPError(http.StatusForbidden, "you are not a member of this channel")
			}

			return next(c)
		}
	}
}

// ============================================================================
// CUSTOM CODE END
// ============================================================================
