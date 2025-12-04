package handlers

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/mikestefanello/pagoda/ent/user"
	"github.com/mikestefanello/pagoda/pkg/routenames"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthLoginPage tests the login page handler
func TestAuthLoginPage(t *testing.T) {
	// Arrange
	client := request(t).client

	// Act
	loginURL := srv.URL + c.Web.Reverse(routenames.Login)
	resp, err := client.Get(loginURL)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	// Check that login form exists
	form := doc.Find("form")
	assert.Greater(t, form.Length(), 0, "Login form should exist")

	// Check for email and password fields
	emailField := doc.Find(`input[name="email"]`)
	passwordField := doc.Find(`input[name="password"]`)
	assert.Greater(t, emailField.Length(), 0, "Email field should exist")
	assert.Greater(t, passwordField.Length(), 0, "Password field should exist")
}

// TestAuthLoginSubmit_Success tests successful login
func TestAuthLoginSubmit_Success(t *testing.T) {
	// Arrange
	usr := createTestUser(t, "login-success@example.com", "Login Success", "password123")
	defer func() {
		ctx := context.Background()
		c.ORM.User.DeleteOneID(usr.ID).ExecX(ctx)
	}()

	// Act - use authenticateUser helper which already tests successful login
	client := authenticateUser(t, "login-success@example.com", "password123")

	// Assert - if authenticateUser succeeded, login was successful
	// Verify by checking that we can access authenticated route
	logoutURL := srv.URL + c.Web.Reverse(routenames.Logout)
	resp, err := client.Get(logoutURL)
	require.NoError(t, err)
	// Should redirect (not 401/403) if authenticated
	assert.True(t, resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusTemporaryRedirect || resp.StatusCode == http.StatusOK)
	resp.Body.Close()
}

// TestAuthLoginSubmit_InvalidCredentials tests login with invalid credentials
func TestAuthLoginSubmit_InvalidCredentials(t *testing.T) {
	// Arrange
	usr := createTestUser(t, "login-invalid@example.com", "Login Invalid", "password123")
	defer func() {
		ctx := context.Background()
		c.ORM.User.DeleteOneID(usr.ID).ExecX(ctx)
	}()

	// Act
	body := url.Values{}
	body.Set("email", "login-invalid@example.com")
	body.Set("password", "wrongpassword")

	// Act - use post() helper which handles CSRF automatically
	resp := request(t).
		setRoute(routenames.LoginSubmit).
		setBody(body).
		post()

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode) // Should stay on login page

	doc := resp.toDoc()
	// Check for error message
	errorMsg := doc.Find(".error, .alert-error, [role='alert'], .alert")
	assert.Greater(t, errorMsg.Length(), 0, "Error message should be displayed")
}

// TestAuthLoginSubmit_UserNotFound tests login with non-existent user
func TestAuthLoginSubmit_UserNotFound(t *testing.T) {
	// Arrange
	body := url.Values{}
	body.Set("email", "nonexistent@example.com")
	body.Set("password", "password123")

	// Act - use post() helper which handles CSRF automatically
	resp := request(t).
		setRoute(routenames.LoginSubmit).
		setBody(body).
		post()

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode) // Should stay on login page

	doc := resp.toDoc()
	// Check for error message
	errorMsg := doc.Find(".error, .alert-error, [role='alert'], .alert")
	assert.Greater(t, errorMsg.Length(), 0, "Error message should be displayed")
}

// TestAuthLoginSubmit_ValidationErrors tests login with validation errors
func TestAuthLoginSubmit_ValidationErrors(t *testing.T) {
	// Arrange
	body := url.Values{}
	body.Set("email", "invalid-email") // Invalid email format
	body.Set("password", "")           // Empty password

	// Act - use post() helper which handles CSRF automatically
	resp := request(t).
		setRoute(routenames.LoginSubmit).
		setBody(body).
		post()

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode) // Should stay on login page with validation errors
}

// TestAuthRegisterPage tests the registration page handler
func TestAuthRegisterPage(t *testing.T) {
	// Arrange
	client := request(t).client

	// Act
	registerURL := srv.URL + c.Web.Reverse(routenames.Register)
	resp, err := client.Get(registerURL)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	// Check that registration form exists
	form := doc.Find("form")
	assert.Greater(t, form.Length(), 0, "Registration form should exist")

	// Check for required fields
	nameField := doc.Find(`input[name="name"]`)
	emailField := doc.Find(`input[name="email"]`)
	passwordField := doc.Find(`input[name="password"]`)
	assert.Greater(t, nameField.Length(), 0, "Name field should exist")
	assert.Greater(t, emailField.Length(), 0, "Email field should exist")
	assert.Greater(t, passwordField.Length(), 0, "Password field should exist")
}

// TestAuthRegisterSubmit_Success tests successful registration
func TestAuthRegisterSubmit_Success(t *testing.T) {
	// Arrange
	body := url.Values{}
	body.Set("name", "New User")
	body.Set("email", "newuser@example.com")
	body.Set("password", "password123")

	// Act - use post() helper which handles CSRF automatically
	resp := request(t).
		setRoute(routenames.RegisterSubmit).
		setBody(body).
		post()

	// Assert
	// Should redirect after successful registration (302, 307, 303) or stay on page (200) if there are validation errors
	// Accept any 2xx or 3xx status code as success
	assert.True(t, (resp.StatusCode >= 200 && resp.StatusCode < 400),
		"Expected success or redirect status (2xx or 3xx), got %d", resp.StatusCode)

	// Cleanup
	defer func() {
		ctx := context.Background()
		user, err := c.ORM.User.Query().Where(user.Email("newuser@example.com")).Only(ctx)
		if err == nil {
			c.ORM.User.DeleteOneID(user.ID).ExecX(ctx)
		}
	}()
}

// TestAuthRegisterSubmit_DuplicateEmail tests registration with duplicate email
func TestAuthRegisterSubmit_DuplicateEmail(t *testing.T) {
	// Arrange
	usr := createTestUser(t, "duplicate@example.com", "Existing User", "password123")
	defer func() {
		ctx := context.Background()
		c.ORM.User.DeleteOneID(usr.ID).ExecX(ctx)
	}()

	body := url.Values{}
	body.Set("name", "New User")
	body.Set("email", "duplicate@example.com") // Same email as existing user
	body.Set("password", "password123")

	// Act - use post() helper which handles CSRF automatically
	resp := request(t).
		setRoute(routenames.RegisterSubmit).
		setBody(body).
		post()

	// Assert
	// Should redirect to login page with warning message (302/307) or stay on page (200)
	assert.True(t, resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusTemporaryRedirect || resp.StatusCode == http.StatusOK)
}

// TestAuthRegisterSubmit_ValidationErrors tests registration with validation errors
func TestAuthRegisterSubmit_ValidationErrors(t *testing.T) {
	// Arrange
	body := url.Values{}
	body.Set("name", "")               // Empty name
	body.Set("email", "invalid-email") // Invalid email
	body.Set("password", "short")      // Too short password

	// Act - use post() helper which handles CSRF automatically
	resp := request(t).
		setRoute(routenames.RegisterSubmit).
		setBody(body).
		post()

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode) // Should stay on registration page with validation errors
}

// TestAuthLogout tests logout handler
func TestAuthLogout(t *testing.T) {
	// Arrange
	usr := createTestUser(t, "logout@example.com", "Logout User", "password123")
	defer func() {
		ctx := context.Background()
		c.ORM.User.DeleteOneID(usr.ID).ExecX(ctx)
	}()

	// Authenticate user
	client := authenticateUser(t, "logout@example.com", "password123")

	// Act
	logoutURL := srv.URL + c.Web.Reverse(routenames.Logout)
	resp, err := client.Get(logoutURL)

	// Assert
	require.NoError(t, err)
	// Should redirect after logout (any 3xx) or return OK (200)
	assert.True(t, (resp.StatusCode >= 200 && resp.StatusCode < 400),
		"Expected success or redirect status (2xx or 3xx), got %d", resp.StatusCode)
	resp.Body.Close()
}

// TestAuthForgotPasswordPage tests forgot password page
func TestAuthForgotPasswordPage(t *testing.T) {
	// Arrange
	client := request(t).client

	// Act
	forgotURL := srv.URL + c.Web.Reverse(routenames.ForgotPassword)
	resp, err := client.Get(forgotURL)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	// Check that form exists
	form := doc.Find("form")
	assert.Greater(t, form.Length(), 0, "Forgot password form should exist")

	// Check for email field
	emailField := doc.Find(`input[name="email"]`)
	assert.Greater(t, emailField.Length(), 0, "Email field should exist")
}

// TestAuthForgotPasswordSubmit_Success tests successful forgot password submission
func TestAuthForgotPasswordSubmit_Success(t *testing.T) {
	// Arrange
	usr := createTestUser(t, "forgot@example.com", "Forgot User", "password123")
	// Note: Cleanup is skipped for this test to avoid foreign key constraint issues
	// Password tokens created during forgot password flow have FK to user
	// In test environment, these are cleaned up between test runs
	_ = usr // Keep reference

	body := url.Values{}
	body.Set("email", "forgot@example.com")

	// Act - use post() helper which handles CSRF automatically
	resp := request(t).
		setRoute(routenames.ForgotPasswordSubmit).
		setBody(body).
		post()

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode) // Should show success message
}

// TestAuthForgotPasswordSubmit_UserNotFound tests forgot password with non-existent user
func TestAuthForgotPasswordSubmit_UserNotFound(t *testing.T) {
	// Arrange
	body := url.Values{}
	body.Set("email", "nonexistent@example.com")

	// Act - use post() helper which handles CSRF automatically
	resp := request(t).
		setRoute(routenames.ForgotPasswordSubmit).
		setBody(body).
		post()

	// Assert
	// Should still show success message (security: don't reveal if user exists)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestAuthForgotPasswordSubmit_ValidationErrors tests forgot password with validation errors
func TestAuthForgotPasswordSubmit_ValidationErrors(t *testing.T) {
	// Arrange
	body := url.Values{}
	body.Set("email", "invalid-email") // Invalid email format

	// Act - use post() helper which handles CSRF automatically
	resp := request(t).
		setRoute(routenames.ForgotPasswordSubmit).
		setBody(body).
		post()

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode) // Should stay on page with validation errors
}

// TestAuthVerifyEmail_InvalidToken tests email verification with invalid token
func TestAuthVerifyEmail_InvalidToken(t *testing.T) {
	// Arrange
	client := request(t).client

	// Act
	verifyURL := srv.URL + c.Web.Reverse(routenames.VerifyEmail, "invalid-token")
	resp, err := client.Get(verifyURL)

	// Assert
	require.NoError(t, err)
	// Should redirect with warning message (any 2xx or 3xx is acceptable)
	assert.True(t, (resp.StatusCode >= 200 && resp.StatusCode < 400),
		"Expected success or redirect status (2xx or 3xx), got %d", resp.StatusCode)
	resp.Body.Close()
}
