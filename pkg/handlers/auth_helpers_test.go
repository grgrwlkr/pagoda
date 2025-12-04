package handlers

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	"github.com/mikestefanello/pagoda/ent"
	"github.com/mikestefanello/pagoda/pkg/routenames"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestUser creates a test user in the database and returns it.
// The user is created with the provided email, name, and password.
// It's the caller's responsibility to clean up the user after the test.
func createTestUser(t *testing.T, email, name, password string) *ent.User {
	ctx := context.Background()

	// Create user
	usr, err := c.ORM.User.Create().
		SetEmail(email).
		SetName(name).
		SetPassword(password).
		Save(ctx)

	require.NoError(t, err, "Failed to create test user")
	return usr
}

// authenticateUser authenticates a test user and returns an authenticated HTTP client.
// The client will have session cookies set, allowing it to make authenticated requests.
// This helper automatically handles CSRF token retrieval and login form submission.
func authenticateUser(t *testing.T, email, password string) *http.Client {
	jar, err := cookiejar.New(nil)
	require.NoError(t, err, "Failed to create cookie jar")

	client := &http.Client{
		Jar: jar,
	}

	// Get login page to retrieve CSRF token
	loginURL := srv.URL + c.Web.Reverse(routenames.Login)
	resp, err := client.Get(loginURL)
	require.NoError(t, err, "Failed to get login page")
	require.Equal(t, http.StatusOK, resp.StatusCode, "Login page should return 200")

	// Extract CSRF token from login form
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	require.NoError(t, err, "Failed to parse login page")
	resp.Body.Close()

	csrf := doc.Find(`input[name="csrf"]`).First()
	token, exists := csrf.Attr("value")
	require.True(t, exists, "CSRF token should exist in login form")

	// Prepare login request
	loginBody := url.Values{}
	loginBody.Set("email", email)
	loginBody.Set("password", password)
	loginBody.Set("csrf", token)

	// Submit login form
	resp, err = client.PostForm(loginURL, loginBody)
	require.NoError(t, err, "Failed to submit login form")
	// Accept both 302 (Found) and 307 (Temporary Redirect) as success
	assert.True(t, resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusTemporaryRedirect,
		"Login should redirect (302 or 307), got %d", resp.StatusCode)
	resp.Body.Close()

	return client
}
