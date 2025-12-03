package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/log"
	"github.com/mikestefanello/pagoda/pkg/msg"
	"github.com/mikestefanello/pagoda/pkg/routenames"
	"github.com/mikestefanello/pagoda/pkg/services"

	"github.com/labstack/echo/v4"
)

// LoadAuthenticatedUser loads the authenticated user, if one, and stores in context.
func LoadAuthenticatedUser(authClient *services.AuthClient) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := log.Ctx(c)
			logger.Info("=== LOAD AUTHENTICATED USER MIDDLEWARE START ===")
			logger.Info("Request URL", "url", c.Request().URL.String())
			logger.Info("Request method", "method", c.Request().Method)

			// Log cookies
			cookies := c.Request().Cookies()
			logger.Info("Request cookies", "count", len(cookies))
			for _, cookie := range cookies {
				logger.Info("Cookie", "name", cookie.Name, "value_length", len(cookie.Value), "domain", cookie.Domain, "path", cookie.Path, "secure", cookie.Secure, "http_only", cookie.HttpOnly)
			}

			u, err := authClient.GetAuthenticatedUser(c)
			switch err.(type) {
			case *ent.NotFoundError:
				logger.Warn("auth user not found", "error", err)
			case services.NotAuthenticatedError:
				logger.Info("User not authenticated", "error", err)
			case nil:
				logger.Info("User authenticated and loaded", "user_id", u.ID, "user_email", u.Email, "user_name", u.Name)
				c.Set(context.AuthenticatedUserKey, u)
			default:
				logger.Error("Error querying for authenticated user", "error", err)
				return echo.NewHTTPError(
					http.StatusInternalServerError,
					fmt.Sprintf("error querying for authenticated user: %v", err),
				)
			}

			logger.Info("=== LOAD AUTHENTICATED USER MIDDLEWARE END ===")
			return next(c)
		}
	}
}

// LoadValidPasswordToken loads a valid password token entity that matches the user and token
// provided in path parameters
// If the token is invalid, the user will be redirected to the forgot password route
// This requires that the user owning the token is loaded in to context.
func LoadValidPasswordToken(authClient *services.AuthClient) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract the user parameter
			if c.Get(context.UserKey) == nil {
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			usr := c.Get(context.UserKey).(*ent.User)

			// Extract the token ID.
			tokenID, err := strconv.Atoi(c.Param("password_token"))
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound)
			}

			// Attempt to load a valid password token.
			token, err := authClient.GetValidPasswordToken(
				c,
				usr.ID,
				tokenID,
				c.Param("token"),
			)

			switch err.(type) {
			case nil:
				c.Set(context.PasswordTokenKey, token)
				return next(c)
			case services.InvalidPasswordTokenError:
				msg.Warning(c, "The link is either invalid or has expired. Please request a new one.")
				return c.Redirect(http.StatusFound, c.Echo().Reverse(routenames.ForgotPassword))
			default:
				return echo.NewHTTPError(
					http.StatusInternalServerError,
					fmt.Sprintf("error loading password token: %v", err),
				)
			}
		}
	}
}

// RequireAuthentication requires that the user be authenticated in order to proceed.
func RequireAuthentication(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if u := c.Get(context.AuthenticatedUserKey); u == nil {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		return next(c)
	}
}

// RequireNoAuthentication requires that the user not be authenticated in order to proceed.
func RequireNoAuthentication(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if u := c.Get(context.AuthenticatedUserKey); u != nil {
			return echo.NewHTTPError(http.StatusForbidden)
		}

		return next(c)
	}
}

// RequireAdmin requires that the authenticated user be an admin in order to proceed.
func RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if u := c.Get(context.AuthenticatedUserKey); u != nil {
			if user, ok := u.(*ent.User); ok {
				if user.Admin {
					return next(c)
				}
			}
		}

		return echo.NewHTTPError(http.StatusUnauthorized)
	}
}
