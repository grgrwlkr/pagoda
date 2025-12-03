package handlers

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/ent/channel"
	"github.com/mikestefanello/pagoda/ent/message"
	"github.com/mikestefanello/pagoda/ent/user"
	"github.com/mikestefanello/pagoda/ent/workspacemember"
	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/log"
	messengerMiddleware "github.com/mikestefanello/pagoda/pkg/middleware"
	"github.com/mikestefanello/pagoda/pkg/pager"
	"github.com/mikestefanello/pagoda/pkg/routenames"
	"github.com/mikestefanello/pagoda/pkg/services"
	"github.com/mikestefanello/pagoda/pkg/ui"
	messengerComponents "github.com/mikestefanello/pagoda/pkg/ui/components/messenger"
	messengerPages "github.com/mikestefanello/pagoda/pkg/ui/pages/messenger"
)

type Search struct {
	orm *ent.Client
}

func init() {
	Register(new(Search))
}

func (h *Search) Init(c *services.Container) error {
	h.orm = c.ORM
	return nil
}

func (h *Search) Routes(g *echo.Group) {
	// All search routes require authentication
	g = g.Group("", messengerMiddleware.RequireAuthentication)

	// Search endpoints
	g.GET("/search/messages", h.SearchMessages).Name = routenames.MessengerSearchMessages
	g.GET("/search/users", h.SearchUsers).Name = routenames.MessengerSearchUsers
	g.GET("/search", h.SearchPage).Name = routenames.MessengerSearch
}

// Note: MessageSearchResult and UserSearchResult are defined in messengerComponents package

// SearchMessages searches for messages
func (h *Search) SearchMessages(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("Searching messages")

	// Get authenticated user
	authUser := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get query parameters
	query := strings.TrimSpace(ctx.QueryParam("q"))
	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "query parameter 'q' is required")
	}

	// Build query
	q := h.orm.Message.Query().
		Where(message.ContentContains(query))

	// Filter by workspace (user must be a member)
	if workspaceIDStr := ctx.QueryParam("workspace_id"); workspaceIDStr != "" {
		workspaceID, err := strconv.Atoi(workspaceIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace_id")
		}

		// Check if user is a member of this workspace
		isMember, err := h.orm.WorkspaceMember.
			Query().
			Where(
				workspacemember.WorkspaceIDEQ(workspaceID),
				workspacemember.UserIDEQ(authUser.ID),
			).
			Exist(ctx.Request().Context())

		if err != nil {
			logger.Error("Failed to check workspace membership", "error", err)
			return fail(err, "failed to check workspace membership")
		}

		if !isMember {
			return echo.NewHTTPError(http.StatusForbidden, "user is not a member of this workspace")
		}

		// Filter messages by channels in this workspace
		q = q.Where(message.HasChannelWith(
			channel.WorkspaceIDEQ(workspaceID),
		))
	}

	// Filter by channel
	if channelIDStr := ctx.QueryParam("channel_id"); channelIDStr != "" {
		channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid channel_id")
		}
		q = q.Where(message.ChannelIDEQ(int(channelID)))
	}

	// Filter by user (author)
	if userIDStr := ctx.QueryParam("user_id"); userIDStr != "" {
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
		}
		q = q.Where(message.UserIDEQ(userID))
	}

	// Filter by date range
	if fromDateStr := ctx.QueryParam("from_date"); fromDateStr != "" {
		fromDate, err := time.Parse("2006-01-02", fromDateStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid from_date format (expected YYYY-MM-DD)")
		}
		q = q.Where(message.CreatedAtGTE(fromDate))
	}

	if toDateStr := ctx.QueryParam("to_date"); toDateStr != "" {
		toDate, err := time.Parse("2006-01-02", toDateStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid to_date format (expected YYYY-MM-DD)")
		}
		// Add one day to include the entire day
		toDate = toDate.Add(24 * time.Hour)
		q = q.Where(message.CreatedAtLTE(toDate))
	}

	// Pagination
	pgr := pager.NewPager(ctx, 20)
	offset := pgr.GetOffset()
	limit := pgr.ItemsPerPage

	// Get total count for pagination
	total, err := q.Clone().Count(ctx.Request().Context())
	if err != nil {
		logger.Error("Failed to count messages", "error", err)
		return fail(err, "failed to count messages")
	}
	pgr.SetItems(total)

	// Get messages with channel and user info
	messages, err := q.
		WithChannel().
		WithUser().
		Order(ent.Desc(message.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to search messages", "error", err)
		return fail(err, "failed to search messages")
	}

	// Build results
	results := make([]messengerComponents.MessageSearchResult, len(messages))
	for i, msg := range messages {
		results[i] = messengerComponents.MessageSearchResult{
			MessageID:   int64(msg.ID),
			Content:     msg.Content,
			ChannelID:   int64(msg.Edges.Channel.ID),
			ChannelName: msg.Edges.Channel.Name,
			UserID:      msg.Edges.User.ID,
			UserName:    msg.Edges.User.Name,
			CreatedAt:   msg.CreatedAt,
		}
	}

	logger.Info("Message search completed", "query", query, "results", len(results), "total", total)

	// Return JSON or HTML based on request
	if ctx.Request().Header.Get("HX-Request") != "" {
		// HTMX request - return HTML
		r := ui.NewRequest(ctx)

		// Convert to component types
		messageResults := make([]messengerComponents.MessageSearchResult, len(results))
		for i, res := range results {
			messageResults[i] = messengerComponents.MessageSearchResult{
				MessageID:   res.MessageID,
				Content:     res.Content,
				ChannelID:   res.ChannelID,
				ChannelName: res.ChannelName,
				UserID:      res.UserID,
				UserName:    res.UserName,
				CreatedAt:   res.CreatedAt,
			}
		}

		// Render search results
		var buf bytes.Buffer
		if err := messengerComponents.SearchResults(r, messageResults, []messengerComponents.UserSearchResult{}, query).Render(&buf); err != nil {
			logger.Error("Failed to render search results", "error", err)
			return fail(err, "failed to render search results")
		}

		return ctx.HTML(http.StatusOK, buf.String())
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"results": results,
		"pager":   pgr,
	})
}

// SearchUsers searches for users in a workspace
func (h *Search) SearchUsers(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("Searching users")

	// Get authenticated user
	authUser := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	// Get query parameters
	query := strings.TrimSpace(ctx.QueryParam("q"))
	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "query parameter 'q' is required")
	}

	workspaceIDStr := ctx.QueryParam("workspace_id")
	if workspaceIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "workspace_id parameter is required")
	}

	workspaceID, err := strconv.Atoi(workspaceIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace_id")
	}

	// Check if user is a member of this workspace
	isMember, err := h.orm.WorkspaceMember.
		Query().
		Where(
			workspacemember.WorkspaceIDEQ(workspaceID),
			workspacemember.UserIDEQ(authUser.ID),
		).
		Exist(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to check workspace membership", "error", err)
		return fail(err, "failed to check workspace membership")
	}

	if !isMember {
		return echo.NewHTTPError(http.StatusForbidden, "user is not a member of this workspace")
	}

	// Search users who are members of this workspace
	// Search by name or email
	workspaceMembers, err := h.orm.WorkspaceMember.
		Query().
		Where(workspacemember.WorkspaceIDEQ(workspaceID)).
		WithUser(func(uq *ent.UserQuery) {
			uq.Where(
				user.Or(
					user.NameContains(query),
					user.EmailContains(query),
				),
			)
		}).
		All(ctx.Request().Context())

	if err != nil {
		logger.Error("Failed to search users", "error", err)
		return fail(err, "failed to search users")
	}

	// Build results
	results := make([]messengerComponents.UserSearchResult, 0, len(workspaceMembers))
	for _, member := range workspaceMembers {
		if member.Edges.User != nil {
			results = append(results, messengerComponents.UserSearchResult{
				UserID:    member.Edges.User.ID,
				Name:      member.Edges.User.Name,
				Email:     member.Edges.User.Email,
				AvatarURL: "", // TODO: Get from UserProfile when implemented
			})
		}
	}

	// Limit results
	limit := 20
	if limitStr := ctx.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if len(results) > limit {
		results = results[:limit]
	}

	logger.Info("User search completed", "query", query, "workspace_id", workspaceID, "results", len(results))

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"results": results,
	})
}

// SearchPage renders the search results page
func (h *Search) SearchPage(ctx echo.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("Rendering search page")

	// Get query parameters
	query := strings.TrimSpace(ctx.QueryParam("q"))
	workspaceIDStr := ctx.QueryParam("workspace_id")

	// Get authenticated user
	authUser := ctx.Get(context.AuthenticatedUserKey).(*ent.User)

	var workspaceID int64
	var workspaceName string
	if workspaceIDStr != "" {
		id, err := strconv.ParseInt(workspaceIDStr, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid workspace_id")
		}
		workspaceID = id

		// Check if user is a member
		isMember, err := h.orm.WorkspaceMember.
			Query().
			Where(
				workspacemember.WorkspaceIDEQ(int(workspaceID)),
				workspacemember.UserIDEQ(authUser.ID),
			).
			Exist(ctx.Request().Context())

		if err != nil {
			return fail(err, "failed to check workspace membership")
		}

		if !isMember {
			return echo.NewHTTPError(http.StatusForbidden, "user is not a member of this workspace")
		}

		// Get workspace name
		ws, err := h.orm.Workspace.Get(ctx.Request().Context(), int(workspaceID))
		if err == nil {
			workspaceName = ws.Name
		}
	}

	// Get sidebar data if workspace is specified
	var sidebarData messengerComponents.SidebarData
	if workspaceID > 0 {
		// Load sidebar data (simplified - just workspace info)
		sidebarData = messengerComponents.SidebarData{
			WorkspaceID:    workspaceID,
			WorkspaceName:  workspaceName,
			Channels:       []messengerComponents.ChannelData{},
			DirectMessages: []messengerComponents.DMData{},
		}
	}

	// Store sidebar data in context
	ctx.Set(context.MessengerSidebarKey, sidebarData)

	// Search messages and users if query provided
	var messageResults []messengerComponents.MessageSearchResult
	var userResults []messengerComponents.UserSearchResult

	if query != "" {
		// Search messages
		msgResults, err := h.searchMessagesInternal(ctx, query, int(workspaceID), 0, 0, "", "", 0, 50)
		if err == nil {
			messageResults = msgResults
		}

		// Search users if workspace is specified
		if workspaceID > 0 {
			usrResults, err := h.searchUsersInternal(ctx, query, int(workspaceID), 50)
			if err == nil {
				userResults = usrResults
			}
		}
	}

	return messengerPages.SearchPage(ctx, query, workspaceID, messageResults, userResults)
}

// searchMessagesInternal performs message search and returns results
func (h *Search) searchMessagesInternal(ctx echo.Context, query string, workspaceID, channelID, userID int, fromDate, toDate string, offset, limit int) ([]messengerComponents.MessageSearchResult, error) {
	q := h.orm.Message.Query().
		Where(message.ContentContains(query))

	if workspaceID > 0 {
		q = q.Where(message.HasChannelWith(
			channel.WorkspaceIDEQ(workspaceID),
		))
	}

	if channelID > 0 {
		q = q.Where(message.ChannelIDEQ(channelID))
	}

	if userID > 0 {
		q = q.Where(message.UserIDEQ(userID))
	}

	// Date filters
	if fromDate != "" {
		from, err := time.Parse("2006-01-02", fromDate)
		if err == nil {
			q = q.Where(message.CreatedAtGTE(from))
		}
	}

	if toDate != "" {
		to, err := time.Parse("2006-01-02", toDate)
		if err == nil {
			to = to.Add(24 * time.Hour)
			q = q.Where(message.CreatedAtLTE(to))
		}
	}

	messages, err := q.
		WithChannel().
		WithUser().
		Order(ent.Desc(message.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx.Request().Context())

	if err != nil {
		return nil, err
	}

	results := make([]messengerComponents.MessageSearchResult, len(messages))
	for i, msg := range messages {
		results[i] = messengerComponents.MessageSearchResult{
			MessageID:   int64(msg.ID),
			Content:     msg.Content,
			ChannelID:   int64(msg.Edges.Channel.ID),
			ChannelName: msg.Edges.Channel.Name,
			UserID:      msg.Edges.User.ID,
			UserName:    msg.Edges.User.Name,
			CreatedAt:   msg.CreatedAt,
		}
	}

	return results, nil
}

// searchUsersInternal performs user search and returns results
func (h *Search) searchUsersInternal(ctx echo.Context, query string, workspaceID, limit int) ([]messengerComponents.UserSearchResult, error) {
	workspaceMembers, err := h.orm.WorkspaceMember.
		Query().
		Where(workspacemember.WorkspaceIDEQ(workspaceID)).
		WithUser(func(uq *ent.UserQuery) {
			uq.Where(
				user.Or(
					user.NameContains(query),
					user.EmailContains(query),
				),
			)
		}).
		All(ctx.Request().Context())

	if err != nil {
		return nil, err
	}

	results := make([]messengerComponents.UserSearchResult, 0, len(workspaceMembers))
	for _, member := range workspaceMembers {
		if member.Edges.User != nil {
			results = append(results, messengerComponents.UserSearchResult{
				UserID:    member.Edges.User.ID,
				Name:      member.Edges.User.Name,
				Email:     member.Edges.User.Email,
				AvatarURL: "",
			})
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}
