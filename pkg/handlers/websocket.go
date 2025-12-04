package handlers

// ============================================================================
// Slack Messenger WebSocket Handler
// ============================================================================
// This file handles WebSocket connections for real-time messaging.
// ============================================================================

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	gorillaWS "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/config"
	"github.com/mikestefanello/pagoda/ent"
	pkgcontext "github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/log"
	"github.com/mikestefanello/pagoda/pkg/middleware"
	"github.com/mikestefanello/pagoda/pkg/routenames"
	"github.com/mikestefanello/pagoda/pkg/services"
	ws "github.com/mikestefanello/pagoda/pkg/websocket"
)

// WebSocket handles WebSocket connections.
type WebSocket struct {
	hub    *ws.Hub
	config *config.Config
}

func init() {
	Register(new(WebSocket))
}

// Init initializes the handler with dependencies from the container.
// Init initializes the handler with dependencies from the container.
// Creates and starts the WebSocket hub for real-time communication.
// Parameters:
//   - c: service container with all dependencies
//
// Returns:
//   - error: initialization error if any dependency is missing
func (h *WebSocket) Init(c *services.Container) error {
	// Create WebSocket hub
	h.hub = ws.NewHub(c.ORM)
	h.config = c.Config

	// Set global hub for access from other handlers
	ws.SetHub(h.hub)

	// Start the hub in a goroutine
	go h.hub.Run()

	return nil
}

// Routes registers WebSocket routes.
func (h *WebSocket) Routes(g *echo.Group) {
	// WebSocket endpoint - requires authentication
	g.GET("/ws", h.HandleWebSocket, middleware.RequireAuthentication).Name = routenames.WebSocket
}

// HandleWebSocket handles WebSocket upgrade requests.
func (h *WebSocket) HandleWebSocket(ctx echo.Context) error {
	logger := log.Ctx(ctx)

	// Get authenticated user from context
	user := ctx.Get(pkgcontext.AuthenticatedUserKey)
	if user == nil {
		logger.Warn("WebSocket connection failed: user not authenticated")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	userEntity := user.(*ent.User)

	// Get app host from config for origin checking
	appHostURL, err := url.Parse(h.config.App.Host)
	var appHost string
	if err == nil && appHostURL.Host != "" {
		appHost = appHostURL.Host
	} else {
		// Fallback to server address or default
		appHost = ctx.Echo().Server.Addr
		if appHost == "" {
			appHost = "localhost:8000"
		}
	}
	logger.Info("App host for origin check", "app_host", appHost, "origin_header", ctx.Request().Header.Get("Origin"))

	// Upgrade connection to WebSocket
	upgrader := gorillaWS.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			logger.Info("Checking WebSocket origin", "origin", origin, "app_host", appHost, "request_host", r.Host)

			// Allow requests without Origin header (e.g., from same origin or browser extensions)
			if origin == "" {
				logger.Debug("No Origin header, allowing connection")
				return true
			}

			originURL, err := url.Parse(origin)
			if err != nil {
				logger.Warn("Failed to parse origin URL", "origin", origin, "error", err)
				return false
			}

			originHost := originURL.Host
			if originHost == "" {
				logger.Warn("Empty origin host", "origin", origin)
				return false
			}

			// Normalize hosts for comparison (remove port if default)
			normalizeHost := func(host string) string {
				host = strings.TrimSuffix(host, ":80")
				host = strings.TrimSuffix(host, ":443")
				return host
			}

			normalizedOriginHost := normalizeHost(originHost)
			normalizedAppHost := normalizeHost(appHost)

			// Check allowed origins from config
			allowedOrigins := h.config.App.WebSocket.AllowedOrigins
			if len(allowedOrigins) > 0 {
				for _, allowed := range allowedOrigins {
					allowedURL, err := url.Parse(allowed)
					if err != nil {
						continue
					}
					allowedHost := normalizeHost(allowedURL.Host)
					// Match host exactly (with or without port)
					if normalizedOriginHost == allowedHost || originHost == allowedURL.Host {
						logger.Debug("Origin allowed by config", "origin", origin, "allowed", allowed)
						return true
					}
				}
			}

			// Allow same origin
			if normalizedOriginHost == normalizedAppHost || originHost == appHost {
				logger.Debug("Origin matches app host", "origin_host", originHost, "app_host", appHost)
				return true
			}

			// Fallback: Allow localhost connections for development
			if h.config.App.Environment == "local" || h.config.App.Environment == "dev" {
				allowed := normalizedOriginHost == "localhost" ||
					normalizedOriginHost == "127.0.0.1" ||
					originHost == "localhost:8000" ||
					originHost == "127.0.0.1:8000" ||
					originHost == "localhost" ||
					originHost == "127.0.0.1" ||
					(originURL.Scheme == "http" && (normalizedOriginHost == "localhost" || normalizedOriginHost == "127.0.0.1"))
				if allowed {
					logger.Debug("Origin allowed for development", "origin", origin)
					return true
				}
			}

			logger.Warn("Origin not allowed", "origin", origin, "origin_host", originHost, "app_host", appHost)
			return false
		},
	}

	wsConn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		logger.Error("WebSocket upgrade failed", "error", err, "user_id", userEntity.ID)
		return err
	}

	logger.Info("WebSocket connection upgraded successfully", "user_id", userEntity.ID)

	// Create a new context for WebSocket connection that won't be canceled
	// The HTTP request context is canceled after the upgrade, so we need a separate context
	// that will live for the lifetime of the WebSocket connection
	wsCtx := context.Background()

	// Create connection with context and ORM
	conn := &ws.Connection{
		WS:     wsConn,
		Send:   make(chan []byte, 256),
		UserID: int64(userEntity.ID),
		Hub:    h.hub,
		Ctx:    wsCtx, // Use WebSocket-specific context that won't be canceled
		ORM:    h.hub.ORM,
		Logger: logger,
	}

	// Register connection (non-blocking, sends to channel)
	h.hub.Register(conn)
	logger.Info("WebSocket connection registered", "user_id", userEntity.ID)

	// Start connection pumps in goroutines
	go conn.WritePump()
	go conn.ReadPump()

	logger.Info("WebSocket pumps started", "user_id", userEntity.ID)

	// Send online status
	onlineEvent := ws.UserOnlineEvent(conn.UserID)
	h.hub.Broadcast(onlineEvent.ToJSON())

	// For WebSocket connections, we don't return an error
	// The connection is handled by the pumps in goroutines
	// Returning nil tells Echo that the response has been handled
	return nil
}

// ============================================================================
// CUSTOM CODE END
// ============================================================================
