package handlers

// ============================================================================
// Slack Messenger Handlers
// ============================================================================
// This file contains HTTP handlers for messenger functionality.
// ============================================================================

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/config"
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
	"github.com/mikestefanello/pagoda/pkg/htmx"
	"github.com/mikestefanello/pagoda/pkg/log"
	messengerMiddleware "github.com/mikestefanello/pagoda/pkg/middleware"
	"github.com/mikestefanello/pagoda/pkg/pager"
	"github.com/mikestefanello/pagoda/pkg/redirect"
	"github.com/mikestefanello/pagoda/pkg/routenames"
	"github.com/mikestefanello/pagoda/pkg/services"
	"github.com/mikestefanello/pagoda/pkg/ui"
	messengerComponents "github.com/mikestefanello/pagoda/pkg/ui/components/messenger"
	messengerPages "github.com/mikestefanello/pagoda/pkg/ui/pages/messenger"
	ws "github.com/mikestefanello/pagoda/pkg/websocket"
	"github.com/spf13/afero"
)

// getWorkspaceMemberRole возвращает роль пользователя в workspace
// Используется для проверки прав доступа (owner, admin, member)
// Параметры:
//   - ctx: контекст Echo с запросом
//   - workspaceID: ID workspace
//   - userID: ID пользователя
//
// Возвращает:
//   - Role: роль пользователя (owner, admin, member)
//   - error: ошибка если пользователь не найден или не является членом workspace
func (h *Messenger) getWorkspaceMemberRole(ctx echo.Context, workspaceID, userID int) (workspacemember.Role, error) {
	logger := log.Ctx(ctx)
	logger.Debug("Getting workspace member role", "workspace_id", workspaceID, "user_id", userID)

	// Ищем запись WorkspaceMember в базе данных
	// WorkspaceMember связывает пользователя с workspace и хранит его роль
	member, err := h.orm.WorkspaceMember.
		Query().                                           // Начинаем запрос
		Where(workspacemember.WorkspaceIDEQ(workspaceID)). // Фильтр: workspace_id = workspaceID
		Where(workspacemember.UserIDEQ(userID)).           // Фильтр: user_id = userID
		Only(ctx.Request().Context())                      // Only возвращает одну запись или ошибку если не найдено/найдено несколько

	if err != nil {
		// Если запись не найдена, пользователь не является членом workspace
		logger.Debug("Failed to get workspace member role", "workspace_id", workspaceID, "user_id", userID, "error", err)
		return "", err
	}

	logger.Debug("Workspace member role retrieved", "workspace_id", workspaceID, "user_id", userID, "role", member.Role)
	return member.Role, nil
}

// requireWorkspaceOwnerOrAdmin проверяет, является ли пользователь владельцем или администратором workspace
// Используется для защиты действий, которые могут выполнять только owner или admin
// Например: удаление workspace, изменение настроек, добавление/удаление членов
// Параметры:
//   - ctx: контекст Echo с запросом
//   - workspaceID: ID workspace
//   - userID: ID пользователя для проверки
//
// Возвращает:
//   - error: nil если пользователь owner или admin, иначе HTTP 403 Forbidden
func (h *Messenger) requireWorkspaceOwnerOrAdmin(ctx echo.Context, workspaceID, userID int) error {
	logger := log.Ctx(ctx)
	logger.Debug("Checking workspace owner/admin permission", "workspace_id", workspaceID, "user_id", userID)

	// Получаем роль пользователя в workspace
	role, err := h.getWorkspaceMemberRole(ctx, workspaceID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			// Пользователь не является членом workspace
			logger.Warn("Permission check failed: user not a workspace member", "workspace_id", workspaceID, "user_id", userID)
			return echo.NewHTTPError(http.StatusForbidden, "you are not a member of this workspace")
		}
		// Другая ошибка (например, проблема с базой данных)
		logger.Error("Failed to check workspace role", "error", err, "workspace_id", workspaceID, "user_id", userID)
		return fail(err, "failed to check workspace role")
	}

	// Проверяем, что роль - owner или admin
	if role != workspacemember.RoleOwner && role != workspacemember.RoleAdmin {
		logger.Warn("Permission check failed: user is not owner or admin", "workspace_id", workspaceID, "user_id", userID, "role", role)
		return echo.NewHTTPError(http.StatusForbidden, "only owners and admins can perform this action")
	}

	logger.Debug("Permission check passed: user is owner or admin", "workspace_id", workspaceID, "user_id", userID, "role", role)
	return nil
}

// requireWorkspaceOwner checks if user is owner of workspace
func (h *Messenger) requireWorkspaceOwner(ctx echo.Context, workspaceID, userID int) error {
	logger := log.Ctx(ctx)
	logger.Debug("Checking workspace owner permission", "workspace_id", workspaceID, "user_id", userID)

	role, err := h.getWorkspaceMemberRole(ctx, workspaceID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warn("Permission check failed: user not a workspace member", "workspace_id", workspaceID, "user_id", userID)
			return echo.NewHTTPError(http.StatusForbidden, "you are not a member of this workspace")
		}
		logger.Error("Failed to check workspace role", "error", err, "workspace_id", workspaceID, "user_id", userID)
		return fail(err, "failed to check workspace role")
	}

	if role != workspacemember.RoleOwner {
		logger.Warn("Permission check failed: user is not workspace owner", "workspace_id", workspaceID, "user_id", userID, "role", role)
		return echo.NewHTTPError(http.StatusForbidden, "only workspace owner can perform this action")
	}

	logger.Debug("Permission check passed: user is workspace owner", "workspace_id", workspaceID, "user_id", userID)
	return nil
}

// getSidebarData загружает данные для боковой панели (sidebar) workspace
// Sidebar содержит список каналов и прямых сообщений пользователя
// Параметры:
//   - ctx: контекст Echo с запросом
//   - workspaceID: ID workspace для загрузки каналов
//   - userID: ID пользователя (для фильтрации каналов и DM, где он участник)
//   - activeChannelID: указатель на ID активного канала (nil если нет активного канала)
//   - activeDMID: указатель на ID активного DM (nil если нет активного DM)
//
// Возвращает:
//   - SidebarData: структура с данными для sidebar (каналы, DM, активные элементы)
//   - error: ошибка если не удалось загрузить данные
func (h *Messenger) getSidebarData(ctx echo.Context, workspaceID int, userID int, activeChannelID *int, activeDMID *int) (messengerComponents.SidebarData, error) {
	logger := log.Ctx(ctx)
	logger.Info("=== GET SIDEBAR DATA START ===", "workspace_id", workspaceID, "user_id", userID)

	// Загружаем workspace для получения его имени
	logger.Info("Loading workspace", "workspace_id", workspaceID)
	ws, err := h.orm.Workspace.Get(ctx.Request().Context(), workspaceID)
	if err != nil {
		logger.Error("Failed to load workspace", "error", err, "workspace_id", workspaceID)
		return messengerComponents.SidebarData{}, err
	}
	logger.Info("Workspace loaded", "workspace_id", ws.ID, "workspace_name", ws.Name)

	// Загружаем каналы, где пользователь является участником
	// Используем ChannelMember для фильтрации - показываем только каналы, где пользователь есть в членах
	logger.Info("Loading channels for workspace", "workspace_id", workspaceID, "user_id", userID)
	channels, err := h.orm.ChannelMember.
		Query().                                   // Начинаем запрос с ChannelMember
		Where(channelmember.UserIDEQ(userID)).     // Фильтр: только записи где user_id = userID
		QueryChannel().                            // Переходим к связанным каналам (через edge ChannelMember -> Channel)
		Where(channel.WorkspaceIDEQ(workspaceID)). // Фильтр: только каналы в указанном workspace
		All(ctx.Request().Context())               // Получаем все найденные каналы

	if err != nil {
		logger.Error("Failed to load channels", "error", err, "workspace_id", workspaceID, "user_id", userID)
		return messengerComponents.SidebarData{}, err
	}
	logger.Info("Channels loaded", "workspace_id", workspaceID, "channels_count", len(channels))

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
	logger.Info("Loading direct messages", "user_id", userID)
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
		logger.Error("Failed to load direct messages", "error", err, "user_id", userID)
		return messengerComponents.SidebarData{}, err
	}
	logger.Info("Direct messages loaded", "user_id", userID, "dms_count", len(dms))

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

	logger.Info("Sidebar data prepared", "channels_count", len(channelData), "dms_count", len(dmData), "active_channel_id", activeChID, "active_dm_id", activeDMID64)
	logger.Info("=== GET SIDEBAR DATA END ===")
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
	orm    *ent.Client
	hub    *ws.Hub // WebSocket hub for real-time events
	files  afero.Fs
	config *config.Config
}

func init() {
	Register(new(Messenger))
}

// Init initializes the handler with dependencies from the container.
func (h *Messenger) Init(c *services.Container) error {
	h.orm = c.ORM
	h.files = c.Files
	h.config = c.Config
	// Get hub from WebSocket handler (will be set when WebSocket handler initializes)
	h.hub = ws.GetHub()
	return nil
}

// Routes registers messenger routes.
func (h *Messenger) Routes(g *echo.Group) {
	// Root redirect to first workspace or workspace list (accessible without auth to redirect to login)
	g.GET("/", h.RootRedirect).Name = routenames.MessengerRoot

	// All other routes require authentication
	g = g.Group("", messengerMiddleware.RequireAuthentication)

	// Workspace routes
	g.GET("/workspace", h.WorkspaceList).Name = routenames.MessengerWorkspaceList
	g.GET("/workspace/:id", h.WorkspaceView, messengerMiddleware.LoadWorkspace(h.orm)).Name = routenames.MessengerWorkspaceView
	g.GET("/workspace/create/form", h.WorkspaceCreateForm).Name = routenames.MessengerWorkspaceCreateForm
	g.POST("/workspace", h.WorkspaceCreate).Name = routenames.MessengerWorkspaceCreate
	g.PUT("/workspace/:id", h.WorkspaceUpdate, messengerMiddleware.LoadWorkspace(h.orm), messengerMiddleware.RequireWorkspaceMember(h.orm)).Name = routenames.MessengerWorkspaceUpdate
	g.DELETE("/workspace/:id", h.WorkspaceDelete, messengerMiddleware.LoadWorkspace(h.orm), messengerMiddleware.RequireWorkspaceMember(h.orm)).Name = routenames.MessengerWorkspaceDelete
	g.POST("/workspace/:id/members", h.WorkspaceAddMember, messengerMiddleware.LoadWorkspace(h.orm), messengerMiddleware.RequireWorkspaceMember(h.orm)).Name = routenames.MessengerWorkspaceAddMember
	g.DELETE("/workspace/:id/members/:user_id", h.WorkspaceRemoveMember, messengerMiddleware.LoadWorkspace(h.orm), messengerMiddleware.RequireWorkspaceMember(h.orm)).Name = routenames.MessengerWorkspaceRemoveMember

	// Channel routes
	g.GET("/workspace/:workspace_id/channels", h.ChannelList).Name = routenames.MessengerChannelList
	g.GET("/channel/:id", h.ChannelView, messengerMiddleware.LoadChannel(h.orm), messengerMiddleware.RequireChannelMember(h.orm)).Name = routenames.MessengerChannelView
	g.GET("/workspace/:workspace_id/channels/create/form", h.ChannelCreateForm).Name = routenames.MessengerChannelCreateForm
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
	g.GET("/message/:id/thread", h.MessageThread).Name = routenames.MessengerMessageThread

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
	g.POST("/dm/:id/attachments", h.DMAttachmentUpload).Name = routenames.MessengerDMAttachmentUpload
	g.GET("/attachment/:id", h.AttachmentView).Name = routenames.MessengerAttachmentView
	g.DELETE("/attachment/:id", h.AttachmentDelete).Name = routenames.MessengerAttachmentDelete
}

// ============================================================================
// Workspace Handlers
// ============================================================================

// RootRedirect redirects to the first workspace or workspace list
func (h *Messenger) RootRedirect(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== ROOT REDIRECT START ===")
	logger.Info("Request URL", "url", ctx.Request().URL.String())
	logger.Info("Request method", "method", ctx.Request().Method)

	// Check if user is authenticated
	userInterface := ctx.Get(context.AuthenticatedUserKey)
	if userInterface == nil {
		// Not authenticated, redirect to login
		logger.Info("User not authenticated, redirecting to login")
		return ctx.Redirect(http.StatusFound, "/user/login")
	}

	logger.Info("User is authenticated", "user", userInterface)
	user := userInterface.(*ent.User)
	logger.Info("User loaded", "user_id", user.ID, "user_name", user.Name, "user_email", user.Email)

	// Get first workspace for user
	logger.Info("Fetching user workspaces...")
	workspaces, err := h.orm.WorkspaceMember.
		Query().
		Where(workspacemember.UserIDEQ(int(user.ID))).
		QueryWorkspace().
		Order(ent.Desc(workspace.FieldCreatedAt)).
		Limit(1).
		All(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to fetch workspaces", "error", err)
		return fail(err, "failed to fetch workspaces")
	}

	logger.Info("Workspaces fetched", "count", len(workspaces))

	if len(workspaces) == 0 {
		// No workspaces - show workspace creation page
		logger.Info("No workspaces found, showing workspace creation page")
		logger.Info("=== ROOT REDIRECT END ===")
		return h.WorkspaceCreatePage(ctx)
	} else if len(workspaces) == 1 {
		// One workspace - redirect to it automatically
		redirectURL := ctx.Echo().Reverse(routenames.MessengerWorkspaceView, workspaces[0].ID)
		logger.Info("Redirecting to first workspace", "workspace_id", workspaces[0].ID, "redirect_url", redirectURL)
		logger.Info("=== ROOT REDIRECT END ===")
		return redirect.New(ctx).
			Route(routenames.MessengerWorkspaceView).
			Params(workspaces[0].ID).
			Go()
	} else {
		// Multiple workspaces - show selection page
		logger.Info("Multiple workspaces found, showing selection page", "count", len(workspaces))
		logger.Info("=== ROOT REDIRECT END ===")
		return h.WorkspaceSelectPage(ctx)
	}
}

// WorkspaceList returns a list of workspaces for the authenticated user
// If user has no workspaces, redirects to creation page
// If user has one workspace, redirects to that workspace
// If user has multiple workspaces, shows selection page
func (h *Messenger) WorkspaceList(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== WORKSPACE LIST START ===")

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("Loading workspaces for user", "user_id", user.ID, "user_email", user.Email)

	workspaces, err := h.orm.WorkspaceMember.
		Query().
		Where(workspacemember.UserIDEQ(int(user.ID))).
		QueryWorkspace().
		Order(ent.Desc(workspace.FieldCreatedAt)).
		All(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to fetch workspaces", "error", err, "user_id", user.ID)
		return fail(err, "failed to fetch workspaces")
	}

	logger.Info("Workspaces loaded", "count", len(workspaces), "user_id", user.ID)

	// Redirect based on workspace count
	if len(workspaces) == 0 {
		// No workspaces - redirect to creation page
		logger.Info("No workspaces found, redirecting to creation page")
		logger.Info("=== WORKSPACE LIST END ===")
		return h.WorkspaceCreatePage(ctx)
	} else if len(workspaces) == 1 {
		// One workspace - redirect to it
		logger.Info("One workspace found, redirecting to it", "workspace_id", workspaces[0].ID)
		logger.Info("=== WORKSPACE LIST END ===")
		return redirect.New(ctx).
			Route(routenames.MessengerWorkspaceView).
			Params(workspaces[0].ID).
			Go()
	} else {
		// Multiple workspaces - show selection page
		logger.Info("Multiple workspaces found, showing selection page", "count", len(workspaces))
		logger.Info("=== WORKSPACE LIST END ===")
		return h.WorkspaceSelectPage(ctx)
	}
}

// WorkspaceCreatePage renders a page for creating the first workspace
func (h *Messenger) WorkspaceCreatePage(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== WORKSPACE CREATE PAGE START ===")

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("Rendering workspace creation page", "user_id", user.ID)

	// Empty sidebar data
	sidebarData := messengerComponents.SidebarData{
		Channels:       []messengerComponents.ChannelData{},
		DirectMessages: []messengerComponents.DMData{},
	}
	ctx.Set(context.MessengerSidebarKey, sidebarData)

	logger.Info("=== WORKSPACE CREATE PAGE END ===")
	return messengerPages.WorkspaceCreate(ctx)
}

// WorkspaceSelectPage renders a page for selecting a workspace when user has multiple workspaces
func (h *Messenger) WorkspaceSelectPage(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== WORKSPACE SELECT PAGE START ===")

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("Loading workspaces for selection", "user_id", user.ID)

	workspaces, err := h.orm.WorkspaceMember.
		Query().
		Where(workspacemember.UserIDEQ(int(user.ID))).
		QueryWorkspace().
		Order(ent.Desc(workspace.FieldCreatedAt)).
		All(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to fetch workspaces", "error", err, "user_id", user.ID)
		return fail(err, "failed to fetch workspaces")
	}

	logger.Info("Workspaces loaded for selection", "count", len(workspaces), "user_id", user.ID)

	// Empty sidebar data
	sidebarData := messengerComponents.SidebarData{
		Channels:       []messengerComponents.ChannelData{},
		DirectMessages: []messengerComponents.DMData{},
	}
	ctx.Set(context.MessengerSidebarKey, sidebarData)

	logger.Info("=== WORKSPACE SELECT PAGE END ===")
	return messengerPages.WorkspaceSelect(ctx, workspaces)
}

// WorkspaceView shows a workspace and redirects to the workspace page
func (h *Messenger) WorkspaceView(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== WORKSPACE VIEW START ===")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logger.Error("Invalid workspace ID", "error", err, "id_param", ctx.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}
	logger.Info("Loading workspace", "workspace_id", id)

	workspace, err := h.orm.Workspace.Get(ctx.Request().Context(), id)
	if err != nil {
		logger.Error("Workspace not found", "error", err, "workspace_id", id)
		return echo.NewHTTPError(http.StatusNotFound, "workspace not found")
	}
	logger.Info("Workspace loaded", "workspace_id", workspace.ID, "workspace_name", workspace.Name)

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("Loading sidebar data", "workspace_id", id, "user_id", user.ID)

	// Get sidebar data
	sidebarData, err := h.getSidebarData(ctx, id, int(user.ID), nil, nil)
	if err != nil {
		logger.Error("Failed to load sidebar data", "error", err, "workspace_id", id, "user_id", user.ID)
		return fail(err, "failed to load sidebar data")
	}
	logger.Info("Sidebar data loaded", "channels_count", len(sidebarData.Channels), "dms_count", len(sidebarData.DirectMessages))

	// Store sidebar data in context
	ctx.Set(context.MessengerSidebarKey, sidebarData)

	// Workspace membership is checked by RequireWorkspaceMember middleware
	// Render workspace page
	logger.Info("=== WORKSPACE VIEW END ===")
	return messengerPages.Workspace(ctx)
}

// WorkspaceCreate creates a new workspace
func (h *Messenger) WorkspaceCreate(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== WORKSPACE CREATE START ===")

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("Creating workspace", "user_id", user.ID, "user_email", user.Email)

	// Parse form data
	name := ctx.FormValue("name")
	if name == "" {
		logger.Warn("Workspace creation failed: name is required")
		return echo.NewHTTPError(http.StatusBadRequest, "workspace name is required")
	}

	slug := ctx.FormValue("slug")
	if slug == "" {
		// Generate slug from name if not provided
		slug = generateSlug(name)
		logger.Info("Generated slug from name", "name", name, "slug", slug)
	}

	description := ctx.FormValue("description")
	logger.Info("Workspace form data", "name", name, "slug", slug, "description_length", len(description))

	// Check if slug already exists
	exists, err := h.orm.Workspace.
		Query().
		Where(workspace.SlugEQ(slug)).
		Exist(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to check workspace slug", "error", err, "slug", slug)
		return fail(err, "failed to check workspace slug")
	}

	if exists {
		logger.Warn("Workspace creation failed: slug already exists", "slug", slug)
		return echo.NewHTTPError(http.StatusConflict, "workspace with this slug already exists")
	}

	logger.Info("Creating workspace entity", "name", name, "slug", slug)
	workspaceEntity, err := h.orm.Workspace.
		Create().
		SetName(name).
		SetSlug(slug).
		SetDescription(description).
		SetOwnerID(int(user.ID)).
		Save(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to create workspace", "error", err, "name", name, "slug", slug)
		return fail(err, "failed to create workspace")
	}
	logger.Info("Workspace created", "workspace_id", workspaceEntity.ID, "workspace_name", workspaceEntity.Name)

	// Add creator as owner member
	logger.Info("Adding creator as workspace owner", "workspace_id", workspaceEntity.ID, "user_id", user.ID)
	_, err = h.orm.WorkspaceMember.
		Create().
		SetWorkspaceID(workspaceEntity.ID).
		SetUserID(int(user.ID)).
		SetRole(workspacemember.RoleOwner).
		Save(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to add workspace member", "error", err, "workspace_id", workspaceEntity.ID, "user_id", user.ID)
		return fail(err, "failed to add workspace member")
	}
	logger.Info("Creator added as workspace owner", "workspace_id", workspaceEntity.ID, "user_id", user.ID)

	// If HTMX request, close modal and redirect
	if htmx.GetRequest(ctx).Enabled {
		redirectURL := ctx.Echo().Reverse(routenames.MessengerWorkspaceView, workspaceEntity.ID)
		logger.Info("HTMX request detected, redirecting", "redirect_url", redirectURL, "workspace_id", workspaceEntity.ID)
		htmx.Response{Redirect: redirectURL}.Apply(ctx)
		logger.Info("=== WORKSPACE CREATE END ===")
		return ctx.NoContent(http.StatusOK)
	}

	logger.Info("=== WORKSPACE CREATE END ===")
	return ctx.JSON(http.StatusCreated, workspaceEntity)
}

// WorkspaceCreateForm renders the workspace creation form modal
func (h *Messenger) WorkspaceCreateForm(ctx echo.Context) error {
	r := ui.NewRequest(ctx)
	modal := messengerComponents.WorkspaceCreateModal(r, nil)
	var buf bytes.Buffer
	if err := modal.Render(&buf); err != nil {
		return fail(err, "failed to render modal")
	}
	return ctx.HTML(http.StatusOK, buf.String())
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
	logger := log.Ctx(ctx)
	logger.Info("=== WORKSPACE UPDATE START ===")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logger.Error("Invalid workspace ID", "error", err, "id_param", ctx.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}
	logger.Info("Updating workspace", "workspace_id", id)

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("Checking permissions", "workspace_id", id, "user_id", user.ID)

	// Check permissions (owner/admin only)
	if err := h.requireWorkspaceOwnerOrAdmin(ctx, id, int(user.ID)); err != nil {
		logger.Warn("Permission check failed", "workspace_id", id, "user_id", user.ID, "error", err)
		return err
	}
	logger.Info("Permissions verified", "workspace_id", id, "user_id", user.ID)

	// Get workspace (workspace is already loaded by LoadWorkspace middleware)
	workspaceEntity := ctx.Get(messengerMiddleware.WorkspaceKey).(*ent.Workspace)
	id = workspaceEntity.ID
	logger.Info("Workspace loaded from context", "workspace_id", workspaceEntity.ID, "workspace_name", workspaceEntity.Name)

	// Parse form data
	update := h.orm.Workspace.UpdateOneID(id)
	hasUpdates := false

	if name := ctx.FormValue("name"); name != "" {
		logger.Info("Updating workspace name", "workspace_id", id, "new_name", name)
		update = update.SetName(name)
		hasUpdates = true
	}

	if slug := ctx.FormValue("slug"); slug != "" {
		logger.Info("Checking workspace slug availability", "workspace_id", id, "new_slug", slug)
		// Check if slug already exists (excluding current workspace)
		exists, err := h.orm.Workspace.
			Query().
			Where(workspace.SlugEQ(slug)).
			Where(workspace.IDNEQ(id)).
			Exist(ctx.Request().Context())

		if err != nil {
			logger.Error("Failed to check workspace slug", "error", err, "workspace_id", id, "slug", slug)
			return fail(err, "failed to check workspace slug")
		}

		if exists {
			logger.Warn("Workspace update failed: slug already exists", "workspace_id", id, "slug", slug)
			return echo.NewHTTPError(http.StatusConflict, "workspace with this slug already exists")
		}

		logger.Info("Updating workspace slug", "workspace_id", id, "new_slug", slug)
		update = update.SetSlug(slug)
		hasUpdates = true
	}

	if description := ctx.FormValue("description"); description != "" {
		logger.Info("Updating workspace description", "workspace_id", id, "description_length", len(description))
		update = update.SetDescription(description)
		hasUpdates = true
	}

	if !hasUpdates {
		logger.Warn("Workspace update: no fields to update", "workspace_id", id)
		return ctx.JSON(http.StatusOK, workspaceEntity)
	}

	logger.Info("Saving workspace updates", "workspace_id", id)
	workspaceEntity, err = update.Save(ctx.Request().Context())
	if err != nil {
		logger.Error("Failed to update workspace", "error", err, "workspace_id", id)
		return fail(err, "failed to update workspace")
	}
	logger.Info("Workspace updated successfully", "workspace_id", workspaceEntity.ID, "workspace_name", workspaceEntity.Name)

	// Send WebSocket event
	if h.hub != nil {
		logger.Info("Sending WebSocket event: workspace updated", "workspace_id", id)
		event := &ws.Event{
			Type: ws.EventTypeChannelUpdated,
			Data: map[string]interface{}{
				"workspace_id": int64(id),
			},
		}
		// Send to all workspace members
		h.hub.SendToWorkspace(ctx.Request().Context(), int64(id), event.ToJSON())
	}

	logger.Info("=== WORKSPACE UPDATE END ===")
	return ctx.JSON(http.StatusOK, workspaceEntity)
}

// WorkspaceDelete deletes a workspace
func (h *Messenger) WorkspaceDelete(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== WORKSPACE DELETE START ===")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logger.Error("Invalid workspace ID", "error", err, "id_param", ctx.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}
	logger.Info("Deleting workspace", "workspace_id", id)

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("Checking permissions (owner only)", "workspace_id", id, "user_id", user.ID)

	// Check permissions (owner only)
	if err := h.requireWorkspaceOwner(ctx, id, int(user.ID)); err != nil {
		logger.Warn("Permission check failed: not workspace owner", "workspace_id", id, "user_id", user.ID, "error", err)
		return err
	}
	logger.Info("Permissions verified: user is workspace owner", "workspace_id", id, "user_id", user.ID)

	logger.Info("Deleting workspace from database", "workspace_id", id)
	err = h.orm.Workspace.DeleteOneID(id).Exec(ctx.Request().Context())
	if err != nil {
		logger.Error("Failed to delete workspace", "error", err, "workspace_id", id)
		return fail(err, "failed to delete workspace")
	}
	logger.Info("Workspace deleted successfully", "workspace_id", id)
	logger.Info("=== WORKSPACE DELETE END ===")

	return ctx.NoContent(http.StatusNoContent)
}

// WorkspaceAddMember adds a member to a workspace
func (h *Messenger) WorkspaceAddMember(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== WORKSPACE ADD MEMBER START ===")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logger.Error("Invalid workspace ID", "error", err, "id_param", ctx.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}
	logger.Info("Adding member to workspace", "workspace_id", id)

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("Checking permissions (owner/admin only)", "workspace_id", id, "user_id", user.ID)

	// Check permissions (owner/admin only)
	if err := h.requireWorkspaceOwnerOrAdmin(ctx, id, int(user.ID)); err != nil {
		logger.Warn("Permission check failed", "workspace_id", id, "user_id", user.ID, "error", err)
		return err
	}
	logger.Info("Permissions verified", "workspace_id", id, "user_id", user.ID)

	// Get user ID from form
	userIDStr := ctx.FormValue("user_id")
	if userIDStr == "" {
		logger.Warn("Workspace add member failed: user_id is required", "workspace_id", id)
		return echo.NewHTTPError(http.StatusBadRequest, "user_id is required")
	}

	targetUserID, err := strconv.Atoi(userIDStr)
	if err != nil {
		logger.Error("Invalid user_id", "error", err, "user_id_str", userIDStr)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}
	logger.Info("Target user ID", "workspace_id", id, "target_user_id", targetUserID)

	// Check if target user exists
	logger.Info("Checking if target user exists", "target_user_id", targetUserID)
	_, err = h.orm.User.Get(ctx.Request().Context(), targetUserID)
	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warn("Workspace add member failed: user not found", "workspace_id", id, "target_user_id", targetUserID)
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		logger.Error("Failed to get user", "error", err, "target_user_id", targetUserID)
		return fail(err, "failed to get user")
	}
	logger.Info("Target user found", "target_user_id", targetUserID)

	// Check if user is already a member
	logger.Info("Checking if user is already a workspace member", "workspace_id", id, "target_user_id", targetUserID)
	exists, err := h.orm.WorkspaceMember.
		Query().
		Where(workspacemember.WorkspaceIDEQ(id)).
		Where(workspacemember.UserIDEQ(targetUserID)).
		Exist(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to check membership", "error", err, "workspace_id", id, "target_user_id", targetUserID)
		return fail(err, "failed to check membership")
	}

	if exists {
		logger.Warn("Workspace add member failed: user already a member", "workspace_id", id, "target_user_id", targetUserID)
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
	logger.Info("Adding member with role", "workspace_id", id, "target_user_id", targetUserID, "role", role)

	// Create workspace member
	member, err := h.orm.WorkspaceMember.
		Create().
		SetWorkspaceID(id).
		SetUserID(targetUserID).
		SetRole(role).
		Save(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to add workspace member", "error", err, "workspace_id", id, "target_user_id", targetUserID, "role", role)
		return fail(err, "failed to add workspace member")
	}
	logger.Info("Member added to workspace", "workspace_id", id, "member_id", member.ID, "target_user_id", targetUserID, "role", role)
	logger.Info("=== WORKSPACE ADD MEMBER END ===")

	return ctx.JSON(http.StatusCreated, member)
}

// WorkspaceRemoveMember removes a member from a workspace
func (h *Messenger) WorkspaceRemoveMember(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== WORKSPACE REMOVE MEMBER START ===")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logger.Error("Invalid workspace ID", "error", err, "id_param", ctx.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}

	userIDStr := ctx.Param("user_id")
	if userIDStr == "" {
		logger.Warn("Workspace remove member failed: user_id is required", "workspace_id", id)
		return echo.NewHTTPError(http.StatusBadRequest, "user_id is required")
	}

	targetUserID, err := strconv.Atoi(userIDStr)
	if err != nil {
		logger.Error("Invalid user_id", "error", err, "user_id_str", userIDStr)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}
	logger.Info("Removing member from workspace", "workspace_id", id, "target_user_id", targetUserID)

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("Checking permissions", "workspace_id", id, "user_id", user.ID, "target_user_id", targetUserID)

	// Check permissions (owner/admin only, or user removing themselves)
	if targetUserID != int(user.ID) {
		logger.Info("User removing another user, checking owner/admin permissions", "workspace_id", id, "user_id", user.ID)
		if err := h.requireWorkspaceOwnerOrAdmin(ctx, id, int(user.ID)); err != nil {
			logger.Warn("Permission check failed", "workspace_id", id, "user_id", user.ID, "error", err)
			return err
		}
		logger.Info("Permissions verified", "workspace_id", id, "user_id", user.ID)
	} else {
		logger.Info("User removing themselves", "workspace_id", id, "user_id", user.ID)
	}

	// Prevent removing the owner
	logger.Info("Checking target user role", "workspace_id", id, "target_user_id", targetUserID)
	targetRole, err := h.getWorkspaceMemberRole(ctx, id, targetUserID)
	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warn("Workspace remove member failed: user not a member", "workspace_id", id, "target_user_id", targetUserID)
			return echo.NewHTTPError(http.StatusNotFound, "user is not a member of this workspace")
		}
		logger.Error("Failed to check member role", "error", err, "workspace_id", id, "target_user_id", targetUserID)
		return fail(err, "failed to check member role")
	}

	if targetRole == workspacemember.RoleOwner {
		logger.Warn("Workspace remove member failed: cannot remove owner", "workspace_id", id, "target_user_id", targetUserID)
		return echo.NewHTTPError(http.StatusForbidden, "cannot remove workspace owner")
	}
	logger.Info("Target user role verified", "workspace_id", id, "target_user_id", targetUserID, "role", targetRole)

	// Delete workspace member
	logger.Info("Deleting workspace member", "workspace_id", id, "target_user_id", targetUserID)
	_, err = h.orm.WorkspaceMember.
		Delete().
		Where(workspacemember.WorkspaceIDEQ(id)).
		Where(workspacemember.UserIDEQ(targetUserID)).
		Exec(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to remove workspace member", "error", err, "workspace_id", id, "target_user_id", targetUserID)
		return fail(err, "failed to remove workspace member")
	}
	logger.Info("Member removed from workspace", "workspace_id", id, "target_user_id", targetUserID)
	logger.Info("=== WORKSPACE REMOVE MEMBER END ===")

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

// ChannelView отображает страницу канала с сообщениями
// Это основной handler для просмотра канала - загружает канал, сообщения, sidebar и рендерит страницу
// Параметры:
//   - ctx: контекст Echo с запросом (ID канала берётся из URL параметра :id)
//
// Возвращает:
//   - error: ошибка если канал не найден, пользователь не является участником, или ошибка загрузки данных
//
// Middleware:
//   - LoadChannel: загружает канал в контекст
//   - RequireChannelMember: проверяет, что пользователь является участником канала
func (h *Messenger) ChannelView(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== CHANNEL VIEW START ===")

	// Получаем ID канала из URL параметра
	// Например, для URL /channel/123 параметр "id" будет "123"
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logger.Error("Invalid channel ID", "error", err, "id_param", ctx.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid channel ID")
	}
	logger.Info("Viewing channel", "channel_id", id)

	// Загружаем канал из базы данных
	// Middleware LoadChannel уже должен был загрузить канал, но мы загружаем снова для получения всех данных
	logger.Info("Loading channel", "channel_id", id)
	ch, err := h.orm.Channel.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			// Канал не найден в базе данных
			logger.Warn("Channel not found", "channel_id", id)
			return echo.NewHTTPError(http.StatusNotFound, "channel not found")
		}
		// Другая ошибка (например, проблема с подключением к БД)
		logger.Error("Failed to get channel", "error", err, "channel_id", id)
		return fail(err, "failed to get channel")
	}
	logger.Info("Channel loaded", "channel_id", ch.ID, "channel_name", ch.Name, "workspace_id", ch.WorkspaceID)

	// Получаем текущего пользователя из контекста
	// Пользователь был добавлен в контекст middleware RequireAuthentication
	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("User viewing channel", "user_id", user.ID, "user_name", user.Name)

	// Загружаем данные для sidebar (боковая панель с каналами и DM)
	// Sidebar нужен для навигации между каналами и DM
	logger.Info("Loading sidebar data", "channel_id", id, "workspace_id", ch.WorkspaceID, "user_id", user.ID)
	activeChannelID := &id // Указываем, что этот канал активен (будет подсвечен в sidebar)
	sidebarData, err := h.getSidebarData(ctx, ch.WorkspaceID, int(user.ID), activeChannelID, nil)
	if err != nil {
		logger.Error("Failed to load sidebar data", "error", err, "channel_id", id, "workspace_id", ch.WorkspaceID, "user_id", user.ID)
		return fail(err, "failed to load sidebar data")
	}
	logger.Info("Sidebar data loaded", "channels_count", len(sidebarData.Channels), "dms_count", len(sidebarData.DirectMessages))

	// Загружаем последние 50 сообщений из канала
	// Ограничение в 50 сообщений для производительности - при необходимости можно добавить пагинацию
	logger.Info("Loading messages for channel", "channel_id", id)
	messages, err := h.orm.Message.
		Query().                                 // Начинаем запрос сообщений
		Where(message.ChannelIDEQ(id)).          // Фильтр: только сообщения в этом канале
		Order(ent.Desc(message.FieldCreatedAt)). // Сортируем по дате создания (новые сверху)
		Limit(50).                               // Ограничиваем 50 сообщениями
		WithUser().                              // Загружаем связанного пользователя (автора сообщения) через edge
		WithAttachments().                       // Загружаем вложения через edge
		All(ctx.Request().Context())             // Получаем все найденные сообщения

	if err != nil {
		// Ошибка загрузки сообщений (например, проблема с БД)
		return fail(err, "failed to load messages")
	}

	// Преобразуем сообщения из Ent структур в MessageData для UI компонентов
	// MessageData - это упрощённая структура, оптимизированная для отображения
	messageData := make([]messengerComponents.MessageData, len(messages))
	for i, msg := range messages {
		// Загружаем реакции (эмодзи) для каждого сообщения
		// Реакции хранятся отдельно в таблице Reaction, связаны с Message через edge
		reactions, err := h.orm.Reaction.
			Query().                             // Начинаем запрос реакций
			Where(reaction.MessageIDEQ(msg.ID)). // Фильтр: только реакции для этого сообщения
			WithUser().                          // Загружаем пользователей, поставивших реакции
			All(ctx.Request().Context())

		if err != nil {
			// Если не удалось загрузить реакции, продолжаем с пустым массивом
			// Это не критичная ошибка - сообщение всё равно отобразится
			reactions = []*ent.Reaction{} // Continue with empty reactions
		}

		// Группируем реакции по эмодзи
		// Одна реакция может быть поставлена несколькими пользователями
		// Мы группируем их, чтобы показать: "👍 3" вместо трёх отдельных кнопок "👍"
		reactionMap := make(map[string]*messengerComponents.ReactionData)
		for _, r := range reactions {
			if existing, ok := reactionMap[r.Emoji]; ok {
				// Реакция с таким эмодзи уже есть - увеличиваем счётчик
				existing.Count++
				existing.UserIDs = append(existing.UserIDs, int64(r.UserID))
			} else {
				// Первая реакция с таким эмодзи - создаём новую запись
				reactionMap[r.Emoji] = &messengerComponents.ReactionData{
					Emoji:   r.Emoji,
					Count:   1,
					UserIDs: []int64{int64(r.UserID)},
				}
			}
		}

		// Преобразуем map в slice для передачи в UI
		reactionData := make([]messengerComponents.ReactionData, 0, len(reactionMap))
		for _, r := range reactionMap {
			reactionData = append(reactionData, *r)
		}

		// Преобразуем вложения (файлы) в FileAttachmentData
		// Вложения уже загружены через WithAttachments() выше
		attachmentData := make([]messengerComponents.FileAttachmentData, len(msg.Edges.Attachments))
		for j, att := range msg.Edges.Attachments {
			// Формируем URL для доступа к файлу
			attachmentURL := fmt.Sprintf("/attachment/%d", att.ID)
			attachmentData[j] = messengerComponents.FileAttachmentData{
				ID:       int64(att.ID),
				Filename: att.Filename,
				MimeType: att.MimeType,
				FileSize: att.FileSize,
				URL:      attachmentURL,
			}
		}

		// Создаём MessageData для UI компонента
		messageData[i] = messengerComponents.MessageData{
			ID:          int64(msg.ID),
			Content:     msg.Content,
			UserID:      int64(msg.UserID),
			UserName:    msg.Edges.User.Name, // Имя пользователя из загруженного edge
			CreatedAt:   msg.CreatedAt,
			EditedAt:    msg.EditedAt,
			Reactions:   reactionData,
			Attachments: attachmentData,
			ReplyCount:  msg.ReplyCount, // Количество ответов в потоке
		}
	}

	// Переворачиваем массив, чтобы показать старые сообщения первыми
	// Сообщения были отсортированы по убыванию (новые сверху), но в чате нужно показывать старые сверху
	// Это стандартное поведение чатов - новые сообщения появляются внизу
	for i, j := 0, len(messageData)-1; i < j; i, j = i+1, j-1 {
		messageData[i], messageData[j] = messageData[j], messageData[i]
	}

	// Store sidebar data in context
	ctx.Set(context.MessengerSidebarKey, sidebarData)

	// Channel membership is checked by RequireChannelMember middleware
	logger.Info("Rendering channel page", "channel_id", ch.ID, "channel_name", ch.Name, "messages_count", len(messageData))
	logger.Info("=== CHANNEL VIEW END ===")
	return messengerPages.Channel(ctx, int64(ch.ID), ch.Name, messageData)
}

// ChannelCreate creates a new channel
func (h *Messenger) ChannelCreate(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== CHANNEL CREATE START ===")

	workspaceID, err := strconv.Atoi(ctx.Param("workspace_id"))
	if err != nil {
		logger.Error("Invalid workspace ID", "error", err, "workspace_id_param", ctx.Param("workspace_id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}
	logger.Info("Creating channel in workspace", "workspace_id", workspaceID)

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("Channel creator", "user_id", user.ID, "user_name", user.Name)

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

	// If HTMX request, close modal and redirect
	if htmx.GetRequest(ctx).Enabled {
		// Проверяем, что Echo не nil перед вызовом Reverse
		if echoInstance := ctx.Echo(); echoInstance != nil {
			redirectURL := echoInstance.Reverse(routenames.MessengerChannelView, ch.ID)
			logger.Info("HTMX request detected, redirecting", "redirect_url", redirectURL, "channel_id", ch.ID)
			htmx.Response{Redirect: redirectURL}.Apply(ctx)
		} else {
			// Fallback: формируем URL вручную
			redirectURL := fmt.Sprintf("/channel/%d", ch.ID)
			logger.Warn("Echo instance is nil, using manual URL", "redirect_url", redirectURL, "channel_id", ch.ID)
			htmx.Response{Redirect: redirectURL}.Apply(ctx)
		}
		logger.Info("=== CHANNEL CREATE END ===")
		return ctx.NoContent(http.StatusOK)
	}

	logger.Info("Non-HTMX request, returning JSON response", "channel_id", ch.ID)
	logger.Info("=== CHANNEL CREATE END ===")
	return ctx.JSON(http.StatusCreated, ch)
}

// ChannelCreateForm renders the channel creation form modal
func (h *Messenger) ChannelCreateForm(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== CHANNEL CREATE FORM START ===")

	workspaceIDParam := ctx.Param("workspace_id")
	logger.Info("Parsing workspace ID from URL", "workspace_id_param", workspaceIDParam, "all_params", ctx.ParamNames())

	workspaceID, err := strconv.Atoi(workspaceIDParam)
	if err != nil {
		logger.Error("Invalid workspace ID", "error", err, "workspace_id_param", workspaceIDParam, "all_params", ctx.ParamNames())
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace ID")
	}

	if workspaceID <= 0 {
		logger.Error("Workspace ID is zero or negative", "workspace_id", workspaceID, "workspace_id_param", workspaceIDParam)
		return echo.NewHTTPError(http.StatusBadRequest, "workspace ID must be greater than 0")
	}

	logger.Info("Rendering channel create form", "workspace_id", workspaceID)

	// Проверяем, что контекст не nil
	if ctx == nil {
		logger.Error("Context is nil")
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	// Проверяем, что Request() не nil перед созданием UI Request
	if ctx.Request() == nil {
		logger.Error("Request is nil")
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	r := ui.NewRequest(ctx)
	if r == nil {
		logger.Error("Failed to create UI request")
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	// Проверяем, что Context в Request не nil
	if r.Context == nil {
		logger.Error("Request context is nil")
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	// Проверяем, что Echo доступен
	if r.Context.Echo() == nil {
		logger.Warn("Echo instance is nil, using fallback paths")
		// Продолжаем, так как Path() теперь имеет fallback
	}

	logger.Info("Rendering channel create modal", "workspace_id", workspaceID)

	// Проверяем, что r полностью инициализирован
	if r == nil {
		logger.Error("Request is nil")
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	// Проверяем, что CSRF токен доступен (может быть пустым, но не должен вызывать панику)
	logger.Info("Request details", "csrf_length", len(r.CSRF), "has_context", r.Context != nil, "has_echo", r.Context != nil && r.Context.Echo() != nil)

	// Защита от паники при создании и рендеринге модального окна
	var htmlContent string
	var panicErr interface{}
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				panicErr = rec
				logger.Error("Panic while creating/rendering modal", "panic", rec, "workspace_id", workspaceID, "panic_type", fmt.Sprintf("%T", rec))
				// Создаём простой HTML модал с ошибкой как fallback
				htmlContent = fmt.Sprintf(`<div id="channel-create-modal" class="modal"><div class="modal-box"><div class="text-error">Error: Failed to create modal. Please try again.</div></div></div>`)
			}
		}()

		logger.Info("Creating modal component", "workspace_id", workspaceID)
		// Создаём модальное окно
		modal := messengerComponents.ChannelCreateModal(r, int64(workspaceID), nil)

		if modal == nil {
			logger.Error("Modal component is nil")
			htmlContent = fmt.Sprintf(`<div id="channel-create-modal" class="modal"><div class="modal-box"><div class="text-error">Error: Modal component is nil</div></div></div>`)
			return
		}

		logger.Info("Modal component created, rendering to buffer")
		// Рендерим в буфер
		var buf bytes.Buffer
		if err := modal.Render(&buf); err != nil {
			logger.Error("Failed to render modal", "error", err)
			htmlContent = fmt.Sprintf(`<div id="channel-create-modal" class="modal"><div class="modal-box"><div class="text-error">Error: Failed to render modal: %v</div></div></div>`, err)
			return
		}
		htmlContent = buf.String()
		logger.Info("Modal rendered successfully", "buffer_size", len(htmlContent))
		// Логируем первые 500 символов HTML для отладки
		if len(htmlContent) > 500 {
			logger.Info("Modal HTML preview", "html_preview", htmlContent[:500]+"...")
		} else {
			logger.Info("Modal HTML", "html", htmlContent)
		}
	}()

	if panicErr != nil {
		logger.Error("Panic occurred during modal creation/render", "panic", panicErr, "html_content_length", len(htmlContent))
	}

	if htmlContent == "" {
		// Если по какой-то причине контент пустой, возвращаем ошибку
		logger.Error("Modal content is empty")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to render modal")
	}

	logger.Info("=== CHANNEL CREATE FORM END ===")
	return ctx.HTML(http.StatusOK, htmlContent)
}

// ChannelUpdate updates a channel
func (h *Messenger) ChannelUpdate(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== CHANNEL UPDATE START ===")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logger.Error("Invalid channel ID", "error", err, "id_param", ctx.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid channel ID")
	}
	logger.Info("Updating channel", "channel_id", id)

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("Channel updater", "user_id", user.ID, "user_name", user.Name)

	// Get channel
	logger.Info("Loading channel", "channel_id", id)
	ch, err := h.orm.Channel.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warn("Channel update failed: channel not found", "channel_id", id)
			return echo.NewHTTPError(http.StatusNotFound, "channel not found")
		}
		logger.Error("Failed to get channel", "error", err, "channel_id", id)
		return fail(err, "failed to get channel")
	}
	logger.Info("Channel loaded", "channel_id", ch.ID, "channel_name", ch.Name, "workspace_id", ch.WorkspaceID, "created_by", ch.CreatedBy)

	// Check if user is creator or workspace owner/admin
	if ch.CreatedBy != int(user.ID) {
		logger.Info("User is not channel creator, checking workspace role", "channel_id", id, "user_id", user.ID)
		// Check workspace role
		if err := h.requireWorkspaceOwnerOrAdmin(ctx, ch.WorkspaceID, int(user.ID)); err != nil {
			logger.Warn("Permission check failed", "channel_id", id, "user_id", user.ID, "error", err)
			return err
		}
		logger.Info("Permission verified: user is workspace owner/admin", "channel_id", id, "user_id", user.ID)
	} else {
		logger.Info("User is channel creator", "channel_id", id, "user_id", user.ID)
	}

	// Parse form data
	update := h.orm.Channel.UpdateOneID(id)
	hasUpdates := false

	if name := ctx.FormValue("name"); name != "" {
		logger.Info("Updating channel name", "channel_id", id, "new_name", name)
		update = update.SetName(name)
		hasUpdates = true
	}

	if slug := ctx.FormValue("slug"); slug != "" {
		logger.Info("Checking channel slug availability", "channel_id", id, "new_slug", slug)
		// Check if slug already exists in workspace (excluding current channel)
		exists, err := h.orm.Channel.
			Query().
			Where(channel.WorkspaceIDEQ(ch.WorkspaceID)).
			Where(channel.SlugEQ(slug)).
			Where(channel.IDNEQ(id)).
			Exist(ctx.Request().Context())

		if err != nil {
			logger.Error("Failed to check channel slug", "error", err, "channel_id", id, "slug", slug)
			return fail(err, "failed to check channel slug")
		}

		if exists {
			logger.Warn("Channel update failed: slug already exists", "channel_id", id, "slug", slug)
			return echo.NewHTTPError(http.StatusConflict, "channel with this slug already exists in this workspace")
		}

		logger.Info("Updating channel slug", "channel_id", id, "new_slug", slug)
		update = update.SetSlug(slug)
		hasUpdates = true
	}

	if description := ctx.FormValue("description"); description != "" {
		logger.Info("Updating channel description", "channel_id", id, "description_length", len(description))
		update = update.SetDescription(description)
		hasUpdates = true
	}

	if isPrivateStr := ctx.FormValue("is_private"); isPrivateStr != "" {
		isPrivate := isPrivateStr == "true" || isPrivateStr == "1"
		logger.Info("Updating channel privacy", "channel_id", id, "is_private", isPrivate)
		update = update.SetIsPrivate(isPrivate)
		hasUpdates = true
	}

	if !hasUpdates {
		logger.Warn("Channel update: no fields to update", "channel_id", id)
		return ctx.JSON(http.StatusOK, ch)
	}

	logger.Info("Saving channel updates", "channel_id", id)
	ch, err = update.Save(ctx.Request().Context())
	if err != nil {
		logger.Error("Failed to update channel", "error", err, "channel_id", id)
		return fail(err, "failed to update channel")
	}
	logger.Info("Channel updated successfully", "channel_id", ch.ID, "channel_name", ch.Name)

	// Send WebSocket event
	if h.hub != nil {
		logger.Info("Sending WebSocket event: channel updated", "channel_id", id)
		event := ws.ChannelUpdatedEvent(int64(id))
		h.hub.SendToChannel(int64(id), event.ToJSON())
	}

	logger.Info("=== CHANNEL UPDATE END ===")
	return ctx.JSON(http.StatusOK, ch)
}

// ChannelDelete deletes a channel
func (h *Messenger) ChannelDelete(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== CHANNEL DELETE START ===")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logger.Error("Invalid channel ID", "error", err, "id_param", ctx.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid channel ID")
	}
	logger.Info("Deleting channel", "channel_id", id)

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("Channel deleter", "user_id", user.ID, "user_name", user.Name)

	// Get channel
	logger.Info("Loading channel", "channel_id", id)
	ch, err := h.orm.Channel.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warn("Channel delete failed: channel not found", "channel_id", id)
			return echo.NewHTTPError(http.StatusNotFound, "channel not found")
		}
		logger.Error("Failed to get channel", "error", err, "channel_id", id)
		return fail(err, "failed to get channel")
	}
	logger.Info("Channel loaded", "channel_id", ch.ID, "channel_name", ch.Name, "workspace_id", ch.WorkspaceID, "created_by", ch.CreatedBy)

	// Check if user is creator or workspace owner/admin
	if ch.CreatedBy != int(user.ID) {
		logger.Info("User is not channel creator, checking workspace role", "channel_id", id, "user_id", user.ID)
		// Check workspace role
		if err := h.requireWorkspaceOwnerOrAdmin(ctx, ch.WorkspaceID, int(user.ID)); err != nil {
			logger.Warn("Permission check failed", "channel_id", id, "user_id", user.ID, "error", err)
			return err
		}
		logger.Info("Permission verified: user is workspace owner/admin", "channel_id", id, "user_id", user.ID)
	} else {
		logger.Info("User is channel creator", "channel_id", id, "user_id", user.ID)
	}

	// Delete channel
	logger.Info("Deleting channel from database", "channel_id", id, "workspace_id", ch.WorkspaceID)
	err = h.orm.Channel.DeleteOneID(id).Exec(ctx.Request().Context())
	if err != nil {
		logger.Error("Failed to delete channel", "error", err, "channel_id", id)
		return fail(err, "failed to delete channel")
	}
	logger.Info("Channel deleted successfully", "channel_id", id, "workspace_id", ch.WorkspaceID)

	// Send WebSocket event
	if h.hub != nil {
		logger.Info("Sending WebSocket event: channel deleted", "channel_id", id, "workspace_id", ch.WorkspaceID)
		event := &ws.Event{
			Type: ws.EventTypeChannelUpdated,
			Data: map[string]interface{}{
				"channel_id": int64(id),
				"deleted":    true,
			},
		}
		h.hub.SendToWorkspace(ctx.Request().Context(), int64(ch.WorkspaceID), event.ToJSON())
	}

	logger.Info("=== CHANNEL DELETE END ===")
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
	logger := log.Ctx(ctx)
	logger.Info("=== MESSAGE CREATE START ===")

	channelID, err := strconv.Atoi(ctx.Param("channel_id"))
	if err != nil {
		logger.Error("Invalid channel ID", "error", err, "channel_id_param", ctx.Param("channel_id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid channel ID")
	}
	logger.Info("Creating message in channel", "channel_id", channelID)

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("User creating message", "user_id", user.ID, "user_name", user.Name)

	// Parse form data
	content := ctx.FormValue("content")

	// Check if there are files being uploaded
	form, _ := ctx.MultipartForm()
	hasFiles := form != nil && form.File != nil && len(form.File["files"]) > 0

	if content == "" && !hasFiles {
		logger.Warn("Message content is empty and no files provided")
		return echo.NewHTTPError(http.StatusBadRequest, "message content or file is required")
	}
	logger.Info("Message content parsed", "content_length", len(content), "has_files", hasFiles)

	// Создаём сообщение в базе данных
	// Используем Ent ORM для создания записи в таблице Message
	msg, err := h.orm.Message.
		Create().                                // Начинаем создание новой записи
		SetContent(content).                     // Устанавливаем текст сообщения
		SetMessageType(message.MessageTypeText). // Тип сообщения: text (может быть также file или thread_reply)
		SetChannelID(channelID).                 // Связываем с каналом
		SetUserID(int(user.ID)).                 // Устанавливаем автора сообщения
		Save(ctx.Request().Context())            // Сохраняем в базу данных

	if err != nil {
		// Ошибка создания сообщения (например, нарушение внешнего ключа, проблема с БД)
		logger.Error("Failed to create message", "error", err, "channel_id", channelID, "user_id", user.ID)
		return fail(err, "failed to create message")
	}
	logger.Info("Message created successfully", "message_id", msg.ID, "channel_id", channelID)

	// Handle file attachments (reuse form if already parsed)
	if form == nil {
		form, _ = ctx.MultipartForm()
	}
	if err == nil && form != nil && form.File != nil {
		files := form.File["files"]
		if len(files) > 0 {
			logger.Info("Processing file attachments", "file_count", len(files))
			for _, fileHeader := range files {
				// Validate file size
				maxSize := int64(10485760) // Default 10MB
				if h.config != nil && h.config.Messenger.MaxFileSize > 0 {
					maxSize = h.config.Messenger.MaxFileSize
				}
				if fileHeader.Size > maxSize {
					logger.Warn("File size exceeds maximum", "filename", fileHeader.Filename, "size", fileHeader.Size, "max_size", maxSize)
					continue
				}

				// Validate MIME type
				mimeType := fileHeader.Header.Get("Content-Type")
				if h.config != nil && len(h.config.Messenger.AllowedFileTypes) > 0 {
					allowed := false
					for _, allowedType := range h.config.Messenger.AllowedFileTypes {
						if matchesMimeType(mimeType, allowedType) {
							allowed = true
							break
						}
					}
					if !allowed {
						logger.Warn("File type not allowed", "filename", fileHeader.Filename, "mime_type", mimeType)
						continue
					}
				}

				// Открываем загруженный файл для чтения
				// fileHeader.Open() возвращает io.ReadCloser для чтения содержимого файла
				src, err := fileHeader.Open()
				if err != nil {
					logger.Error("Failed to open uploaded file", "error", err, "filename", fileHeader.Filename)
					continue // Пропускаем этот файл, продолжаем с остальными
				}

				// Создаём путь для сохранения файла
				// Организуем файлы по каналам: attachments/channel/{channel_id}_{filename}
				// Это упрощает управление файлами и позволяет легко найти все файлы канала
				filePath := filepath.Join("attachments", "channel", fmt.Sprintf("%d_%s", channelID, fileHeader.Filename))
				dst, err := h.files.Create(filePath) // Создаём файл в файловой системе (afero.Fs)
				if err != nil {
					src.Close() // Закрываем источник при ошибке
					logger.Error("Failed to create file", "error", err, "filepath", filePath)
					continue
				}

				// Копируем содержимое файла из источника в место назначения
				// io.Copy читает из src и записывает в dst до EOF
				if _, err = io.Copy(dst, src); err != nil {
					src.Close()
					dst.Close() // Закрываем оба файла при ошибке
					logger.Error("Failed to save file", "error", err)
					continue
				}
				src.Close() // Закрываем источник после успешного копирования
				dst.Close() // Закрываем файл назначения (сохраняем изменения)

				// Получаем информацию о сохранённом файле
				// Нужно для получения реального размера файла (может отличаться от fileHeader.Size)
				fileInfo, err := h.files.Stat(filePath)
				if err != nil {
					logger.Error("Failed to get file info", "error", err)
					continue
				}

				// Создаём запись о вложении в базе данных
				// Attachment связывает файл с сообщением и хранит метаданные
				_, err = h.orm.Attachment.
					Create().                         // Начинаем создание записи
					SetFilename(fileHeader.Filename). // Оригинальное имя файла
					SetFilepath(filePath).            // Путь к файлу в файловой системе
					SetFileSize(fileInfo.Size()).     // Размер файла в байтах
					SetMimeType(mimeType).            // MIME тип файла
					SetMessageID(msg.ID).             // Связываем с сообщением
					SetUploadedBy(int(user.ID)).      // Кто загрузил файл
					Save(ctx.Request().Context())     // Сохраняем в БД

				if err != nil {
					logger.Error("Failed to create attachment record", "error", err)
				} else {
					logger.Info("Attachment created", "filename", fileHeader.Filename, "message_id", msg.ID)
				}
			}
		}
	}

	// Send WebSocket event to channel members
	if h.hub != nil {
		event := ws.MessageNewEvent(int64(msg.ID), int64(channelID), int64(user.ID), content)
		h.hub.SendToChannel(int64(channelID), event.ToJSON())
		logger.Info("WebSocket event sent", "message_id", msg.ID, "channel_id", channelID)
	}

	// If HTMX request, return HTML for the new message
	if htmx.GetRequest(ctx).Enabled {
		logger.Info("HTMX request detected, returning HTML message item")
		// Load user for message display
		msgWithUser, err := h.orm.Message.Query().Where(message.IDEQ(msg.ID)).WithUser().Only(ctx.Request().Context())
		if err != nil {
			logger.Warn("Failed to load user for message, using basic data", "error", err)
			msgWithUser = msg
		}

		// Load attachments
		msgWithAttachments, err := h.orm.Message.Query().Where(message.IDEQ(msg.ID)).WithAttachments().Only(ctx.Request().Context())
		if err != nil {
			msgWithAttachments = msg
		}

		// Convert attachments
		attachmentData := make([]messengerComponents.FileAttachmentData, len(msgWithAttachments.Edges.Attachments))
		for j, att := range msgWithAttachments.Edges.Attachments {
			attachmentURL := fmt.Sprintf("/attachment/%d", att.ID)
			attachmentData[j] = messengerComponents.FileAttachmentData{
				ID:       int64(att.ID),
				Filename: att.Filename,
				MimeType: att.MimeType,
				FileSize: att.FileSize,
				URL:      attachmentURL,
			}
		}

		// Convert to MessageData
		messageData := messengerComponents.MessageData{
			ID:          int64(msg.ID),
			Content:     msg.Content,
			UserID:      int64(msg.UserID),
			UserName:    msgWithUser.Edges.User.Name,
			CreatedAt:   msg.CreatedAt,
			EditedAt:    msg.EditedAt,
			Reactions:   []messengerComponents.ReactionData{},
			Attachments: attachmentData,
		}

		r := ui.NewRequest(ctx)
		messageItem := messengerComponents.MessageItem(r, messageData)
		var buf bytes.Buffer
		if err := messageItem.Render(&buf); err != nil {
			logger.Error("Failed to render message item", "error", err)
			return fail(err, "failed to render message")
		}
		logger.Info("=== MESSAGE CREATE END ===")
		return ctx.HTML(http.StatusOK, buf.String())
	}

	logger.Info("Non-HTMX request, returning JSON", "message_id", msg.ID)
	logger.Info("=== MESSAGE CREATE END ===")
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
	logger := log.Ctx(ctx)
	logger.Info("=== MESSAGE DELETE START ===")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logger.Error("Invalid message ID", "error", err, "id_param", ctx.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid message ID")
	}
	logger.Info("Deleting message", "message_id", id)

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("Message deleter", "user_id", user.ID, "user_name", user.Name)

	// Get message
	logger.Info("Loading message", "message_id", id)
	msg, err := h.orm.Message.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warn("Message delete failed: message not found", "message_id", id)
			return echo.NewHTTPError(http.StatusNotFound, "message not found")
		}
		logger.Error("Failed to get message", "error", err, "message_id", id)
		return fail(err, "failed to get message")
	}
	logger.Info("Message loaded", "message_id", msg.ID, "channel_id", msg.ChannelID, "message_owner_id", msg.UserID)

	// Check if user owns the message
	if msg.UserID != int(user.ID) {
		logger.Warn("Message delete failed: user does not own message", "message_id", id, "user_id", user.ID, "message_owner_id", msg.UserID)
		return echo.NewHTTPError(http.StatusForbidden, "you can only delete your own messages")
	}
	logger.Info("Ownership verified", "message_id", id, "user_id", user.ID)

	channelID := msg.ChannelID

	// Delete message
	logger.Info("Deleting message from database", "message_id", id, "channel_id", channelID)
	err = h.orm.Message.DeleteOneID(id).Exec(ctx.Request().Context())
	if err != nil {
		logger.Error("Failed to delete message", "error", err, "message_id", id)
		return fail(err, "failed to delete message")
	}
	logger.Info("Message deleted successfully", "message_id", id, "channel_id", channelID)

	// Send WebSocket event
	if h.hub != nil {
		logger.Info("Sending WebSocket event: message deleted", "message_id", id, "channel_id", channelID)
		event := &ws.Event{
			Type: ws.EventTypeMessageDeleted,
			Data: map[string]interface{}{
				"message_id": int64(id),
				"channel_id": int64(channelID),
			},
		}
		h.hub.SendToChannel(int64(channelID), event.ToJSON())
	}

	logger.Info("=== MESSAGE DELETE END ===")
	return ctx.NoContent(http.StatusNoContent)
}

// MessageReplies returns replies to a message (thread)
func (h *Messenger) MessageReplies(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("Getting message replies")

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
		WithUser().
		WithAttachments().
		All(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to fetch replies", "error", err, "message_id", id)
		return fail(err, "failed to fetch replies")
	}

	// If HTMX request, return HTML
	if htmx.GetRequest(ctx).Enabled {
		r := ui.NewRequest(ctx)

		// Convert to MessageData
		replyData := make([]messengerComponents.MessageData, len(replies))
		for i, reply := range replies {
			// Get reactions
			reactions, err := h.orm.Reaction.
				Query().
				Where(reaction.MessageIDEQ(reply.ID)).
				WithUser().
				All(ctx.Request().Context())

			if err != nil {
				reactions = []*ent.Reaction{}
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

			// Convert attachments
			attachmentData := make([]messengerComponents.FileAttachmentData, len(reply.Edges.Attachments))
			for j, att := range reply.Edges.Attachments {
				attachmentURL := fmt.Sprintf("/attachment/%d", att.ID)
				attachmentData[j] = messengerComponents.FileAttachmentData{
					ID:       int64(att.ID),
					Filename: att.Filename,
					MimeType: att.MimeType,
					FileSize: att.FileSize,
					URL:      attachmentURL,
				}
			}

			replyData[i] = messengerComponents.MessageData{
				ID:          int64(reply.ID),
				Content:     reply.Content,
				UserID:      int64(reply.UserID),
				UserName:    reply.Edges.User.Name,
				CreatedAt:   reply.CreatedAt,
				EditedAt:    reply.EditedAt,
				Reactions:   reactionData,
				Attachments: attachmentData,
				ReplyCount:  0, // Replies don't have their own replies
			}
		}

		// Render replies
		var buf bytes.Buffer
		for _, reply := range replyData {
			if err := messengerComponents.MessageItem(r, reply).Render(&buf); err != nil {
				logger.Error("Failed to render reply", "error", err)
				continue
			}
		}

		return ctx.HTML(http.StatusOK, buf.String())
	}

	return ctx.JSON(http.StatusOK, replies)
}

// MessageThread renders the thread view page
func (h *Messenger) MessageThread(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== MESSAGE THREAD VIEW START ===")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logger.Error("Invalid message ID", "error", err, "id_param", ctx.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid message ID")
	}
	logger.Info("Viewing thread", "message_id", id)

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("User viewing thread", "user_id", user.ID, "user_name", user.Name)

	// Get parent message
	logger.Info("Loading parent message", "message_id", id)
	parentMsg, err := h.orm.Message.
		Query().
		Where(message.IDEQ(id)).
		WithUser().
		WithChannel().
		WithAttachments().
		Only(ctx.Request().Context())

	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warn("Thread view failed: message not found", "message_id", id)
			return echo.NewHTTPError(http.StatusNotFound, "message not found")
		}
		logger.Error("Failed to get parent message", "error", err, "message_id", id)
		return fail(err, "failed to get message")
	}
	logger.Info("Parent message loaded", "message_id", parentMsg.ID, "channel_id", parentMsg.ChannelID)

	// Check if user is a member of the channel
	isMember, err := h.orm.ChannelMember.
		Query().
		Where(
			channelmember.ChannelIDEQ(parentMsg.ChannelID),
			channelmember.UserIDEQ(int(user.ID)),
		).
		Exist(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to check channel membership", "error", err)
		return fail(err, "failed to check channel membership")
	}

	if !isMember {
		logger.Warn("Thread view failed: user not a channel member", "message_id", id, "channel_id", parentMsg.ChannelID, "user_id", user.ID)
		return echo.NewHTTPError(http.StatusForbidden, "you are not a member of this channel")
	}
	logger.Info("User verified as channel member", "channel_id", parentMsg.ChannelID)

	// Get sidebar data
	logger.Info("Loading sidebar data", "channel_id", parentMsg.ChannelID, "workspace_id", parentMsg.Edges.Channel.WorkspaceID, "user_id", user.ID)
	activeChannelID := &parentMsg.ChannelID
	sidebarData, err := h.getSidebarData(ctx, parentMsg.Edges.Channel.WorkspaceID, int(user.ID), activeChannelID, nil)
	if err != nil {
		logger.Error("Failed to load sidebar data", "error", err)
		return fail(err, "failed to load sidebar data")
	}
	logger.Info("Sidebar data loaded", "channels_count", len(sidebarData.Channels), "dms_count", len(sidebarData.DirectMessages))
	ctx.Set(context.MessengerSidebarKey, sidebarData)

	// Get replies
	logger.Info("Loading thread replies", "message_id", id)
	replies, err := h.orm.Message.
		Query().
		Where(message.ThreadIDEQ(id)).
		Order(ent.Asc(message.FieldCreatedAt)).
		WithUser().
		WithAttachments().
		All(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to load replies", "error", err, "message_id", id)
		return fail(err, "failed to load replies")
	}
	logger.Info("Replies loaded", "message_id", id, "replies_count", len(replies))

	// Convert parent message to MessageData
	parentReactions, _ := h.orm.Reaction.
		Query().
		Where(reaction.MessageIDEQ(parentMsg.ID)).
		WithUser().
		All(ctx.Request().Context())

	reactionMap := make(map[string]*messengerComponents.ReactionData)
	for _, r := range parentReactions {
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

	attachmentData := make([]messengerComponents.FileAttachmentData, len(parentMsg.Edges.Attachments))
	for j, att := range parentMsg.Edges.Attachments {
		attachmentURL := fmt.Sprintf("/attachment/%d", att.ID)
		attachmentData[j] = messengerComponents.FileAttachmentData{
			ID:       int64(att.ID),
			Filename: att.Filename,
			MimeType: att.MimeType,
			FileSize: att.FileSize,
			URL:      attachmentURL,
		}
	}

	parentData := messengerComponents.MessageData{
		ID:          int64(parentMsg.ID),
		Content:     parentMsg.Content,
		UserID:      int64(parentMsg.UserID),
		UserName:    parentMsg.Edges.User.Name,
		CreatedAt:   parentMsg.CreatedAt,
		EditedAt:    parentMsg.EditedAt,
		Reactions:   reactionData,
		Attachments: attachmentData,
		ReplyCount:  parentMsg.ReplyCount,
	}

	// Convert replies to MessageData
	replyData := make([]messengerComponents.MessageData, len(replies))
	for i, reply := range replies {
		replyReactions, _ := h.orm.Reaction.
			Query().
			Where(reaction.MessageIDEQ(reply.ID)).
			WithUser().
			All(ctx.Request().Context())

		replyReactionMap := make(map[string]*messengerComponents.ReactionData)
		for _, r := range replyReactions {
			if existing, ok := replyReactionMap[r.Emoji]; ok {
				existing.Count++
				existing.UserIDs = append(existing.UserIDs, int64(r.UserID))
			} else {
				replyReactionMap[r.Emoji] = &messengerComponents.ReactionData{
					Emoji:   r.Emoji,
					Count:   1,
					UserIDs: []int64{int64(r.UserID)},
				}
			}
		}

		replyReactionData := make([]messengerComponents.ReactionData, 0, len(replyReactionMap))
		for _, r := range replyReactionMap {
			replyReactionData = append(replyReactionData, *r)
		}

		replyAttachmentData := make([]messengerComponents.FileAttachmentData, len(reply.Edges.Attachments))
		for j, att := range reply.Edges.Attachments {
			attachmentURL := fmt.Sprintf("/attachment/%d", att.ID)
			replyAttachmentData[j] = messengerComponents.FileAttachmentData{
				ID:       int64(att.ID),
				Filename: att.Filename,
				MimeType: att.MimeType,
				FileSize: att.FileSize,
				URL:      attachmentURL,
			}
		}

		replyData[i] = messengerComponents.MessageData{
			ID:          int64(reply.ID),
			Content:     reply.Content,
			UserID:      int64(reply.UserID),
			UserName:    reply.Edges.User.Name,
			CreatedAt:   reply.CreatedAt,
			EditedAt:    reply.EditedAt,
			Reactions:   replyReactionData,
			Attachments: replyAttachmentData,
			ReplyCount:  0,
		}
	}

	logger.Info("Rendering thread page", "message_id", id, "channel_id", parentMsg.ChannelID, "replies_count", len(replyData))
	logger.Info("=== MESSAGE THREAD VIEW END ===")
	return messengerPages.Thread(ctx, int64(parentMsg.ChannelID), parentData, replyData)
}

// MessageReply creates a reply to a message (thread)
func (h *Messenger) MessageReply(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== MESSAGE REPLY START ===")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logger.Error("Invalid message ID", "error", err, "id_param", ctx.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid message ID")
	}
	logger.Info("Creating reply to message", "parent_message_id", id)

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("Reply creator", "user_id", user.ID, "user_name", user.Name)

	// Get parent message
	logger.Info("Loading parent message", "parent_message_id", id)
	parentMsg, err := h.orm.Message.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warn("Message reply failed: parent message not found", "parent_message_id", id)
			return echo.NewHTTPError(http.StatusNotFound, "message not found")
		}
		logger.Error("Failed to get parent message", "error", err, "parent_message_id", id)
		return fail(err, "failed to get message")
	}
	logger.Info("Parent message loaded", "parent_message_id", parentMsg.ID, "channel_id", parentMsg.ChannelID)

	// Parse form data
	content := ctx.FormValue("content")
	if content == "" {
		logger.Warn("Message reply failed: content is required", "parent_message_id", id)
		return echo.NewHTTPError(http.StatusBadRequest, "message content is required")
	}
	logger.Info("Reply content", "parent_message_id", id, "content_length", len(content))

	// Create reply
	logger.Info("Creating reply in database", "parent_message_id", id, "channel_id", parentMsg.ChannelID, "user_id", user.ID)
	reply, err := h.orm.Message.
		Create().
		SetContent(content).
		SetMessageType(message.MessageTypeThreadReply).
		SetChannelID(parentMsg.ChannelID).
		SetUserID(int(user.ID)).
		SetThreadID(id).
		Save(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to create reply", "error", err, "parent_message_id", id, "channel_id", parentMsg.ChannelID, "user_id", user.ID)
		return fail(err, "failed to create reply")
	}
	logger.Info("Reply created", "reply_id", reply.ID, "parent_message_id", id, "channel_id", parentMsg.ChannelID)

	// Update reply count on parent message
	logger.Info("Updating reply count on parent message", "parent_message_id", id)
	_, err = h.orm.Message.
		UpdateOneID(id).
		AddReplyCount(1).
		Save(ctx.Request().Context())

	if err != nil {
		// Log but don't fail
		logger.Warn("Failed to update reply count", "parent_message_id", id, "error", err)
	} else {
		logger.Info("Reply count updated", "parent_message_id", id)
	}

	// Send WebSocket event
	if h.hub != nil {
		logger.Info("Sending WebSocket event: message new (reply)", "reply_id", reply.ID, "parent_message_id", id, "channel_id", parentMsg.ChannelID)
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

	// If HTMX request, return HTML for the new reply
	if htmx.GetRequest(ctx).Enabled {
		logger.Info("HTMX request detected, returning HTML reply item")
		// Load user for message display
		replyWithUser, err := h.orm.Message.Query().Where(message.IDEQ(reply.ID)).WithUser().WithAttachments().Only(ctx.Request().Context())
		if err != nil {
			logger.Warn("Failed to load user for reply, using basic data", "error", err)
			replyWithUser = reply
		}

		// Get reactions
		reactions, _ := h.orm.Reaction.
			Query().
			Where(reaction.MessageIDEQ(reply.ID)).
			WithUser().
			All(ctx.Request().Context())

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

		// Convert attachments
		attachmentData := make([]messengerComponents.FileAttachmentData, len(replyWithUser.Edges.Attachments))
		for j, att := range replyWithUser.Edges.Attachments {
			attachmentURL := fmt.Sprintf("/attachment/%d", att.ID)
			attachmentData[j] = messengerComponents.FileAttachmentData{
				ID:       int64(att.ID),
				Filename: att.Filename,
				MimeType: att.MimeType,
				FileSize: att.FileSize,
				URL:      attachmentURL,
			}
		}

		// Convert to MessageData
		messageData := messengerComponents.MessageData{
			ID:          int64(reply.ID),
			Content:     reply.Content,
			UserID:      int64(reply.UserID),
			UserName:    replyWithUser.Edges.User.Name,
			CreatedAt:   reply.CreatedAt,
			EditedAt:    reply.EditedAt,
			Reactions:   reactionData,
			Attachments: attachmentData,
			ReplyCount:  0, // Replies don't have their own replies
		}

		r := ui.NewRequest(ctx)
		messageItem := messengerComponents.MessageItem(r, messageData)
		var buf bytes.Buffer
		if err := messageItem.Render(&buf); err != nil {
			logger.Error("Failed to render reply item", "error", err)
			return fail(err, "failed to render reply")
		}
		logger.Info("=== MESSAGE REPLY END ===")
		return ctx.HTML(http.StatusOK, buf.String())
	}

	logger.Info("Non-HTMX request, returning JSON", "reply_id", reply.ID)
	logger.Info("=== MESSAGE REPLY END ===")
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

// DMView отображает страницу прямого сообщения (DM) с историей переписки
// Прямое сообщение - это приватная переписка между двумя пользователями
// Параметры:
//   - ctx: контекст Echo с запросом (ID DM берётся из URL параметра :id)
//
// Возвращает:
//   - error: ошибка если DM не найден, пользователь не является участником, или ошибка загрузки данных
func (h *Messenger) DMView(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== DM VIEW START ===")

	// Получаем ID прямого сообщения из URL параметра
	// Например, для URL /dm/123 параметр "id" будет "123"
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logger.Error("Invalid DM ID", "error", err, "id_param", ctx.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid DM ID")
	}
	logger.Info("Viewing direct message conversation", "dm_id", id)

	// Получаем текущего пользователя из контекста
	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("User viewing DM", "user_id", user.ID, "user_name", user.Name)

	// Загружаем DM из базы данных
	// DirectMessage содержит двух участников: User1ID и User2ID
	logger.Info("Loading DM conversation", "dm_id", id)
	dm, err := h.orm.DirectMessage.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			// DM не найден в базе данных
			logger.Warn("Direct message not found", "dm_id", id)
			return echo.NewHTTPError(http.StatusNotFound, "direct message not found")
		}
		logger.Error("Failed to get direct message", "error", err, "dm_id", id)
		return fail(err, "failed to get direct message")
	}
	logger.Info("DM loaded", "dm_id", dm.ID, "user1_id", dm.User1ID, "user2_id", dm.User2ID)

	// Проверяем, является ли пользователь участником этого DM
	// Пользователь должен быть либо User1ID, либо User2ID
	// Это защита от просмотра чужих приватных переписок
	if dm.User1ID != int(user.ID) && dm.User2ID != int(user.ID) {
		logger.Warn("User not authorized to view this DM", "user_id", user.ID, "dm_id", id)
		return echo.NewHTTPError(http.StatusForbidden, "you are not part of this conversation")
	}
	logger.Info("User verified as part of DM", "user_id", user.ID)

	// Get other user
	var otherUserID int
	if dm.User1ID == int(user.ID) {
		otherUserID = dm.User2ID
	} else {
		otherUserID = dm.User1ID
	}
	logger.Info("Determined other user in DM", "other_user_id", otherUserID)

	logger.Info("Loading other user", "other_user_id", otherUserID)
	otherUser, err := h.orm.User.Get(ctx.Request().Context(), otherUserID)
	if err != nil {
		logger.Error("Failed to get other user", "error", err, "other_user_id", otherUserID)
		return fail(err, "failed to get other user")
	}
	logger.Info("Other user loaded", "other_user_id", otherUser.ID, "other_user_name", otherUser.Name)

	// Get sidebar data (for DM, we need workspace context - get from user's first workspace or use nil)
	logger.Info("Loading sidebar data for DM view")
	// For DM view, we don't have a specific workspace, so we'll get the first workspace or use empty sidebar
	var sidebarData messengerComponents.SidebarData
	workspaces, err := h.orm.WorkspaceMember.
		Query().
		Where(workspacemember.UserIDEQ(int(user.ID))).
		QueryWorkspace().
		Limit(1).
		All(ctx.Request().Context())

	if err == nil && len(workspaces) > 0 {
		activeDMID := &id
		sidebarData, err = h.getSidebarData(ctx, workspaces[0].ID, int(user.ID), nil, activeDMID)
		if err != nil {
			logger.Warn("Failed to load sidebar data for DM view, continuing without sidebar", "error", err)
			sidebarData = messengerComponents.SidebarData{
				Channels:       []messengerComponents.ChannelData{},
				DirectMessages: []messengerComponents.DMData{},
			}
		}
	} else {
		sidebarData = messengerComponents.SidebarData{
			Channels:       []messengerComponents.ChannelData{},
			DirectMessages: []messengerComponents.DMData{},
		}
	}
	logger.Info("Sidebar data prepared", "channels_count", len(sidebarData.Channels), "dms_count", len(sidebarData.DirectMessages))
	ctx.Set(context.MessengerSidebarKey, sidebarData)

	// Get messages (last 50)
	logger.Info("Loading messages for DM", "dm_id", id)
	dmMessages, err := h.orm.DirectMessageContent.
		Query().
		Where(directmessagecontent.DmIDEQ(id)).
		Order(ent.Desc(directmessagecontent.FieldCreatedAt)).
		Limit(50).
		WithUser().
		WithAttachments().
		All(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to load DM messages", "error", err, "dm_id", id)
		return fail(err, "failed to load messages")
	}
	logger.Info("DM messages loaded", "dm_id", id, "messages_count", len(dmMessages))

	// Convert to MessageData
	messageData := make([]messengerComponents.MessageData, len(dmMessages))
	for i, msg := range dmMessages {
		// Convert attachments
		attachmentData := make([]messengerComponents.FileAttachmentData, len(msg.Edges.Attachments))
		for j, att := range msg.Edges.Attachments {
			attachmentURL := fmt.Sprintf("/attachment/%d", att.ID)
			attachmentData[j] = messengerComponents.FileAttachmentData{
				ID:       int64(att.ID),
				Filename: att.Filename,
				MimeType: att.MimeType,
				FileSize: att.FileSize,
				URL:      attachmentURL,
			}
		}

		messageData[i] = messengerComponents.MessageData{
			ID:          int64(msg.ID),
			Content:     msg.Content,
			UserID:      int64(msg.UserID),
			UserName:    msg.Edges.User.Name,
			CreatedAt:   msg.CreatedAt,
			EditedAt:    nil,                                  // DM messages don't have EditedAt
			Reactions:   []messengerComponents.ReactionData{}, // DM messages don't have reactions yet
			Attachments: attachmentData,
		}
	}

	// Reverse to show oldest first
	for i, j := 0, len(messageData)-1; i < j; i, j = i+1, j-1 {
		messageData[i], messageData[j] = messageData[j], messageData[i]
	}
	logger.Info("Messages converted and reversed", "dm_id", id, "messages_count", len(messageData))

	logger.Info("Rendering direct message page", "dm_id", id, "other_user_name", otherUser.Name, "messages_count", len(messageData))
	logger.Info("=== DM VIEW END ===")
	return messengerPages.DirectMessage(ctx, int64(id), otherUser.Name, messageData)
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
	logger := log.Ctx(ctx)
	logger.Info("=== DM MESSAGE CREATE START ===")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logger.Error("Invalid DM ID", "error", err, "id_param", ctx.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid DM ID")
	}
	logger.Info("Creating DM message", "dm_id", id)

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("DM message creator", "user_id", user.ID, "user_name", user.Name)

	// Get DM and verify user is part of it
	logger.Info("Loading DM conversation", "dm_id", id)
	dm, err := h.orm.DirectMessage.Get(ctx.Request().Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warn("DM message create failed: DM not found", "dm_id", id)
			return echo.NewHTTPError(http.StatusNotFound, "direct message not found")
		}
		logger.Error("Failed to get direct message", "error", err, "dm_id", id)
		return fail(err, "failed to get direct message")
	}
	logger.Info("DM loaded", "dm_id", dm.ID, "user1_id", dm.User1ID, "user2_id", dm.User2ID)

	if dm.User1ID != int(user.ID) && dm.User2ID != int(user.ID) {
		logger.Warn("DM message create failed: user not part of conversation", "dm_id", id, "user_id", user.ID)
		return echo.NewHTTPError(http.StatusForbidden, "you are not part of this conversation")
	}
	logger.Info("User verified as part of DM", "dm_id", id, "user_id", user.ID)

	// Parse form data
	content := ctx.FormValue("content")

	// Check if there are files being uploaded
	form, _ := ctx.MultipartForm()
	hasFiles := form != nil && form.File != nil && len(form.File["files"]) > 0

	if content == "" && !hasFiles {
		logger.Warn("DM message create failed: content is required and no files provided", "dm_id", id)
		return echo.NewHTTPError(http.StatusBadRequest, "message content or file is required")
	}
	logger.Info("DM message content", "dm_id", id, "content_length", len(content), "has_files", hasFiles)

	// Создаём сообщение в базе данных
	// DirectMessageContent - это отдельная таблица для сообщений в DM (не Message)
	// Это позволяет иметь разные структуры для каналов и DM
	logger.Info("Creating DM message in database", "dm_id", id, "user_id", user.ID)
	msg, err := h.orm.DirectMessageContent.
		Create().                     // Начинаем создание записи
		SetContent(content).          // Текст сообщения
		SetDmID(id).                  // Связываем с DM
		SetUserID(int(user.ID)).      // Устанавливаем автора
		Save(ctx.Request().Context()) // Сохраняем в БД

	if err != nil {
		logger.Error("Failed to create DM message", "error", err, "dm_id", id, "user_id", user.ID)
		return fail(err, "failed to create message")
	}
	logger.Info("DM message created", "message_id", msg.ID, "dm_id", id, "user_id", user.ID)

	// Handle file attachments (reuse form if already parsed)
	if form == nil {
		form, _ = ctx.MultipartForm()
	}
	if err == nil && form != nil && form.File != nil {
		files := form.File["files"]
		if len(files) > 0 {
			logger.Info("Processing file attachments for DM", "file_count", len(files))
			for _, fileHeader := range files {
				// Validate file size
				maxSize := int64(10485760) // Default 10MB
				if h.config != nil && h.config.Messenger.MaxFileSize > 0 {
					maxSize = h.config.Messenger.MaxFileSize
				}
				if fileHeader.Size > maxSize {
					logger.Warn("File size exceeds maximum", "filename", fileHeader.Filename, "size", fileHeader.Size, "max_size", maxSize)
					continue
				}

				// Validate MIME type
				mimeType := fileHeader.Header.Get("Content-Type")
				if h.config != nil && len(h.config.Messenger.AllowedFileTypes) > 0 {
					allowed := false
					for _, allowedType := range h.config.Messenger.AllowedFileTypes {
						if matchesMimeType(mimeType, allowedType) {
							allowed = true
							break
						}
					}
					if !allowed {
						logger.Warn("File type not allowed", "filename", fileHeader.Filename, "mime_type", mimeType)
						continue
					}
				}

				// Open uploaded file
				src, err := fileHeader.Open()
				if err != nil {
					logger.Error("Failed to open uploaded file", "error", err, "filename", fileHeader.Filename)
					continue
				}

				// Create file path (organized by DM)
				filePath := filepath.Join("attachments", "dm", fmt.Sprintf("%d_%s", id, fileHeader.Filename))
				dst, err := h.files.Create(filePath)
				if err != nil {
					src.Close()
					logger.Error("Failed to create file", "error", err, "filepath", filePath)
					continue
				}

				// Copy file content
				if _, err = io.Copy(dst, src); err != nil {
					src.Close()
					dst.Close()
					logger.Error("Failed to save file", "error", err)
					continue
				}
				src.Close()
				dst.Close()

				// Get file info
				fileInfo, err := h.files.Stat(filePath)
				if err != nil {
					logger.Error("Failed to get file info", "error", err)
					continue
				}

				// Create attachment record (for DM, we use DMContentID)
				_, err = h.orm.Attachment.
					Create().
					SetFilename(fileHeader.Filename).
					SetFilepath(filePath).
					SetFileSize(fileInfo.Size()).
					SetMimeType(mimeType).
					SetDmContentID(msg.ID).
					SetUploadedBy(int(user.ID)).
					Save(ctx.Request().Context())

				if err != nil {
					logger.Error("Failed to create attachment record", "error", err)
				} else {
					logger.Info("DM attachment created", "filename", fileHeader.Filename, "dm_content_id", msg.ID)
				}
			}
		}
	}

	// Update last_message_at
	logger.Info("Updating DM last_message_at", "dm_id", id)
	_, err = h.orm.DirectMessage.
		UpdateOneID(id).
		SetLastMessageAt(time.Now()).
		Save(ctx.Request().Context())

	if err != nil {
		// Log but don't fail
		logger.Warn("Failed to update DM last_message_at", "dm_id", id, "error", err)
	} else {
		logger.Info("DM last_message_at updated", "dm_id", id)
	}

	// Send WebSocket event to other user
	if h.hub != nil {
		otherUserID := dm.User1ID
		if dm.User1ID == int(user.ID) {
			otherUserID = dm.User2ID
		}
		logger.Info("Sending WebSocket event: DM message new", "message_id", msg.ID, "dm_id", id, "other_user_id", otherUserID)

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
		logger.Info("WebSocket event sent for new DM message", "message_id", msg.ID, "dm_id", id, "other_user_id", otherUserID)
	}

	// If HTMX request, return HTML for the new message
	if htmx.GetRequest(ctx).Enabled {
		logger.Info("HTMX request detected, returning HTML message item")
		// Load user for message display
		msgWithUser, err := h.orm.DirectMessageContent.Query().Where(directmessagecontent.IDEQ(msg.ID)).WithUser().Only(ctx.Request().Context())
		if err != nil {
			logger.Warn("Failed to load user for DM message, using basic data", "error", err)
			msgWithUser = msg
		}

		// Load attachments
		msgWithAttachments, err := h.orm.DirectMessageContent.Query().Where(directmessagecontent.IDEQ(msg.ID)).WithAttachments().Only(ctx.Request().Context())
		if err != nil {
			msgWithAttachments = msg
		}

		// Convert attachments
		attachmentData := make([]messengerComponents.FileAttachmentData, len(msgWithAttachments.Edges.Attachments))
		for j, att := range msgWithAttachments.Edges.Attachments {
			attachmentURL := fmt.Sprintf("/attachment/%d", att.ID)
			attachmentData[j] = messengerComponents.FileAttachmentData{
				ID:       int64(att.ID),
				Filename: att.Filename,
				MimeType: att.MimeType,
				FileSize: att.FileSize,
				URL:      attachmentURL,
			}
		}

		// Load message with reply count (DM messages don't have threads, so reply_count is 0)
		// Convert to MessageData
		messageData := messengerComponents.MessageData{
			ID:          int64(msg.ID),
			Content:     msg.Content,
			UserID:      int64(msg.UserID),
			UserName:    msgWithUser.Edges.User.Name,
			CreatedAt:   msg.CreatedAt,
			EditedAt:    nil,                                  // DM messages don't have EditedAt
			Reactions:   []messengerComponents.ReactionData{}, // DM messages don't have reactions
			Attachments: attachmentData,
			ReplyCount:  0, // DM messages don't have threads
		}

		r := ui.NewRequest(ctx)
		messageItem := messengerComponents.MessageItem(r, messageData)
		var buf bytes.Buffer
		if err := messageItem.Render(&buf); err != nil {
			logger.Error("Failed to render DM message item", "error", err)
			return fail(err, "failed to render message")
		}
		logger.Info("=== DM MESSAGE CREATE END ===")
		return ctx.HTML(http.StatusOK, buf.String())
	}

	logger.Info("Non-HTMX request, returning JSON", "message_id", msg.ID)
	logger.Info("=== DM MESSAGE CREATE END ===")
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
	logger := log.Ctx(ctx)
	logger.Info("=== REACTION REMOVE START ===")

	messageID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logger.Error("Invalid message ID", "error", err, "id_param", ctx.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid message ID")
	}
	logger.Info("Removing reaction", "message_id", messageID)

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("Reaction remover", "user_id", user.ID, "user_name", user.Name)

	// Get emoji from URL parameter
	emoji := ctx.Param("emoji")
	if emoji == "" {
		logger.Warn("Reaction remove failed: emoji is required", "message_id", messageID)
		return echo.NewHTTPError(http.StatusBadRequest, "emoji is required")
	}
	logger.Info("Reaction emoji to remove", "message_id", messageID, "emoji", emoji)

	// Get message to find channel ID
	logger.Info("Loading message to find channel ID", "message_id", messageID)
	msg, err := h.orm.Message.Get(ctx.Request().Context(), messageID)
	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warn("Reaction remove failed: message not found", "message_id", messageID)
			return echo.NewHTTPError(http.StatusNotFound, "message not found")
		}
		logger.Error("Failed to get message", "error", err, "message_id", messageID)
		return fail(err, "failed to get message")
	}
	logger.Info("Message found", "message_id", msg.ID, "channel_id", msg.ChannelID)

	// Find and delete reaction
	logger.Info("Deleting reaction from database", "message_id", messageID, "user_id", user.ID, "emoji", emoji)
	_, err = h.orm.Reaction.
		Delete().
		Where(reaction.MessageIDEQ(messageID)).
		Where(reaction.UserIDEQ(int(user.ID))).
		Where(reaction.EmojiEQ(emoji)).
		Exec(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to delete reaction", "error", err, "message_id", messageID, "user_id", user.ID, "emoji", emoji)
		return fail(err, "failed to delete reaction")
	}
	logger.Info("Reaction deleted", "message_id", messageID, "user_id", user.ID, "emoji", emoji)

	// Send WebSocket event
	if h.hub != nil {
		logger.Info("Sending WebSocket event: reaction removed", "message_id", messageID, "channel_id", msg.ChannelID, "emoji", emoji)
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

	logger.Info("=== REACTION REMOVE END ===")
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

	// Validate file size
	maxSize := int64(10485760) // Default 10MB
	if h.config != nil && h.config.Messenger.MaxFileSize > 0 {
		maxSize = h.config.Messenger.MaxFileSize
	}
	if file.Size > maxSize {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("file size exceeds maximum allowed size of %d bytes", maxSize))
	}

	// Validate MIME type
	mimeType := file.Header.Get("Content-Type")
	if h.config != nil && len(h.config.Messenger.AllowedFileTypes) > 0 {
		allowed := false
		for _, allowedType := range h.config.Messenger.AllowedFileTypes {
			if matchesMimeType(mimeType, allowedType) {
				allowed = true
				break
			}
		}
		if !allowed {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("file type %s is not allowed", mimeType))
		}
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

// DMAttachmentUpload uploads a file attachment to a direct message
func (h *Messenger) DMAttachmentUpload(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("=== DM ATTACHMENT UPLOAD START ===")

	dmID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		logger.Error("Invalid DM ID", "error", err, "id_param", ctx.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid DM ID")
	}
	logger.Info("Uploading attachment to DM", "dm_id", dmID)

	user := ctx.Get(context.AuthenticatedUserKey).(*ent.User)
	logger.Info("User uploading attachment", "user_id", user.ID, "user_name", user.Name)

	// Verify DM exists and user is part of it
	dm, err := h.orm.DirectMessage.Get(ctx.Request().Context(), dmID)
	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warn("DM not found", "dm_id", dmID)
			return echo.NewHTTPError(http.StatusNotFound, "direct message not found")
		}
		logger.Error("Failed to get DM", "error", err, "dm_id", dmID)
		return fail(err, "failed to get direct message")
	}

	// Check if user is part of this DM
	if dm.User1ID != int(user.ID) && dm.User2ID != int(user.ID) {
		logger.Warn("User not authorized to upload to this DM", "user_id", user.ID, "dm_id", dmID)
		return echo.NewHTTPError(http.StatusForbidden, "you are not part of this conversation")
	}
	logger.Info("User verified as part of DM", "user_id", user.ID)

	// Get file from form
	file, err := ctx.FormFile("file")
	if err != nil {
		logger.Error("File not provided", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "file is required")
	}
	logger.Info("File received", "filename", file.Filename, "size", file.Size)

	// Validate file size
	maxSize := int64(10485760) // Default 10MB
	if h.config != nil && h.config.Messenger.MaxFileSize > 0 {
		maxSize = h.config.Messenger.MaxFileSize
	}
	if file.Size > maxSize {
		logger.Warn("File size exceeds maximum", "size", file.Size, "max_size", maxSize)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("file size exceeds maximum allowed size of %d bytes", maxSize))
	}

	// Validate MIME type
	mimeType := file.Header.Get("Content-Type")
	if h.config != nil && len(h.config.Messenger.AllowedFileTypes) > 0 {
		allowed := false
		for _, allowedType := range h.config.Messenger.AllowedFileTypes {
			if matchesMimeType(mimeType, allowedType) {
				allowed = true
				break
			}
		}
		if !allowed {
			logger.Warn("File type not allowed", "mime_type", mimeType)
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("file type %s is not allowed", mimeType))
		}
	}
	logger.Info("File validation passed", "mime_type", mimeType)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		logger.Error("Failed to open uploaded file", "error", err)
		return fail(err, "failed to open uploaded file")
	}
	defer src.Close()

	// Create file path in attachments directory (organized by DM)
	filePath := filepath.Join("attachments", "dm", fmt.Sprintf("%d_%s", dmID, file.Filename))
	dst, err := h.files.Create(filePath)
	if err != nil {
		logger.Error("Failed to create file", "error", err, "filepath", filePath)
		return fail(err, "failed to create file")
	}
	defer dst.Close()

	// Copy file content
	if _, err = io.Copy(dst, src); err != nil {
		logger.Error("Failed to save file", "error", err)
		return fail(err, "failed to save file")
	}
	logger.Info("File saved", "filepath", filePath)

	// Get file info
	fileInfo, err := h.files.Stat(filePath)
	if err != nil {
		logger.Error("Failed to get file info", "error", err)
		return fail(err, "failed to get file info")
	}

	// Create attachment record (for DM, we use DMContentID instead of MessageID)
	// For now, we'll create attachment without DMContentID and link it later when message is created
	attachment, err := h.orm.Attachment.
		Create().
		SetFilename(file.Filename).
		SetFilepath(filePath).
		SetFileSize(fileInfo.Size()).
		SetMimeType(mimeType).
		SetNillableDmContentID(nil). // Will be set when DM message is created
		SetUploadedBy(int(user.ID)).
		Save(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to create attachment record", "error", err)
		return fail(err, "failed to create attachment record")
	}

	logger.Info("=== DM ATTACHMENT UPLOAD END ===", "attachment_id", attachment.ID)
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

// matchesMimeType checks if a MIME type matches a pattern (supports wildcards like "image/*")
func matchesMimeType(mimeType, pattern string) bool {
	if pattern == "*" || pattern == "*/*" {
		return true
	}
	if mimeType == pattern {
		return true
	}
	// Check wildcard patterns like "image/*"
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(mimeType, prefix+"/")
	}
	return false
}

// ============================================================================
// CUSTOM CODE END
// ============================================================================
