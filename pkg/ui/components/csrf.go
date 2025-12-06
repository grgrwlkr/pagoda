package components

import (
	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// CSRFInput creates a hidden input field for CSRF token.
// The token is taken from the request context, which is set by Echo CSRF middleware.
// This ensures the token in the form matches the _csrf cookie for the session.
func CSRFInput(r *ui.Request) Node {
	if r.CSRF == "" {
		return nil
	}
	return Input(
		Type("hidden"),
		Name("csrf"),
		Value(r.CSRF),
	)
}
