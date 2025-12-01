package handlers

// ============================================================================
// CUSTOM CODE START - Slack Messenger Handlers
// ============================================================================
// This file contains HTTP handlers for messenger functionality.
//
// File location: app/handlers/
// This is YOUR code, not part of Pagoda core.
// ============================================================================

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	messengerMiddleware "github.com/mikestefanello/pagoda/app/middleware"
	"github.com/mikestefanello/pagoda/app/routenames"
	messengerPages "github.com/mikestefanello/pagoda/app/ui/pages/messenger"
	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/ent/channel"
	"github.com/mikestefanello/pagoda/ent/channelmember"
	"github.com/mikestefanello/pagoda/ent/message"
	"github.com/mikestefanello/pagoda/ent/workspacemember"
	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/handlers"
	"github.com/mikestefanello/pagoda/pkg/middleware"
	"github.com/mikestefanello/pagoda/pkg/services"
)

// fail is a helper to fail a request by returning a 500 error
func fail(err error, log string) error {
	return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("%s: %v", log, err))
}

// Messenger handles all messenger-related routes
type Messenger struct {
	orm *ent.Client
}

func init() {
	handlers.Register(new(Messenger))
}

// Init initializes the handler with dependencies from the container.
func (h *Messenger) Init(c *services.Container) error {
	h.orm = c.ORM
	return nil
}

// Routes registers messenger routes.
func (h *Messenger) Routes(g *echo.Group) {
	// All routes require authentication
	g = g.Group("", middleware.RequireAuthentication)

	// Workspace routes
	g.GET("/workspace", h.WorkspaceList).Name = routenames.MessengerWorkspaceList
	g.GET("/workspace/:id", h.WorkspaceView, messengerMiddleware.LoadWorkspace(h.orm)).Name = routenames.MessengerWorkspaceView
	g.POST("/workspace", h.WorkspaceCreate).Name = routenames.MessengerWorkspaceCreate
	g.PUT("/workspace/:id", h.WorkspaceUpdate, messengerMiddleware.LoadWorkspace(h.orm), messengerMiddleware.RequireWorkspaceMember(h.orm)).Name = routenames.MessengerWorkspaceUpdate
	g.DELETE("/workspace/:id", h.WorkspaceDelete, messengerMiddleware.LoadWorkspace(h.orm), messengerMiddleware.RequireWorkspaceMember(h.orm)).Name = routenames.MessengerWorkspaceDelete
	g.POST("/workspace/:id/members", h.WorkspaceAddMember, messengerMiddleware.LoadWorkspace(h.orm), messengerMiddleware.RequireWorkspaceMember(h.orm)).Name = routenames.MessengerWorkspaceAddMember
	g.DELETE("/workspace/:id/members/:user_id", h.WorkspaceRemoveMember, messengerMiddleware.LoadWorkspace(h.orm), messengerMiddleware.RequireWorkspaceMember(h.orm)).Name = routenames.MessengerWorkspaceRemoveMember

	// Channel routes
	g.GET("/workspace/:workspace_id/channels", h.ChannelList).Name = routenames.MessengerChannelList
	g.GET("/channel/:id", h.ChannelView, messengerMiddleware.LoadChannel(h.orm), messengerMiddleware.RequireChannelMember(h.orm)).Name = routenames.MessengerChannelView
	g.POST("/workspace/:workspace_id/channels", h.ChannelCreate).Name = routenames.MessengerChannelCreate
	g.PUT("/channel/:id", h.ChannelUpdate, messengerMiddleware.LoadChannel(h.orm), messengerMiddleware.RequireChannelMember(h.orm)).Name = routenames.MessengerChannelUpdate
	g.DELETE("/channel/:id", h.ChannelDelete, messengerMiddleware.LoadChannel(h.orm), messengerMiddleware.RequireChannelMember(h.orm)).Name = routenames.MessengerChannelDelete
	g.POST("/channel/:id/members", h.ChannelAddMember, messengerMiddleware.LoadChannel(h.orm), messengerMiddleware.RequireChannelMember(h.orm)).Name = routenames.MessengerChannelAddMember
	g.DELETE("/channel/:id/members/:user_id", h.ChannelRemoveMember, messengerMiddleware.LoadChannel(h.orm), messengerMiddleware.RequireChannelMember(h.orm)).Name = routenames.MessengerChannelRemoveMember
	g.GET("/channel/:id/messages", h.ChannelMessages, messengerMiddleware.LoadChannel(h.orm), messengerMiddleware.RequireChannelMember(h.orm)).Name = routenames.MessengerChannelMessages

	// Message routes
	g.POST("/channel/:channel_id/messages", h.MessageCreate, messengerMiddleware.LoadChannel(h.orm), messengerMiddleware.RequireChannelMember(h.orm)).Name = routenames.MessengerMessageCreate
	g.PUT("/message/:id", h.MessageUpdate).Name = routenames.MessengerMessageUpdate
	g.DELETE("/message/:id", h.MessageDelete).Name = routenames.MessengerMessageDelete
	g.GET("/message/:id/replies", h.MessageReplies).Name = routenames.MessengerMessageReplies
	g.POST("/message/:id/replies", h.MessageReply).Name = routenames.MessengerMessageReply

	// Direct Message routes
	g.GET("/dms", h.DMList).Name = routenames.MessengerDMList
	g.GET("/dm/:id", h.DMView).Name = routenames.MessengerDMView
	g.POST("/dm", h.DMCreate).Name = routenames.MessengerDMCreate
	g.GET("/dm/:id/messages", h.DMMessages).Name = routenames.MessengerDMMessages
	g.POST("/dm/:id/messages", h.DMMessageCreate).Name = routenames.MessengerDMMessageCreate

	// Reaction routes
	g.POST("/message/:id/reactions", h.ReactionAdd).Name = routenames.MessengerReactionAdd
	g.DELETE("/message/:id/reactions/:emoji", h.ReactionRemove).Name = routenames.MessengerReactionRemove

	// Attachment routes
	g.POST("/message/:id/attachments", h.AttachmentUpload).Name = routenames.MessengerAttachmentUpload
	g.GET("/attachment/:id", h.AttachmentView).Name = routenames.MessengerAttachmentView
	g.DELETE("/attachment/:id", h.AttachmentDelete).Name = routenames.MessengerAttachmentDelete
}

// ============================================================================
// Workspace Handlers
// ============================================================================

// WorkspaceList returns a list of workspaces for the authenticated user
func (h *Messenger) WorkspaceList(ctx echo.Context) error {
	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	workspaces, err := h.orm.WorkspaceMember.
		Query().
		Where(workspacemember.UserIDEQ(int(user.ID))).
		QueryWorkspace().
		All(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to fetch workspaces")
	}

	// TODO: Render workspace list page or return JSON
	return ctx.JSON(http.StatusOK, workspaces)
}

// WorkspaceView shows a workspace and redirects to the workspace page
func (h *Messenger) WorkspaceView(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}

	_, err = h.orm.Workspace.Get(ctx.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "workspace not found")
	}

	// TODO: Check if user is a member
	// TODO: Render workspace page
	return messengerPages.Workspace(ctx)
}

// WorkspaceCreate creates a new workspace
func (h *Messenger) WorkspaceCreate(ctx echo.Context) error {
	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// TODO: Parse form data
	// For now, create a default workspace
	ws, err := h.orm.Workspace.
		Create().
		SetName("My Workspace").
		SetSlug("my-workspace").
		SetOwnerID(int(user.ID)).
		Save(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to create workspace")
	}

	// Add creator as owner member
	_, err = h.orm.WorkspaceMember.
		Create().
		SetWorkspaceID(ws.ID).
		SetUserID(int(user.ID)).
		SetRole("owner").
		Save(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to add workspace member")
	}

	return ctx.JSON(http.StatusCreated, ws)
}

// WorkspaceUpdate updates a workspace
func (h *Messenger) WorkspaceUpdate(ctx echo.Context) error {
	_, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}

	// TODO: Parse form data and update
	// TODO: Check permissions (owner/admin only)

	return ctx.JSON(http.StatusOK, map[string]string{"status": "updated"})
}

// WorkspaceDelete deletes a workspace
func (h *Messenger) WorkspaceDelete(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}

	// TODO: Check permissions (owner only)
	err = h.orm.Workspace.DeleteOneID(id).Exec(ctx.Request().Context())
	if err != nil {
		return fail(err, "failed to delete workspace")
	}

	return ctx.NoContent(http.StatusNoContent)
}

// WorkspaceAddMember adds a member to a workspace
func (h *Messenger) WorkspaceAddMember(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// WorkspaceRemoveMember removes a member from a workspace
func (h *Messenger) WorkspaceRemoveMember(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// ============================================================================
// Channel Handlers
// ============================================================================

// ChannelList returns a list of channels in a workspace
func (h *Messenger) ChannelList(ctx echo.Context) error {
	workspaceID, err := strconv.Atoi(ctx.Param("workspace_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get channels where user is a member
	channels, err := h.orm.ChannelMember.
		Query().
		Where(channelmember.UserIDEQ(int(user.ID))).
		QueryChannel().
		Where(channel.WorkspaceIDEQ(workspaceID)).
		All(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to fetch channels")
	}

	return ctx.JSON(http.StatusOK, channels)
}

// ChannelView shows a channel page
func (h *Messenger) ChannelView(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid channel ID")
	}

	ch, err := h.orm.Channel.Get(ctx.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "channel not found")
	}

	// TODO: Check if user is a member
	return messengerPages.Channel(ctx, int64(ch.ID), ch.Name)
}

// ChannelCreate creates a new channel
func (h *Messenger) ChannelCreate(ctx echo.Context) error {
	workspaceID, err := strconv.Atoi(ctx.Param("workspace_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// TODO: Parse form data
	// For now, create a default channel
	ch, err := h.orm.Channel.
		Create().
		SetName("New Channel").
		SetSlug("new-channel").
		SetWorkspaceID(workspaceID).
		SetCreatedBy(int(user.ID)).
		Save(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to create channel")
	}

	// Add creator as member
	_, err = h.orm.ChannelMember.
		Create().
		SetChannelID(ch.ID).
		SetUserID(int(user.ID)).
		Save(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to add channel member")
	}

	return ctx.JSON(http.StatusCreated, ch)
}

// ChannelUpdate updates a channel
func (h *Messenger) ChannelUpdate(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// ChannelDelete deletes a channel
func (h *Messenger) ChannelDelete(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// ChannelAddMember adds a member to a channel
func (h *Messenger) ChannelAddMember(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// ChannelRemoveMember removes a member from a channel
func (h *Messenger) ChannelRemoveMember(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// ChannelMessages returns messages in a channel with pagination
func (h *Messenger) ChannelMessages(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid channel ID")
	}

	// TODO: Add pagination
	messages, err := h.orm.Message.
		Query().
		Where(message.ChannelIDEQ(id)).
		Order(ent.Desc(message.FieldCreatedAt)).
		Limit(50).
		All(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to fetch messages")
	}

	return ctx.JSON(http.StatusOK, messages)
}

// ============================================================================
// Message Handlers
// ============================================================================

// MessageCreate creates a new message in a channel
func (h *Messenger) MessageCreate(ctx echo.Context) error {
	channelID, err := strconv.Atoi(ctx.Param("channel_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid channel ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// TODO: Parse form data
	content := ctx.FormValue("content")
	if content == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "message content is required")
	}

	msg, err := h.orm.Message.
		Create().
		SetContent(content).
		SetMessageType(message.MessageTypeText).
		SetChannelID(channelID).
		SetUserID(int(user.ID)).
		Save(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to create message")
	}

	// TODO: Send WebSocket event

	return ctx.JSON(http.StatusCreated, msg)
}

// MessageUpdate updates a message
func (h *Messenger) MessageUpdate(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// MessageDelete deletes a message
func (h *Messenger) MessageDelete(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// MessageReplies returns replies to a message (thread)
func (h *Messenger) MessageReplies(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// MessageReply creates a reply to a message (thread)
func (h *Messenger) MessageReply(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// ============================================================================
// Direct Message Handlers
// ============================================================================

// DMList returns a list of direct message conversations
func (h *Messenger) DMList(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, []interface{}{})
}

// DMView shows a direct message conversation
func (h *Messenger) DMView(ctx echo.Context) error {
	// TODO: Implement
	return messengerPages.DirectMessage(ctx, 0, "User")
}

// DMCreate creates or retrieves a direct message conversation
func (h *Messenger) DMCreate(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// DMMessages returns messages in a direct message conversation
func (h *Messenger) DMMessages(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, []interface{}{})
}

// DMMessageCreate creates a message in a direct message conversation
func (h *Messenger) DMMessageCreate(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// ============================================================================
// Reaction Handlers
// ============================================================================

// ReactionAdd adds a reaction to a message
func (h *Messenger) ReactionAdd(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// ReactionRemove removes a reaction from a message
func (h *Messenger) ReactionRemove(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// ============================================================================
// Attachment Handlers
// ============================================================================

// AttachmentUpload uploads a file attachment
func (h *Messenger) AttachmentUpload(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// AttachmentView serves an attachment file
func (h *Messenger) AttachmentView(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// AttachmentDelete deletes an attachment
func (h *Messenger) AttachmentDelete(ctx echo.Context) error {
	// TODO: Implement
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not implemented"})
}

// ============================================================================
// CUSTOM CODE END
// ============================================================================
