package handlers

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// httpRequest represents an HTTP request builder for tests.
// It provides a fluent API for constructing and executing HTTP requests.
type httpRequest struct {
	route  string
	client http.Client
	body   url.Values
	t      *testing.T
}

// request creates a new HTTP request builder with a fresh cookie jar.
// This is the entry point for making HTTP requests in tests.
func request(t *testing.T) *httpRequest {
	jar, err := cookiejar.New(nil)
	require.NoError(t, err, "Failed to create cookie jar")

	return &httpRequest{
		t:    t,
		body: url.Values{},
		client: http.Client{
			Jar: jar,
		},
	}
}

// setClient sets a custom HTTP client to use for the request.
// This is useful when you need to use an authenticated client.
func (h *httpRequest) setClient(client http.Client) *httpRequest {
	h.client = client
	return h
}

// setRoute sets the route for the request using a route name and optional parameters.
// The route is automatically prefixed with the test server URL.
func (h *httpRequest) setRoute(route string, params ...any) *httpRequest {
	h.route = srv.URL + c.Web.Reverse(route, params)
	return h
}

// setBody sets the form body for POST/PUT requests.
func (h *httpRequest) setBody(body url.Values) *httpRequest {
	h.body = body
	return h
}

// get executes a GET request and returns the response.
func (h *httpRequest) get() *httpResponse {
	resp, err := h.client.Get(h.route)
	require.NoError(h.t, err, "Failed to execute GET request")
	return &httpResponse{
		t:        h.t,
		Response: resp,
	}
}

// post executes a POST request with automatic CSRF token handling.
// It first makes a GET request to the same route to retrieve the CSRF token,
// then includes it in the POST request body.
func (h *httpRequest) post() *httpResponse {
	// Make a GET request to retrieve the CSRF token
	doc := h.get().
		assertStatusCode(http.StatusOK).
		toDoc()

	// Extract the CSRF token and include it in the POST request body
	csrf := doc.Find(`input[name="csrf"]`).First()
	token, exists := csrf.Attr("value")
	assert.True(h.t, exists, "CSRF token should exist in form")
	h.body["csrf"] = []string{token}

	// Execute the POST request
	resp, err := h.client.PostForm(h.route, h.body)
	require.NoError(h.t, err, "Failed to execute POST request")
	return &httpResponse{
		t:        h.t,
		Response: resp,
	}
}

// postAuthenticated executes a POST request using an authenticated HTTP client.
// It first makes a GET request using the provided client to retrieve the CSRF token,
// then includes it in the POST request body.
// This is useful when you need to make authenticated POST requests.
func (h *httpRequest) postAuthenticated(client *http.Client) *httpResponse {
	// Make a GET request to retrieve the CSRF token using authenticated client
	resp, err := client.Get(h.route)
	require.NoError(h.t, err, "Failed to get CSRF token page")
	require.Equal(h.t, http.StatusOK, resp.StatusCode, "CSRF token page should return 200")

	// Extract the CSRF token
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	require.NoError(h.t, err, "Failed to parse CSRF token page")
	resp.Body.Close()

	csrf := doc.Find(`input[name="csrf"]`).First()
	token, exists := csrf.Attr("value")
	require.True(h.t, exists, "CSRF token should exist in form")
	h.body["csrf"] = []string{token}

	// Execute the POST request using authenticated client
	resp2, err := client.PostForm(h.route, h.body)
	require.NoError(h.t, err, "Failed to execute authenticated POST request")
	return &httpResponse{
		t:        h.t,
		Response: resp2,
	}
}

// httpResponse represents an HTTP response wrapper for tests.
// It provides convenient methods for asserting response properties.
type httpResponse struct {
	*http.Response
	t *testing.T
}

// assertStatusCode asserts that the response has the expected status code.
// Returns itself for method chaining.
func (h *httpResponse) assertStatusCode(code int) *httpResponse {
	assert.Equal(h.t, code, h.Response.StatusCode,
		"Expected status code %d, got %d", code, h.Response.StatusCode)
	return h
}

// assertRedirect asserts that the response is a redirect to the expected route.
// Returns itself for method chaining.
func (h *httpResponse) assertRedirect(t *testing.T, route string, params ...any) *httpResponse {
	expectedURL := c.Web.Reverse(route, params)
	actualURL := h.Header.Get("Location")
	assert.Equal(t, expectedURL, actualURL,
		"Expected redirect to %s, got %s", expectedURL, actualURL)
	return h
}

// toDoc parses the response body as HTML and returns a goquery document.
// The response body is automatically closed after parsing.
func (h *httpResponse) toDoc() *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(h.Body)
	require.NoError(h.t, err, "Failed to parse response body as HTML")
	err = h.Body.Close()
	assert.NoError(h.t, err, "Failed to close response body")
	return doc
}
