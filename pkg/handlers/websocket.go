package handlers

// ============================================================================
// Slack Messenger WebSocket Handler
// ============================================================================
// This file handles WebSocket connections for real-time messaging.
// ============================================================================

import (
	"net/http"
	"net/url"

	gorillaWS "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/config"
	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/pkg/context"
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
	logger.Info("=== WEBSOCKET HANDLER START ===")
	logger.Info("WebSocket connection request", "remote_addr", ctx.Request().RemoteAddr, "origin", ctx.Request().Header.Get("Origin"))

	// Get authenticated user from context
	user := ctx.Get(context.AuthenticatedUserKey)
	if user == nil {
		logger.Warn("WebSocket connection failed: user not authenticated")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}

	userEntity := user.(*ent.User)
	logger.Info("Authenticated user", "user_id", userEntity.ID, "user_email", userEntity.Email)

	// Get app host from config for origin checking
	appHost := ctx.Echo().Server.Addr
	if appHost == "" {
		appHost = "localhost:8000"
	}
	logger.Info("App host", "app_host", appHost)

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

	logger.Info("Upgrading connection to WebSocket", "user_id", userEntity.ID)
	wsConn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		logger.Error("WebSocket upgrade failed", "error", err, "user_id", userEntity.ID)
		return err
	}
	logger.Info("WebSocket connection upgraded successfully", "user_id", userEntity.ID)

	// Create connection with context and ORM
	logger.Info("Creating WebSocket connection object", "user_id", userEntity.ID)
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
	logger.Info("Registering WebSocket connection", "user_id", conn.UserID)
	h.hub.Register(conn)
	logger.Info("WebSocket connection registered", "user_id", conn.UserID, "active_connections", h.hub.GetConnectionCount())

	// Start connection pumps
	logger.Info("Starting WebSocket pumps", "user_id", conn.UserID)
	go conn.WritePump()
	go conn.ReadPump()

	// Send online status
	logger.Info("Broadcasting user online event", "user_id", conn.UserID)
	onlineEvent := ws.UserOnlineEvent(conn.UserID)
	h.hub.Broadcast(onlineEvent.ToJSON())

	logger.Info("WebSocket connection established and ready", "user_id", conn.UserID)
	logger.Info("=== WEBSOCKET HANDLER END ===")

	return nil
}

// ============================================================================
// CUSTOM CODE END
// ============================================================================
