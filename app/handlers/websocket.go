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

	gorillaWS "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/ent"
	ws "github.com/mikestefanello/pagoda/app/websocket"
	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/handlers"
	"github.com/mikestefanello/pagoda/pkg/middleware"
	"github.com/mikestefanello/pagoda/pkg/routenames"
	"github.com/mikestefanello/pagoda/pkg/services"
)

// WebSocket handles WebSocket connections.
type WebSocket struct {
	hub *ws.Hub
}

func init() {
	handlers.Register(new(WebSocket))
}

// Init initializes the handler with dependencies from the container.
func (h *WebSocket) Init(c *services.Container) error {
	// Create WebSocket hub
	h.hub = ws.NewHub(c.ORM)
	
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

	// Upgrade connection to WebSocket
	upgrader := gorillaWS.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// TODO: Add proper origin checking based on config
			return true
		},
	}

	wsConn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		return err
	}

	// Create connection
	conn := &ws.Connection{
		WS:     wsConn,
		Send:   make(chan []byte, 256),
		UserID: int64(user.(*ent.User).ID),
		Hub:    h.hub,
	}

	// Register connection
	h.hub.Register(conn)

	// Start connection pumps
	go conn.WritePump()
	go conn.ReadPump()

	// Send online status
	onlineEvent := ws.UserOnlineEvent(conn.UserID)
	h.hub.Broadcast(onlineEvent.ToJSON())

	return nil
}

// ============================================================================
// CUSTOM CODE END
// ============================================================================

