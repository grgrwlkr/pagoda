package session

import (
	"errors"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/context"
	"github.com/mikestefanello/pagoda/pkg/log"
)

// ErrStoreNotFound indicates that the session store was not present in the context
var ErrStoreNotFound = errors.New("session store not found")

// Get returns a session
func Get(ctx echo.Context, name string) (*sessions.Session, error) {
	logger := log.Ctx(ctx)
	logger.Info("=== SESSION GET START ===", "session_name", name)

	s := ctx.Get(context.SessionKey)
	if s == nil {
		logger.Error("Session store not found in context")
		return nil, ErrStoreNotFound
	}
	logger.Info("Session store found in context")

	store := s.(sessions.Store)
	sess, err := store.Get(ctx.Request(), name)
	if err != nil {
		logger.Error("Failed to get session from store", "error", err)
		logger.Info("=== SESSION GET END ===")
		return nil, err
	}

	logger.Info("Session retrieved", "session_id", sess.ID, "session_is_new", sess.IsNew, "session_values_count", len(sess.Values))
	logger.Info("Session values", "values", sess.Values)
	logger.Info("=== SESSION GET END ===")
	return sess, nil
}

// Store sets the session storage in the context
func Store(ctx echo.Context, store sessions.Store) {
	ctx.Set(context.SessionKey, store)
}
