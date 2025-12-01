package handlers

// ============================================================================
// CUSTOM CODE START - Slack Messenger WebSocket Handler
// ============================================================================
// This file handles WebSocket connections for real-time messaging.
//
// File location: app/handlers/
// This is YOUR code, not part of Pagoda core.
// ============================================================================

import (
	"net/http"
	"net/url"

	gorillaWS "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	ws "github.com/mikestefanello/pagoda/app/websocket"
	"github.com/mikestefanello/pagoda/config"
	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/handlers"
	"github.com/mikestefanello/pagoda/pkg/log"
	"github.com/mikestefanello/pagoda/pkg/middleware"
	"github.com/mikestefanello/pagoda/pkg/routenames"
	"github.com/mikestefanello/pagoda/pkg/services"
)

// WebSocket handles WebSocket connections.
type WebSocket struct {
	hub    *ws.Hub
	config *config.Config
}

func init() {
	handlers.Register(new(WebSocket))
}

// ShouldRegister returns true if handler should be registered for the given app mode
func (h *WebSocket) ShouldRegister(appMode string) bool {
	return appMode == "slack"
}

// Init initializes the handler with dependencies from the container.
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
	// Get authenticated user from context
	user := ctx.Get(context.AuthenticatedUserKey)
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	// Get logger from context
	logger := log.Ctx(ctx)

	// Get app host from config for origin checking
	appHost := ctx.Echo().Server.Addr
	if appHost == "" {
		appHost = "localhost:8000"
	}

	// Upgrade connection to WebSocket
	upgrader := gorillaWS.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				// Allow requests without Origin header (e.g., from same origin)
				return true
			}

			originURL, err := url.Parse(origin)
			if err != nil {
				return false
			}

			// In development, allow localhost connections
			// In production, you should check against allowed origins from config
			host := originURL.Host
			if host == "" {
				return false
			}

			// Check allowed origins from config
			allowedOrigins := h.config.App.WebSocket.AllowedOrigins
			if len(allowedOrigins) > 0 {
				for _, allowed := range allowedOrigins {
					allowedURL, err := url.Parse(allowed)
					if err != nil {
						continue
					}
					if host == allowedURL.Host || host == appHost {
						return true
					}
				}
				// Also allow same origin
				return host == appHost
			}

			// Fallback: Allow same origin and localhost for development (if no config)
			return host == appHost ||
				host == "localhost:8000" ||
				host == "127.0.0.1:8000" ||
				originURL.Scheme == "http" && (host == "localhost" || host == "127.0.0.1")
		},
	}

	wsConn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		logger.Error("WebSocket upgrade failed", "error", err)
		return err
	}

	userEntity := user.(*ent.User)

	// Create connection with context and ORM
	conn := &ws.Connection{
		WS:     wsConn,
		Send:   make(chan []byte, 256),
		UserID: int64(userEntity.ID),
		Hub:    h.hub,
		Ctx:    ctx.Request().Context(),
		ORM:    h.hub.ORM,
		Logger: logger,
	}

	// Register connection
	h.hub.Register(conn)

	// Start connection pumps
	go conn.WritePump()
	go conn.ReadPump()

	// Send online status
	onlineEvent := ws.UserOnlineEvent(conn.UserID)
	h.hub.Broadcast(onlineEvent.ToJSON())

	logger.Info("WebSocket connection established", "user_id", conn.UserID)

	return nil
}

// ============================================================================
// CUSTOM CODE END
// ============================================================================
