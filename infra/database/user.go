package database

import (
	"context"
	"database/sql"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/nasa9084/ident/domain/entity"
	"github.com/nasa9084/ident/domain/repository"
	"github.com/nasa9084/ident/util"
)

var nilUser = entity.User{}

type userRepository struct {
	MySQL *sql.DB
	Redis redis.Conn
}

// NewUserRepository returns a new UserRepo instance.
func NewUserRepository(rdb *sql.DB, kvs redis.Conn) repository.UserRepository {
	return &userRepository{
		MySQL: rdb,
		Redis: kvs,
	}
}

// ExistsUser returns whether the user id has been used or not.
func (repo *userRepository) ExistsUser(ctx context.Context, userID string) (bool, error) {
	existsInRedis, err := repo.existsInRedis(userID)
	if err != nil {
		return false, err
	}
	existsInMySQL, err := repo.existsInMySQL(ctx, userID)
	if err != nil {
		return false, err
	}

	return existsInRedis || existsInMySQL, nil
}

// CreateUser creates a new user into Redis and returns the session id.
// The user is temporary user.
func (repo *userRepository) CreateUser(ctx context.Context, userID, password string) (string, error) {
	secret := util.SHA512Digest(uuid.New().String())
	userKey := "user:" + userID

	repo.Redis.Send("MULTI")
	repo.Redis.Send("HMSET", userKey,
		"password", util.Hash(password, userID),
		"totp_secret", secret,
	)
	repo.Redis.Send("EXPIRE", userKey, 60*10)
	if _, err := repo.Redis.Do("EXEC"); err != nil {
		return "", err
	}
	return repo.createSession(userID)
}

// FindUserBySessionID finds user using user id associated with given session id.
func (repo *userRepository) FindUserBySessionID(ctx context.Context, sessid string) (entity.User, error) {
	userID, err := repo.findSession(sessid)
	if err != nil {
		return nilUser, err
	}
	return repo.FindUserByID(ctx, userID)
}

func (repo *userRepository) findSession(sessid string) (string, error) {
	return redis.String(repo.Redis.Do("GET", "session:"+sessid))
}

// FindUserByID finds user using given user id.
func (repo *userRepository) FindUserByID(ctx context.Context, userID string) (entity.User, error) {
	var u entity.User
	var err error
	u, err = repo.findFromRedis(userID)
	if err != nil && err != ErrUserNotFound {
		return nilUser, err
	}
	if u != nilUser {
		return u, nil
	}
	return repo.findFromMySQL(ctx, userID)
}

// UpdateUser updates user information.
func (repo *userRepository) UpdateUser(ctx context.Context, u entity.User) error {
	inRedis, err := repo.existsInRedis(u.ID)
	if err != nil {
		return err
	}
	if inRedis {
		return repo.updateRedis(u)
	}
	inMySQL, err := repo.existsInMySQL(ctx, u.ID)
	if err != nil {
		return err
	}
	if inMySQL {
		return repo.updateMySQL(ctx, u)
	}
	return ErrUserNotFound
}

// Verify makes user non-temporary.
func (repo *userRepository) Verify(ctx context.Context, u entity.User) error {
	exists, err := repo.existsInRedis(u.ID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserNotFound
	}
	if err := repo.createInMySQL(ctx, u); err != nil {
		return err
	}
	return repo.deleteFromRedis(u)
}

// DeleteUser deletes user from Redis and MySQL.
func (repo *userRepository) DeleteUser(ctx context.Context, u entity.User) error {
	if err := repo.deleteFromRedis(u); err != nil {
		return err
	}
	return repo.deleteFromMySQL(ctx, u)
}

func (repo *userRepository) CreateSession(u entity.User) (string, error) {
	sessid := uuid.New().String()
	_, err := repo.Redis.Do("SET", "session:"+sessid, u.ID)
	if err != nil {
		return "", err
	}
	return sessid, nil
}
