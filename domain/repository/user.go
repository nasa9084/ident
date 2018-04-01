package repository

import (
	"context"

	"github.com/nasa9084/ident/domain/entity"
)

// UserRepository is an interface of operations with user.
type UserRepository interface {
	ExistsUser(ctx context.Context, userID string) (exists bool, err error)
	CreateUser(ctx context.Context, userID, password string) (sessionID string, err error)
	FindUserBySessionID(ctx context.Context, sessionID string) (entity.User, error)
	FindUserByID(ctx context.Context, userID string) (entity.User, error)
	UpdateUser(context.Context, entity.User) error
	Verify(context.Context, entity.User) error
	CreateSession(entity.User) (sessionID string, err error)
}
