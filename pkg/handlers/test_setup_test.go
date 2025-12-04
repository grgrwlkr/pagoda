package handlers

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mikestefanello/pagoda/config"
	"github.com/mikestefanello/pagoda/pkg/services"
)

// TestMain sets up the test environment and runs all tests.
// It initializes the services container and starts a test HTTP server.
// All test files in this package will use these global variables.
func TestMain(m *testing.M) {
	// Set the environment to test
	config.SwitchEnvironment(config.EnvTest)

	// Start a new container
	c = services.NewContainer()

	// Start a test HTTP server
	if err := BuildRouter(c); err != nil {
		panic(err)
	}
	srv = httptest.NewServer(c.Web)

	// Run tests
	exitVal := m.Run()

	// Shutdown the container and test server
	if err := c.Shutdown(); err != nil {
		panic(err)
	}
	srv.Close()

	os.Exit(exitVal)
}
