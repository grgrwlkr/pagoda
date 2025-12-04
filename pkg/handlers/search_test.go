package handlers

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/mikestefanello/pagoda/pkg/routenames"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSearchPage tests the search page handler
func TestSearchPage(t *testing.T) {
	ctx := context.Background()

	// Arrange
	usr := createTestUser(t, "search-page@example.com", "Search Page", "password123")
	defer func() {
		workspaces, _ := c.ORM.Workspace.Query().All(ctx)
		for _, ws := range workspaces {
			c.ORM.WorkspaceMember.Delete().ExecX(ctx)
			c.ORM.Workspace.DeleteOneID(ws.ID).ExecX(ctx)
		}
		c.ORM.User.DeleteOneID(usr.ID).ExecX(ctx)
	}()

	// Create workspace
	workspace, err := c.ORM.Workspace.Create().
		SetName("Test Workspace").
		SetSlug("test-workspace").
		SetOwnerID(int(usr.ID)).
		Save(ctx)
	require.NoError(t, err)

	_, err = c.ORM.WorkspaceMember.Create().
		SetWorkspaceID(workspace.ID).
		SetUserID(int(usr.ID)).
		SetRole("member").
		Save(ctx)
	require.NoError(t, err)

	client := authenticateUser(t, "search-page@example.com", "password123")

	// Act
	searchURL := srv.URL + c.Web.Reverse(routenames.MessengerSearch) + "?q=test&workspace_id=" + strconv.Itoa(workspace.ID)
	resp, err := client.Get(searchURL)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

// TestSearchPage_Unauthorized tests search page without workspace membership
func TestSearchPage_Unauthorized(t *testing.T) {
	ctx := context.Background()

	// Arrange
	usr1 := createTestUser(t, "search-unauth1@example.com", "User 1", "password123")
	usr2 := createTestUser(t, "search-unauth2@example.com", "User 2", "password123")
	defer func() {
		workspaces, _ := c.ORM.Workspace.Query().All(ctx)
		for _, ws := range workspaces {
			c.ORM.WorkspaceMember.Delete().ExecX(ctx)
			c.ORM.Workspace.DeleteOneID(ws.ID).ExecX(ctx)
		}
		c.ORM.User.DeleteOneID(usr1.ID).ExecX(ctx)
		c.ORM.User.DeleteOneID(usr2.ID).ExecX(ctx)
	}()

	// Create workspace for user1
	workspace, err := c.ORM.Workspace.Create().
		SetName("Private Workspace").
		SetSlug("private-ws").
		SetOwnerID(int(usr1.ID)).
		Save(ctx)
	require.NoError(t, err)

	_, err = c.ORM.WorkspaceMember.Create().
		SetWorkspaceID(workspace.ID).
		SetUserID(int(usr1.ID)).
		SetRole("owner").
		Save(ctx)
	require.NoError(t, err)

	// Authenticate as user2 (not a member)
	client := authenticateUser(t, "search-unauth2@example.com", "password123")

	// Act
	searchURL := srv.URL + c.Web.Reverse(routenames.MessengerSearch) + "?q=test&workspace_id=" + strconv.Itoa(workspace.ID)
	resp, err := client.Get(searchURL)

	// Assert
	require.NoError(t, err)
	// Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

// TestSearchMessages tests message search
func TestSearchMessages(t *testing.T) {
	ctx := context.Background()

	// Arrange
	usr := createTestUser(t, "search-msgs@example.com", "Search Messages", "password123")
	defer func() {
		messages, _ := c.ORM.Message.Query().All(ctx)
		for _, msg := range messages {
			c.ORM.Message.DeleteOneID(msg.ID).ExecX(ctx)
		}
		channels, _ := c.ORM.Channel.Query().All(ctx)
		for _, ch := range channels {
			c.ORM.ChannelMember.Delete().ExecX(ctx)
			c.ORM.Channel.DeleteOneID(ch.ID).ExecX(ctx)
		}
		workspaces, _ := c.ORM.Workspace.Query().All(ctx)
		for _, ws := range workspaces {
			c.ORM.WorkspaceMember.Delete().ExecX(ctx)
			c.ORM.Workspace.DeleteOneID(ws.ID).ExecX(ctx)
		}
		c.ORM.User.DeleteOneID(usr.ID).ExecX(ctx)
	}()

	// Create workspace and channel
	workspace, err := c.ORM.Workspace.Create().
		SetName("Test Workspace").
		SetSlug("test-workspace").
		SetOwnerID(int(usr.ID)).
		Save(ctx)
	require.NoError(t, err)

	_, err = c.ORM.WorkspaceMember.Create().
		SetWorkspaceID(workspace.ID).
		SetUserID(int(usr.ID)).
		SetRole("member").
		Save(ctx)
	require.NoError(t, err)

	channel, err := c.ORM.Channel.Create().
		SetName("Test Channel").
		SetSlug("test-channel").
		SetWorkspaceID(workspace.ID).
		SetCreatedBy(int(usr.ID)).
		Save(ctx)
	require.NoError(t, err)

	_, err = c.ORM.ChannelMember.Create().
		SetChannelID(channel.ID).
		SetUserID(int(usr.ID)).
		Save(ctx)
	require.NoError(t, err)

	// Create a test message
	_, err = c.ORM.Message.Create().
		SetContent("This is a test message for search").
		SetChannelID(channel.ID).
		SetUserID(int(usr.ID)).
		Save(ctx)
	require.NoError(t, err)

	client := authenticateUser(t, "search-msgs@example.com", "password123")

	// Act
	searchURL := srv.URL + c.Web.Reverse(routenames.MessengerSearchMessages) + "?q=test&workspace_id=" + strconv.Itoa(workspace.ID)
	resp, err := client.Get(searchURL)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

// TestSearchMessages_NoQuery tests message search without query parameter
func TestSearchMessages_NoQuery(t *testing.T) {
	// Arrange
	usr := createTestUser(t, "search-noq@example.com", "Search No Query", "password123")
	defer func() {
		ctx := context.Background()
		c.ORM.User.DeleteOneID(usr.ID).ExecX(ctx)
	}()

	client := authenticateUser(t, "search-noq@example.com", "password123")

	// Act
	searchURL := srv.URL + c.Web.Reverse(routenames.MessengerSearchMessages)
	resp, err := client.Get(searchURL)

	// Assert
	require.NoError(t, err)
	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

// TestSearchMessages_Unauthorized tests message search without workspace membership
func TestSearchMessages_Unauthorized(t *testing.T) {
	ctx := context.Background()

	// Arrange
	usr1 := createTestUser(t, "search-msg-unauth1@example.com", "User 1", "password123")
	usr2 := createTestUser(t, "search-msg-unauth2@example.com", "User 2", "password123")
	defer func() {
		workspaces, _ := c.ORM.Workspace.Query().All(ctx)
		for _, ws := range workspaces {
			c.ORM.WorkspaceMember.Delete().ExecX(ctx)
			c.ORM.Workspace.DeleteOneID(ws.ID).ExecX(ctx)
		}
		c.ORM.User.DeleteOneID(usr1.ID).ExecX(ctx)
		c.ORM.User.DeleteOneID(usr2.ID).ExecX(ctx)
	}()

	// Create workspace for user1
	workspace, err := c.ORM.Workspace.Create().
		SetName("Private Workspace").
		SetSlug("private-ws").
		SetOwnerID(int(usr1.ID)).
		Save(ctx)
	require.NoError(t, err)

	_, err = c.ORM.WorkspaceMember.Create().
		SetWorkspaceID(workspace.ID).
		SetUserID(int(usr1.ID)).
		SetRole("owner").
		Save(ctx)
	require.NoError(t, err)

	// Authenticate as user2 (not a member)
	client := authenticateUser(t, "search-msg-unauth2@example.com", "password123")

	// Act
	searchURL := srv.URL + c.Web.Reverse(routenames.MessengerSearchMessages) + "?q=test&workspace_id=" + strconv.Itoa(workspace.ID)
	resp, err := client.Get(searchURL)

	// Assert
	require.NoError(t, err)
	// Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

// TestSearchUsers tests user search
func TestSearchUsers(t *testing.T) {
	ctx := context.Background()

	// Arrange
	usr1 := createTestUser(t, "search-users1@example.com", "User One", "password123")
	usr2 := createTestUser(t, "search-users2@example.com", "User Two", "password123")
	defer func() {
		workspaces, _ := c.ORM.Workspace.Query().All(ctx)
		for _, ws := range workspaces {
			c.ORM.WorkspaceMember.Delete().ExecX(ctx)
			c.ORM.Workspace.DeleteOneID(ws.ID).ExecX(ctx)
		}
		c.ORM.User.DeleteOneID(usr1.ID).ExecX(ctx)
		c.ORM.User.DeleteOneID(usr2.ID).ExecX(ctx)
	}()

	// Create workspace
	workspace, err := c.ORM.Workspace.Create().
		SetName("Test Workspace").
		SetSlug("test-workspace").
		SetOwnerID(int(usr1.ID)).
		Save(ctx)
	require.NoError(t, err)

	// Add both users as workspace members
	_, err = c.ORM.WorkspaceMember.Create().
		SetWorkspaceID(workspace.ID).
		SetUserID(int(usr1.ID)).
		SetRole("member").
		Save(ctx)
	require.NoError(t, err)

	_, err = c.ORM.WorkspaceMember.Create().
		SetWorkspaceID(workspace.ID).
		SetUserID(int(usr2.ID)).
		SetRole("member").
		Save(ctx)
	require.NoError(t, err)

	client := authenticateUser(t, "search-users1@example.com", "password123")

	// Act
	searchURL := srv.URL + c.Web.Reverse(routenames.MessengerSearchUsers) + "?q=Two&workspace_id=" + strconv.Itoa(workspace.ID)
	resp, err := client.Get(searchURL)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

// TestSearchUsers_NoQuery tests user search without query parameter
func TestSearchUsers_NoQuery(t *testing.T) {
	ctx := context.Background()

	// Arrange
	usr := createTestUser(t, "search-users-noq@example.com", "Search Users", "password123")
	defer func() {
		workspaces, _ := c.ORM.Workspace.Query().All(ctx)
		for _, ws := range workspaces {
			c.ORM.WorkspaceMember.Delete().ExecX(ctx)
			c.ORM.Workspace.DeleteOneID(ws.ID).ExecX(ctx)
		}
		c.ORM.User.DeleteOneID(usr.ID).ExecX(ctx)
	}()

	// Create workspace
	workspace, err := c.ORM.Workspace.Create().
		SetName("Test Workspace").
		SetSlug("test-workspace").
		SetOwnerID(int(usr.ID)).
		Save(ctx)
	require.NoError(t, err)

	_, err = c.ORM.WorkspaceMember.Create().
		SetWorkspaceID(workspace.ID).
		SetUserID(int(usr.ID)).
		SetRole("member").
		Save(ctx)
	require.NoError(t, err)

	client := authenticateUser(t, "search-users-noq@example.com", "password123")

	// Act
	searchURL := srv.URL + c.Web.Reverse(routenames.MessengerSearchUsers) + "?workspace_id=" + strconv.Itoa(workspace.ID)
	resp, err := client.Get(searchURL)

	// Assert
	require.NoError(t, err)
	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

// TestSearchUsers_NoWorkspaceID tests user search without workspace_id parameter
func TestSearchUsers_NoWorkspaceID(t *testing.T) {
	// Arrange
	usr := createTestUser(t, "search-users-nows@example.com", "Search Users", "password123")
	defer func() {
		ctx := context.Background()
		c.ORM.User.DeleteOneID(usr.ID).ExecX(ctx)
	}()

	client := authenticateUser(t, "search-users-nows@example.com", "password123")

	// Act
	searchURL := srv.URL + c.Web.Reverse(routenames.MessengerSearchUsers) + "?q=test"
	resp, err := client.Get(searchURL)

	// Assert
	require.NoError(t, err)
	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

// TestSearchUsers_Unauthorized tests user search without workspace membership
func TestSearchUsers_Unauthorized(t *testing.T) {
	ctx := context.Background()

	// Arrange
	usr1 := createTestUser(t, "search-users-unauth1@example.com", "User 1", "password123")
	usr2 := createTestUser(t, "search-users-unauth2@example.com", "User 2", "password123")
	defer func() {
		workspaces, _ := c.ORM.Workspace.Query().All(ctx)
		for _, ws := range workspaces {
			c.ORM.WorkspaceMember.Delete().ExecX(ctx)
			c.ORM.Workspace.DeleteOneID(ws.ID).ExecX(ctx)
		}
		c.ORM.User.DeleteOneID(usr1.ID).ExecX(ctx)
		c.ORM.User.DeleteOneID(usr2.ID).ExecX(ctx)
	}()

	// Create workspace for user1
	workspace, err := c.ORM.Workspace.Create().
		SetName("Private Workspace").
		SetSlug("private-ws").
		SetOwnerID(int(usr1.ID)).
		Save(ctx)
	require.NoError(t, err)

	_, err = c.ORM.WorkspaceMember.Create().
		SetWorkspaceID(workspace.ID).
		SetUserID(int(usr1.ID)).
		SetRole("owner").
		Save(ctx)
	require.NoError(t, err)

	// Authenticate as user2 (not a member)
	client := authenticateUser(t, "search-users-unauth2@example.com", "password123")

	// Act
	searchURL := srv.URL + c.Web.Reverse(routenames.MessengerSearchUsers) + "?q=test&workspace_id=" + strconv.Itoa(workspace.ID)
	resp, err := client.Get(searchURL)

	// Assert
	require.NoError(t, err)
	// Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}
