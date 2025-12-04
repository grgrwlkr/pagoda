package handlers

import (
	"net/http/httptest"

	"github.com/mikestefanello/pagoda/pkg/services"
)

var (
	// srv is the test HTTP server instance
	srv *httptest.Server
	// c is the services container used in tests
	c *services.Container
)
