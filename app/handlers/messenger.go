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
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	messengerMiddleware "github.com/mikestefanello/pagoda/app/middleware"
	"github.com/mikestefanello/pagoda/app/routenames"
	messengerComponents "github.com/mikestefanello/pagoda/app/ui/components/messenger"
	messengerPages "github.com/mikestefanello/pagoda/app/ui/pages/messenger"
	ws "github.com/mikestefanello/pagoda/app/websocket"
	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/ent/channel"
	"github.com/mikestefanello/pagoda/ent/channelmember"
	"github.com/mikestefanello/pagoda/ent/directmessage"
	"github.com/mikestefanello/pagoda/ent/directmessagecontent"
	"github.com/mikestefanello/pagoda/ent/message"
	"github.com/mikestefanello/pagoda/ent/reaction"
	"github.com/mikestefanello/pagoda/ent/workspace"
	"github.com/mikestefanello/pagoda/ent/workspacemember"
	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/handlers"
	"github.com/mikestefanello/pagoda/pkg/log"
	"github.com/mikestefanello/pagoda/pkg/middleware"
	"github.com/mikestefanello/pagoda/pkg/pager"
	"github.com/mikestefanello/pagoda/pkg/redirect"
	"github.com/mikestefanello/pagoda/pkg/services"
	"github.com/spf13/afero"
)

// fail is a helper to fail a request by returning a 500 error
func fail(err error, log string) error {
	return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("%s: %v", log, err))
}

// getWorkspaceMemberRole returns the role of a user in a workspace
func (h *Messenger) getWorkspaceMemberRole(ctx echo.Context, workspaceID, userID int) (workspacemember.Role, error) {
	member, err := h.orm.WorkspaceMember.
		Query().
		Where(workspacemember.WorkspaceIDEQ(workspaceID)).
		Where(workspacemember.UserIDEQ(userID)).
		Only(ctx.Request().Context())

	if err != nil {
		return "", err
	}

	return member.Role, nil
}

// requireWorkspaceOwnerOrAdmin checks if user is owner or admin of workspace
func (h *Messenger) requireWorkspaceOwnerOrAdmin(ctx echo.Context, workspaceID, userID int) error {
	role, err := h.getWorkspaceMemberRole(ctx, workspaceID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusForbidden, "you are not a member of this workspace")
		}
		return fail(err, "failed to check workspace role")
	}

	if role != workspacemember.RoleOwner && role != workspacemember.RoleAdmin {
		return echo.NewHTTPError(http.StatusForbidden, "only owners and admins can perform this action")
	}

	return nil
}

// requireWorkspaceOwner checks if user is owner of workspace
func (h *Messenger) requireWorkspaceOwner(ctx echo.Context, workspaceID, userID int) error {
	role, err := h.getWorkspaceMemberRole(ctx, workspaceID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusForbidden, "you are not a member of this workspace")
		}
		return fail(err, "failed to check workspace role")
	}

	if role != workspacemember.RoleOwner {
		return echo.NewHTTPError(http.StatusForbidden, "only workspace owner can perform this action")
	}

	return nil
}

// getSidebarData loads sidebar data for a workspace
func (h *Messenger) getSidebarData(ctx echo.Context, workspaceID int, userID int, activeChannelID *int, activeDMID *int) (messengerComponents.SidebarData, error) {
	// Get workspace
	ws, err := h.orm.Workspace.Get(ctx.Request().Context(), workspaceID)
	if err != nil {
		return messengerComponents.SidebarData{}, err
	}

	// Get channels where user is a member
	channels, err := h.orm.ChannelMember.
		Query().
		Where(channelmember.UserIDEQ(userID)).
		QueryChannel().
		Where(channel.WorkspaceIDEQ(workspaceID)).
		All(ctx.Request().Context())

	if err != nil {
		return messengerComponents.SidebarData{}, err
	}

	channelData := make([]messengerComponents.ChannelData, len(channels))
	for i, ch := range channels {
		isActive := activeChannelID != nil && ch.ID == *activeChannelID
		channelData[i] = messengerComponents.ChannelData{
			ID:       int64(ch.ID),
			Slug:     ch.Slug,
			Name:     ch.Name,
			IsActive: isActive,
		}
	}

	// Get direct messages
	dms, err := h.orm.DirectMessage.
		Query().
		Where(
			directmessage.Or(
				directmessage.User1IDEQ(userID),
				directmessage.User2IDEQ(userID),
			),
		).
		Order(ent.Desc(directmessage.FieldLastMessageAt)).
		All(ctx.Request().Context())

	if err != nil {
		return messengerComponents.SidebarData{}, err
	}

	dmData := make([]messengerComponents.DMData, 0, len(dms))
	for _, dm := range dms {
		var otherUserID int
		var otherUser *ent.User
		if dm.User1ID == userID {
			otherUserID = dm.User2ID
		} else {
			otherUserID = dm.User1ID
		}

		otherUser, err = h.orm.User.Get(ctx.Request().Context(), otherUserID)
		if err != nil {
			continue // Skip if user not found
		}

		var isActive bool
		if activeDMID != nil {
			isActive = int64(dm.ID) == int64(*activeDMID)
		}
		dmData = append(dmData, messengerComponents.DMData{
			ID:       int64(dm.ID),
			UserID:   int64(otherUserID),
			UserName: otherUser.Name,
			IsActive: isActive,
		})
	}

	var activeChID *int64
	if activeChannelID != nil {
		chID := int64(*activeChannelID)
		activeChID = &chID
	}

	var activeDMID64 *int64
	if activeDMID != nil {
		dmID := int64(*activeDMID)
		activeDMID64 = &dmID
	}

	return messengerComponents.SidebarData{
		WorkspaceID:     int64(workspaceID),
		WorkspaceName:   ws.Name,
		Channels:        channelData,
		DirectMessages:  dmData,
		ActiveChannelID: activeChID,
		ActiveDMID:      activeDMID64,
	}, nil
}

// Messenger handles all messenger-related routes
type Messenger struct {
	orm   *ent.Client
	hub   *ws.Hub // WebSocket hub for real-time events
	files afero.Fs
}

func init() {
	handlers.Register(new(Messenger))
}

// ShouldRegister returns true if handler should be registered for the given app mode
func (h *Messenger) ShouldRegister(appMode string) bool {
	return appMode == "slack"
}

// Init initializes the handler with dependencies from the container.
func (h *Messenger) Init(c *services.Container) error {
	h.orm = c.ORM
	h.files = c.Files
	// Get hub from WebSocket handler (will be set when WebSocket handler initializes)
	h.hub = ws.GetHub()
	return nil
}

// Routes registers messenger routes.
func (h *Messenger) Routes(g *echo.Group) {
	// All routes require authentication
	g = g.Group("", middleware.RequireAuthentication)

	// Root redirect to first workspace or workspace list
	g.GET("/", h.RootRedirect).Name = routenames.MessengerRoot

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

// RootRedirect redirects to the first workspace or workspace list
func (h *Messenger) RootRedirect(ctx echo.Context) error {
	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get first workspace for user
	workspaces, err := h.orm.WorkspaceMember.
		Query().
		Where(workspacemember.UserIDEQ(int(user.ID))).
		QueryWorkspace().
		Order(ent.Desc(workspace.FieldCreatedAt)).
		Limit(1).
		All(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to fetch workspaces")
	}

	if len(workspaces) > 0 {
		// Redirect to first workspace
		return redirect.New(ctx).
			Route(routenames.MessengerWorkspaceView).
			Params(workspaces[0].ID).
			Go()
	}

	// No workspaces, redirect to workspace list
	return redirect.New(ctx).
		Route(routenames.MessengerWorkspaceList).
		Go()
}

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

	// Workspace membership is checked by RequireWorkspaceMember middleware
	// Render workspace page
	return messengerPages.Workspace(ctx)
}

// WorkspaceCreate creates a new workspace
func (h *Messenger) WorkspaceCreate(ctx echo.Context) error {
	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Parse form data
	name := ctx.FormValue("name")
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "workspace name is required")
	}

	slug := ctx.FormValue("slug")
	if slug == "" {
		// Generate slug from name if not provided
		slug = generateSlug(name)
	}

	description := ctx.FormValue("description")

	// Check if slug already exists
	exists, err := h.orm.Workspace.
		Query().
		Where(workspace.SlugEQ(slug)).
		Exist(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to check workspace slug")
	}

	if exists {
		return echo.NewHTTPError(http.StatusConflict, "workspace with this slug already exists")
	}

	workspaceEntity, err := h.orm.Workspace.
		Create().
		SetName(name).
		SetSlug(slug).
		SetDescription(description).
		SetOwnerID(int(user.ID)).
		Save(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to create workspace")
	}

	// Add creator as owner member
	_, err = h.orm.WorkspaceMember.
		Create().
		SetWorkspaceID(workspaceEntity.ID).
		SetUserID(int(user.ID)).
		SetRole(workspacemember.RoleOwner).
		Save(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to add workspace member")
	}

	return ctx.JSON(http.StatusCreated, workspaceEntity)
}

// generateSlug generates a URL-friendly slug from a name
func generateSlug(name string) string {
	// Simple slug generation - convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters (keep only alphanumeric and hyphens)
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// WorkspaceUpdate updates a workspace
func (h *Messenger) WorkspaceUpdate(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Check permissions (owner/admin only)
	if err := h.requireWorkspaceOwnerOrAdmin(ctx, id, int(user.ID)); err != nil {
		return err
	}

	// Get workspace (workspace is already loaded by LoadWorkspace middleware)
	workspaceEntity := ctx.Get(messengerMiddleware.WorkspaceKey).(*ent.Workspace)
	id = workspaceEntity.ID

	// Parse form data
	update := h.orm.Workspace.UpdateOneID(id)

	if name := ctx.FormValue("name"); name != "" {
		update = update.SetName(name)
	}

	if slug := ctx.FormValue("slug"); slug != "" {
		// Check if slug already exists (excluding current workspace)
		exists, err := h.orm.Workspace.
			Query().
			Where(workspace.SlugEQ(slug)).
			Where(workspace.IDNEQ(id)).
			Exist(ctx.Request().Context())

		if err != nil {
			return fail(err, "failed to check workspace slug")
		}

		if exists {
			return echo.NewHTTPError(http.StatusConflict, "workspace with this slug already exists")
		}

		update = update.SetSlug(slug)
	}

	if description := ctx.FormValue("description"); description != "" {
		update = update.SetDescription(description)
	}

	workspaceEntity, err = update.Save(ctx.Request().Context())
	if err != nil {
		return fail(err, "failed to update workspace")
	}

	// Send WebSocket event
	if h.hub != nil {
		event := &ws.Event{
			Type: ws.EventTypeChannelUpdated,
			Data: map[string]interface{}{
				"workspace_id": int64(id),
			},
		}
		// Send to all workspace members
		h.hub.SendToWorkspace(ctx.Request().Context(), int64(id), event.ToJSON())
	}

	return ctx.JSON(http.StatusOK, workspaceEntity)
}

// WorkspaceDelete deletes a workspace
func (h *Messenger) WorkspaceDelete(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Check permissions (owner only)
	if err := h.requireWorkspaceOwner(ctx, id, int(user.ID)); err != nil {
		return err
	}

	err = h.orm.Workspace.DeleteOneID(id).Exec(ctx.Request().Context())
	if err != nil {
		return fail(err, "failed to delete workspace")
	}

	return ctx.NoContent(http.StatusNoContent)
}

// WorkspaceAddMember adds a member to a workspace
func (h *Messenger) WorkspaceAddMember(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Check permissions (owner/admin only)
	if err := h.requireWorkspaceOwnerOrAdmin(ctx, id, int(user.ID)); err != nil {
		return err
	}

	// Get user ID from form
	userIDStr := ctx.FormValue("user_id")
	if userIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user_id is required")
	}

	targetUserID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}

	// Check if target user exists
	_, err = h.orm.User.Get(ctx.Request().Context(), targetUserID)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		return fail(err, "failed to get user")
	}

	// Check if user is already a member
	exists, err := h.orm.WorkspaceMember.
		Query().
		Where(workspacemember.WorkspaceIDEQ(id)).
		Where(workspacemember.UserIDEQ(targetUserID)).
		Exist(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to check membership")
	}

	if exists {
		return echo.NewHTTPError(http.StatusConflict, "user is already a member of this workspace")
	}

	// Get role from form (default to member)
	roleStr := ctx.FormValue("role")
	if roleStr == "" {
		roleStr = "member"
	}

	role := workspacemember.Role(roleStr)
	if role != workspacemember.RoleOwner && role != workspacemember.RoleAdmin && role != workspacemember.RoleMember {
		role = workspacemember.RoleMember
	}

	// Create workspace member
	member, err := h.orm.WorkspaceMember.
		Create().
		SetWorkspaceID(id).
		SetUserID(targetUserID).
		SetRole(role).
		Save(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to add workspace member")
	}

	return ctx.JSON(http.StatusCreated, member)
}

// WorkspaceRemoveMember removes a member from a workspace
func (h *Messenger) WorkspaceRemoveMember(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}

	userIDStr := ctx.Param("user_id")
	if userIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user_id is required")
	}

	targetUserID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Check permissions (owner/admin only, or user removing themselves)
	if targetUserID != int(user.ID) {
		if err := h.requireWorkspaceOwnerOrAdmin(ctx, id, int(user.ID)); err != nil {
			return err
		}
	}

	// Prevent removing the owner
	targetRole, err := h.getWorkspaceMemberRole(ctx, id, targetUserID)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "user is not a member of this workspace")
		}
		return fail(err, "failed to check member role")
	}

	if targetRole == workspacemember.RoleOwner {
		return echo.NewHTTPError(http.StatusForbidden, "cannot remove workspace owner")
	}

	// Delete workspace member
	_, err = h.orm.WorkspaceMember.
		Delete().
		Where(workspacemember.WorkspaceIDEQ(id)).
		Where(workspacemember.UserIDEQ(targetUserID)).
		Exec(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to remove workspace member")
	}

	return ctx.NoContent(http.StatusNoContent)
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

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get sidebar data
	activeChannelID := &id
	sidebarData, err := h.getSidebarData(ctx, ch.WorkspaceID, int(user.ID), activeChannelID, nil)
	if err != nil {
		return fail(err, "failed to load sidebar data")
	}

	// Get messages (last 50)
	messages, err := h.orm.Message.
		Query().
		Where(message.ChannelIDEQ(id)).
		Order(ent.Desc(message.FieldCreatedAt)).
		Limit(50).
		WithUser().
		All(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to load messages")
	}

	// Convert to MessageData
	messageData := make([]messengerComponents.MessageData, len(messages))
	for i, msg := range messages {
		// Get reactions
		reactions, err := h.orm.Reaction.
			Query().
			Where(reaction.MessageIDEQ(msg.ID)).
			WithUser().
			All(ctx.Request().Context())

		if err != nil {
			reactions = []*ent.Reaction{} // Continue with empty reactions
		}

		// Group reactions by emoji
		reactionMap := make(map[string]*messengerComponents.ReactionData)
		for _, r := range reactions {
			if existing, ok := reactionMap[r.Emoji]; ok {
				existing.Count++
				existing.UserIDs = append(existing.UserIDs, int64(r.UserID))
			} else {
				reactionMap[r.Emoji] = &messengerComponents.ReactionData{
					Emoji:   r.Emoji,
					Count:   1,
					UserIDs: []int64{int64(r.UserID)},
				}
			}
		}

		reactionData := make([]messengerComponents.ReactionData, 0, len(reactionMap))
		for _, r := range reactionMap {
			reactionData = append(reactionData, *r)
		}

		messageData[i] = messengerComponents.MessageData{
			ID:        int64(msg.ID),
			Content:   msg.Content,
			UserID:    int64(msg.UserID),
			UserName:  msg.Edges.User.Name,
			CreatedAt: msg.CreatedAt,
			EditedAt:  msg.EditedAt,
			Reactions: reactionData,
		}
	}

	// Reverse to show oldest first
	for i, j := 0, len(messageData)-1; i < j; i, j = i+1, j-1 {
		messageData[i], messageData[j] = messageData[j], messageData[i]
	}

	// Store sidebar data in context
	ctx.Set(context.MessengerSidebarKey, sidebarData)

	// Channel membership is checked by RequireChannelMember middleware
	return messengerPages.Channel(ctx, int64(ch.ID), ch.Name, messageData)
}

// ChannelCreate creates a new channel
func (h *Messenger) ChannelCreate(ctx echo.Context) error {
	workspaceID, err := strconv.Atoi(ctx.Param("workspace_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Verify user is a member of workspace
	exists, err := h.orm.WorkspaceMember.
		Query().
		Where(workspacemember.WorkspaceIDEQ(workspaceID)).
		Where(workspacemember.UserIDEQ(int(user.ID))).
		Exist(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to check workspace membership")
	}

	if !exists {
		return echo.NewHTTPError(http.StatusForbidden, "you are not a member of this workspace")
	}

	// Parse form data
	name := ctx.FormValue("name")
	if name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "channel name is required")
	}

	slug := ctx.FormValue("slug")
	if slug == "" {
		slug = generateSlug(name)
	}

	description := ctx.FormValue("description")
	isPrivate := ctx.FormValue("is_private") == "true" || ctx.FormValue("is_private") == "1"

	// Check if slug already exists in workspace
	exists, err = h.orm.Channel.
		Query().
		Where(channel.WorkspaceIDEQ(workspaceID)).
		Where(channel.SlugEQ(slug)).
		Exist(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to check channel slug")
	}

	if exists {
		return echo.NewHTTPError(http.StatusConflict, "channel with this slug already exists in this workspace")
	}

	ch, err := h.orm.Channel.
		Create().
		SetName(name).
		SetSlug(slug).
		SetDescription(description).
		SetIsPrivate(isPrivate).
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

	// Send WebSocket event
	if h.hub != nil {
		event := ws.ChannelUpdatedEvent(int64(ch.ID))
		h.hub.SendToChannel(int64(ch.ID), event.ToJSON())
	}

	return ctx.JSON(http.StatusCreated, ch)
}

// ChannelUpdate updates a channel
func (h *Messenger) ChannelUpdate(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid channel ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get channel
	ch, err := h.orm.Channel.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "channel not found")
		}
		return fail(err, "failed to get channel")
	}

	// Check if user is creator or workspace owner/admin
	if ch.CreatedBy != int(user.ID) {
		// Check workspace role
		if err := h.requireWorkspaceOwnerOrAdmin(ctx, ch.WorkspaceID, int(user.ID)); err != nil {
			return err
		}
	}

	// Parse form data
	update := h.orm.Channel.UpdateOneID(id)

	if name := ctx.FormValue("name"); name != "" {
		update = update.SetName(name)
	}

	if slug := ctx.FormValue("slug"); slug != "" {
		// Check if slug already exists in workspace (excluding current channel)
		exists, err := h.orm.Channel.
			Query().
			Where(channel.WorkspaceIDEQ(ch.WorkspaceID)).
			Where(channel.SlugEQ(slug)).
			Where(channel.IDNEQ(id)).
			Exist(ctx.Request().Context())

		if err != nil {
			return fail(err, "failed to check channel slug")
		}

		if exists {
			return echo.NewHTTPError(http.StatusConflict, "channel with this slug already exists in this workspace")
		}

		update = update.SetSlug(slug)
	}

	if description := ctx.FormValue("description"); description != "" {
		update = update.SetDescription(description)
	}

	if isPrivateStr := ctx.FormValue("is_private"); isPrivateStr != "" {
		isPrivate := isPrivateStr == "true" || isPrivateStr == "1"
		update = update.SetIsPrivate(isPrivate)
	}

	ch, err = update.Save(ctx.Request().Context())
	if err != nil {
		return fail(err, "failed to update channel")
	}

	// Send WebSocket event
	if h.hub != nil {
		event := ws.ChannelUpdatedEvent(int64(id))
		h.hub.SendToChannel(int64(id), event.ToJSON())
	}

	return ctx.JSON(http.StatusOK, ch)
}

// ChannelDelete deletes a channel
func (h *Messenger) ChannelDelete(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid channel ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get channel
	ch, err := h.orm.Channel.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "channel not found")
		}
		return fail(err, "failed to get channel")
	}

	// Check if user is creator or workspace owner/admin
	if ch.CreatedBy != int(user.ID) {
		// Check workspace role
		if err := h.requireWorkspaceOwnerOrAdmin(ctx, ch.WorkspaceID, int(user.ID)); err != nil {
			return err
		}
	}

	// Delete channel
	err = h.orm.Channel.DeleteOneID(id).Exec(ctx.Request().Context())
	if err != nil {
		return fail(err, "failed to delete channel")
	}

	return ctx.NoContent(http.StatusNoContent)
}

// ChannelAddMember adds a member to a channel
func (h *Messenger) ChannelAddMember(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid channel ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get channel
	ch, err := h.orm.Channel.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "channel not found")
		}
		return fail(err, "failed to get channel")
	}

	// Verify user is a member of workspace
	exists, err := h.orm.WorkspaceMember.
		Query().
		Where(workspacemember.WorkspaceIDEQ(ch.WorkspaceID)).
		Where(workspacemember.UserIDEQ(int(user.ID))).
		Exist(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to check workspace membership")
	}

	if !exists {
		return echo.NewHTTPError(http.StatusForbidden, "you are not a member of this workspace")
	}

	// Get user ID from form
	userIDStr := ctx.FormValue("user_id")
	if userIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user_id is required")
	}

	targetUserID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}

	// Check if target user is a workspace member
	exists, err = h.orm.WorkspaceMember.
		Query().
		Where(workspacemember.WorkspaceIDEQ(ch.WorkspaceID)).
		Where(workspacemember.UserIDEQ(targetUserID)).
		Exist(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to check workspace membership")
	}

	if !exists {
		return echo.NewHTTPError(http.StatusForbidden, "target user is not a member of this workspace")
	}

	// Check if user is already a channel member
	exists, err = h.orm.ChannelMember.
		Query().
		Where(channelmember.ChannelIDEQ(id)).
		Where(channelmember.UserIDEQ(targetUserID)).
		Exist(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to check channel membership")
	}

	if exists {
		return echo.NewHTTPError(http.StatusConflict, "user is already a member of this channel")
	}

	// Create channel member
	member, err := h.orm.ChannelMember.
		Create().
		SetChannelID(id).
		SetUserID(targetUserID).
		Save(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to add channel member")
	}

	// Send WebSocket event
	if h.hub != nil {
		event := ws.MemberJoinedEvent(int64(id), int64(targetUserID))
		h.hub.SendToChannel(int64(id), event.ToJSON())
	}

	return ctx.JSON(http.StatusCreated, member)
}

// ChannelRemoveMember removes a member from a channel
func (h *Messenger) ChannelRemoveMember(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid channel ID")
	}

	userIDStr := ctx.Param("user_id")
	if userIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user_id is required")
	}

	targetUserID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get channel
	ch, err := h.orm.Channel.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "channel not found")
		}
		return fail(err, "failed to get channel")
	}

	// Allow user to remove themselves, or require workspace owner/admin/creator
	if targetUserID != int(user.ID) {
		if ch.CreatedBy != int(user.ID) {
			if err := h.requireWorkspaceOwnerOrAdmin(ctx, ch.WorkspaceID, int(user.ID)); err != nil {
				return err
			}
		}
	}

	// Delete channel member
	_, err = h.orm.ChannelMember.
		Delete().
		Where(channelmember.ChannelIDEQ(id)).
		Where(channelmember.UserIDEQ(targetUserID)).
		Exec(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to remove channel member")
	}

	// Send WebSocket event
	if h.hub != nil {
		event := &ws.Event{
			Type: ws.EventTypeMemberLeft,
			Data: map[string]interface{}{
				"channel_id": int64(id),
				"user_id":    int64(targetUserID),
			},
		}
		h.hub.SendToChannel(int64(id), event.ToJSON())
	}

	return ctx.NoContent(http.StatusNoContent)
}

// ChannelMessages returns messages in a channel with pagination
func (h *Messenger) ChannelMessages(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid channel ID")
	}

	// Create pager (50 messages per page)
	pgr := pager.NewPager(ctx, 50)

	// Get total count
	total, err := h.orm.Message.
		Query().
		Where(message.ChannelIDEQ(id)).
		Count(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to count messages")
	}

	pgr.SetItems(total)

	// Get messages with pagination
	messages, err := h.orm.Message.
		Query().
		Where(message.ChannelIDEQ(id)).
		Order(ent.Desc(message.FieldCreatedAt)).
		Limit(pgr.ItemsPerPage).
		Offset(pgr.GetOffset()).
		All(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to fetch messages")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"messages": messages,
		"pager":    pgr,
	})
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

	// Parse form data
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

	// Send WebSocket event to channel members
	if h.hub != nil {
		event := ws.MessageNewEvent(int64(msg.ID), int64(channelID), int64(user.ID), content)
		h.hub.SendToChannel(int64(channelID), event.ToJSON())
	}

	return ctx.JSON(http.StatusCreated, msg)
}

// MessageUpdate updates a message
func (h *Messenger) MessageUpdate(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid message ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get message
	msg, err := h.orm.Message.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "message not found")
		}
		return fail(err, "failed to get message")
	}

	// Check if user owns the message
	if msg.UserID != int(user.ID) {
		return echo.NewHTTPError(http.StatusForbidden, "you can only edit your own messages")
	}

	// Parse form data
	content := ctx.FormValue("content")
	if content == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "message content is required")
	}

	// Update message
	msg, err = msg.Update().
		SetContent(content).
		SetEditedAt(time.Now()).
		Save(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to update message")
	}

	// Send WebSocket event
	if h.hub != nil {
		event := &ws.Event{
			Type: ws.EventTypeMessageEdited,
			Data: map[string]interface{}{
				"message_id": int64(msg.ID),
				"channel_id": int64(msg.ChannelID),
				"content":    content,
			},
		}
		h.hub.SendToChannel(int64(msg.ChannelID), event.ToJSON())
	}

	return ctx.JSON(http.StatusOK, msg)
}

// MessageDelete deletes a message
func (h *Messenger) MessageDelete(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid message ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get message
	msg, err := h.orm.Message.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "message not found")
		}
		return fail(err, "failed to get message")
	}

	// Check if user owns the message
	if msg.UserID != int(user.ID) {
		return echo.NewHTTPError(http.StatusForbidden, "you can only delete your own messages")
	}

	channelID := msg.ChannelID

	// Delete message
	err = h.orm.Message.DeleteOneID(id).Exec(ctx.Request().Context())
	if err != nil {
		return fail(err, "failed to delete message")
	}

	// Send WebSocket event
	if h.hub != nil {
		event := &ws.Event{
			Type: ws.EventTypeMessageDeleted,
			Data: map[string]interface{}{
				"message_id": int64(id),
				"channel_id": int64(channelID),
			},
		}
		h.hub.SendToChannel(int64(channelID), event.ToJSON())
	}

	return ctx.NoContent(http.StatusNoContent)
}

// MessageReplies returns replies to a message (thread)
func (h *Messenger) MessageReplies(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid message ID")
	}

	// Verify message exists
	_, err = h.orm.Message.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "message not found")
		}
		return fail(err, "failed to get message")
	}

	// Get replies (messages with thread_id = id)
	replies, err := h.orm.Message.
		Query().
		Where(message.ThreadIDEQ(id)).
		Order(ent.Asc(message.FieldCreatedAt)).
		All(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to fetch replies")
	}

	return ctx.JSON(http.StatusOK, replies)
}

// MessageReply creates a reply to a message (thread)
func (h *Messenger) MessageReply(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid message ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get parent message
	parentMsg, err := h.orm.Message.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "message not found")
		}
		return fail(err, "failed to get message")
	}

	// Parse form data
	content := ctx.FormValue("content")
	if content == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "message content is required")
	}

	// Create reply
	reply, err := h.orm.Message.
		Create().
		SetContent(content).
		SetMessageType(message.MessageTypeThreadReply).
		SetChannelID(parentMsg.ChannelID).
		SetUserID(int(user.ID)).
		SetThreadID(id).
		Save(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to create reply")
	}

	// Update reply count on parent message
	_, err = h.orm.Message.
		UpdateOneID(id).
		AddReplyCount(1).
		Save(ctx.Request().Context())

	if err != nil {
		// Log but don't fail
		log.Ctx(ctx).Warn("failed to update reply count", "message_id", id, "error", err)
	}

	// Send WebSocket event
	if h.hub != nil {
		event := &ws.Event{
			Type: ws.EventTypeMessageNew,
			Data: map[string]interface{}{
				"message_id": int64(reply.ID),
				"channel_id": int64(parentMsg.ChannelID),
				"thread_id":  int64(id),
				"user_id":    int64(user.ID),
				"content":    content,
			},
		}
		h.hub.SendToChannel(int64(parentMsg.ChannelID), event.ToJSON())
	}

	return ctx.JSON(http.StatusCreated, reply)
}

// ============================================================================
// Direct Message Handlers
// ============================================================================

// DMList returns a list of direct message conversations
func (h *Messenger) DMList(ctx echo.Context) error {
	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get all DMs where user is user1 or user2
	dms, err := h.orm.DirectMessage.
		Query().
		Where(
			directmessage.Or(
				directmessage.User1IDEQ(int(user.ID)),
				directmessage.User2IDEQ(int(user.ID)),
			),
		).
		Order(ent.Desc(directmessage.FieldLastMessageAt)).
		All(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to fetch direct messages")
	}

	return ctx.JSON(http.StatusOK, dms)
}

// DMView shows a direct message conversation
func (h *Messenger) DMView(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid DM ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get DM
	dm, err := h.orm.DirectMessage.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "direct message not found")
		}
		return fail(err, "failed to get direct message")
	}

	// Check if user is part of this DM
	if dm.User1ID != int(user.ID) && dm.User2ID != int(user.ID) {
		return echo.NewHTTPError(http.StatusForbidden, "you are not part of this conversation")
	}

	// Get other user
	var otherUserID int
	if dm.User1ID == int(user.ID) {
		otherUserID = dm.User2ID
	} else {
		otherUserID = dm.User1ID
	}

	otherUser, err := h.orm.User.Get(ctx.Request().Context(), otherUserID)
	if err != nil {
		return fail(err, "failed to get other user")
	}

	return messengerPages.DirectMessage(ctx, int64(id), otherUser.Name)
}

// DMCreate creates or retrieves a direct message conversation
func (h *Messenger) DMCreate(ctx echo.Context) error {
	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get target user ID from form
	userIDStr := ctx.FormValue("user_id")
	if userIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user_id is required")
	}

	targetUserID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}

	if targetUserID == int(user.ID) {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot create DM with yourself")
	}

	// Check if target user exists
	_, err = h.orm.User.Get(ctx.Request().Context(), targetUserID)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "target user not found")
		}
		return fail(err, "failed to get target user")
	}

	// Try to find existing DM (order user IDs to ensure consistency)
	user1ID := int(user.ID)
	user2ID := targetUserID
	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}

	dm, err := h.orm.DirectMessage.
		Query().
		Where(directmessage.User1IDEQ(user1ID)).
		Where(directmessage.User2IDEQ(user2ID)).
		Only(ctx.Request().Context())

	if err != nil {
		if ent.IsNotFound(err) {
			// Create new DM
			dm, err = h.orm.DirectMessage.
				Create().
				SetUser1ID(user1ID).
				SetUser2ID(user2ID).
				Save(ctx.Request().Context())

			if err != nil {
				return fail(err, "failed to create direct message")
			}
		} else {
			return fail(err, "failed to query direct message")
		}
	}

	return ctx.JSON(http.StatusOK, dm)
}

// DMMessages returns messages in a direct message conversation
func (h *Messenger) DMMessages(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid DM ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get DM and verify user is part of it
	dm, err := h.orm.DirectMessage.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "direct message not found")
		}
		return fail(err, "failed to get direct message")
	}

	if dm.User1ID != int(user.ID) && dm.User2ID != int(user.ID) {
		return echo.NewHTTPError(http.StatusForbidden, "you are not part of this conversation")
	}

	// Create pager (50 messages per page)
	pgr := pager.NewPager(ctx, 50)

	// Get total count
	total, err := h.orm.DirectMessageContent.
		Query().
		Where(directmessagecontent.DmIDEQ(id)).
		Count(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to count messages")
	}

	pgr.SetItems(total)

	// Get messages with pagination
	messages, err := h.orm.DirectMessageContent.
		Query().
		Where(directmessagecontent.DmIDEQ(id)).
		Order(ent.Desc(directmessagecontent.FieldCreatedAt)).
		Limit(pgr.ItemsPerPage).
		Offset(pgr.GetOffset()).
		All(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to fetch messages")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"messages": messages,
		"pager":    pgr,
	})
}

// DMMessageCreate creates a message in a direct message conversation
func (h *Messenger) DMMessageCreate(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid DM ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get DM and verify user is part of it
	dm, err := h.orm.DirectMessage.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "direct message not found")
		}
		return fail(err, "failed to get direct message")
	}

	if dm.User1ID != int(user.ID) && dm.User2ID != int(user.ID) {
		return echo.NewHTTPError(http.StatusForbidden, "you are not part of this conversation")
	}

	// Parse form data
	content := ctx.FormValue("content")
	if content == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "message content is required")
	}

	// Create message
	msg, err := h.orm.DirectMessageContent.
		Create().
		SetContent(content).
		SetDmID(id).
		SetUserID(int(user.ID)).
		Save(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to create message")
	}

	// Update last_message_at
	_, err = h.orm.DirectMessage.
		UpdateOneID(id).
		SetLastMessageAt(time.Now()).
		Save(ctx.Request().Context())

	if err != nil {
		// Log but don't fail
		log.Ctx(ctx).Warn("failed to update DM last_message_at", "dm_id", id, "error", err)
	}

	// Send WebSocket event to other user
	if h.hub != nil {
		otherUserID := dm.User1ID
		if dm.User1ID == int(user.ID) {
			otherUserID = dm.User2ID
		}

		event := &ws.Event{
			Type: ws.EventTypeMessageNew,
			Data: map[string]interface{}{
				"message_id": int64(msg.ID),
				"dm_id":      int64(id),
				"user_id":    int64(user.ID),
				"content":    content,
			},
		}
		h.hub.SendToUser(int64(otherUserID), event.ToJSON())
	}

	return ctx.JSON(http.StatusCreated, msg)
}

// ============================================================================
// Reaction Handlers
// ============================================================================

// ReactionAdd adds a reaction to a message
func (h *Messenger) ReactionAdd(ctx echo.Context) error {
	messageID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid message ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get emoji from form
	emoji := ctx.FormValue("emoji")
	if emoji == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "emoji is required")
	}

	// Check if message exists
	_, err = h.orm.Message.Get(ctx.Request().Context(), messageID)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "message not found")
		}
		return fail(err, "failed to get message")
	}

	// Check if reaction already exists
	exists, err := h.orm.Reaction.
		Query().
		Where(reaction.MessageIDEQ(messageID)).
		Where(reaction.UserIDEQ(int(user.ID))).
		Where(reaction.EmojiEQ(emoji)).
		Exist(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to check reaction")
	}

	if exists {
		return echo.NewHTTPError(http.StatusConflict, "reaction already exists")
	}

	// Create reaction
	reaction, err := h.orm.Reaction.
		Create().
		SetEmoji(emoji).
		SetMessageID(messageID).
		SetUserID(int(user.ID)).
		Save(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to create reaction")
	}

	// Send WebSocket event
	if h.hub != nil {
		msg, _ := h.orm.Message.Get(ctx.Request().Context(), messageID)
		event := ws.ReactionAddedEvent(int64(messageID), int64(user.ID), emoji)
		h.hub.SendToChannel(int64(msg.ChannelID), event.ToJSON())
	}

	return ctx.JSON(http.StatusCreated, reaction)
}

// ReactionRemove removes a reaction from a message
func (h *Messenger) ReactionRemove(ctx echo.Context) error {
	messageID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid message ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get emoji from URL parameter
	emoji := ctx.Param("emoji")
	if emoji == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "emoji is required")
	}

	// Get message to find channel ID
	msg, err := h.orm.Message.Get(ctx.Request().Context(), messageID)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "message not found")
		}
		return fail(err, "failed to get message")
	}

	// Find and delete reaction
	_, err = h.orm.Reaction.
		Delete().
		Where(reaction.MessageIDEQ(messageID)).
		Where(reaction.UserIDEQ(int(user.ID))).
		Where(reaction.EmojiEQ(emoji)).
		Exec(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to delete reaction")
	}

	// Send WebSocket event
	if h.hub != nil {
		event := &ws.Event{
			Type: ws.EventTypeReactionRemoved,
			Data: map[string]interface{}{
				"message_id": int64(messageID),
				"user_id":    int64(user.ID),
				"emoji":      emoji,
			},
		}
		h.hub.SendToChannel(int64(msg.ChannelID), event.ToJSON())
	}

	return ctx.NoContent(http.StatusNoContent)
}

// ============================================================================
// Attachment Handlers
// ============================================================================

// AttachmentUpload uploads a file attachment
func (h *Messenger) AttachmentUpload(ctx echo.Context) error {
	messageID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid message ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Verify message exists
	_, err = h.orm.Message.Get(ctx.Request().Context(), messageID)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "message not found")
		}
		return fail(err, "failed to get message")
	}

	// Get file from form
	file, err := ctx.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "file is required")
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return fail(err, "failed to open uploaded file")
	}
	defer src.Close()

	// Create file path in attachments directory
	filePath := filepath.Join("attachments", fmt.Sprintf("%d_%s", messageID, file.Filename))
	dst, err := h.files.Create(filePath)
	if err != nil {
		return fail(err, "failed to create file")
	}
	defer dst.Close()

	// Copy file content
	if _, err = io.Copy(dst, src); err != nil {
		return fail(err, "failed to save file")
	}

	// Get file info
	fileInfo, err := h.files.Stat(filePath)
	if err != nil {
		return fail(err, "failed to get file info")
	}

	// Create attachment record
	attachment, err := h.orm.Attachment.
		Create().
		SetFilename(file.Filename).
		SetFilepath(filePath).
		SetFileSize(fileInfo.Size()).
		SetMimeType(file.Header.Get("Content-Type")).
		SetMessageID(messageID).
		SetUploadedBy(int(user.ID)).
		Save(ctx.Request().Context())

	if err != nil {
		return fail(err, "failed to create attachment record")
	}

	return ctx.JSON(http.StatusCreated, attachment)
}

// AttachmentView serves an attachment file
func (h *Messenger) AttachmentView(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid attachment ID")
	}

	// Get attachment
	attachment, err := h.orm.Attachment.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "attachment not found")
		}
		return fail(err, "failed to get attachment")
	}

	// Open file
	file, err := h.files.Open(attachment.Filepath)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "file not found on disk")
	}
	defer file.Close()

	// Set headers
	ctx.Response().Header().Set("Content-Type", attachment.MimeType)
	ctx.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachment.Filename))

	// Stream file
	_, err = io.Copy(ctx.Response(), file)
	return err
}

// AttachmentDelete deletes an attachment
func (h *Messenger) AttachmentDelete(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid attachment ID")
	}

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get attachment
	attachment, err := h.orm.Attachment.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound, "attachment not found")
		}
		return fail(err, "failed to get attachment")
	}

	// Check if user uploaded the attachment
	if attachment.UploadedBy != int(user.ID) {
		return echo.NewHTTPError(http.StatusForbidden, "you can only delete your own attachments")
	}

	// Delete file from filesystem
	if err := h.files.Remove(attachment.Filepath); err != nil {
		// Log but don't fail - file might already be deleted
		log.Ctx(ctx).Warn("failed to delete attachment file", "attachment_id", id, "filepath", attachment.Filepath, "error", err)
	}

	// Delete attachment record
	err = h.orm.Attachment.DeleteOneID(id).Exec(ctx.Request().Context())
	if err != nil {
		return fail(err, "failed to delete attachment")
	}

	return ctx.NoContent(http.StatusNoContent)
}

// ============================================================================
// CUSTOM CODE END
// ============================================================================
