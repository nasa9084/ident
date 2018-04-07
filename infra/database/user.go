package database

import (
	"context"
	"database/sql"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/nasa9084/ident/domain/entity"
	"github.com/nasa9084/ident/domain/repository"
	"github.com/nasa9084/ident/generator"
	"github.com/nasa9084/ident/infra/database/mysql"
	"github.com/nasa9084/ident/infra/database/redis"
	"github.com/nasa9084/ident/util"
)

var nilUser = entity.User{}

type userRepository struct {
	MySQL *sql.DB
	Redis redigo.Conn
}

// NewUserRepository returns a new UserRepo instance.
func NewUserRepository(rdb *sql.DB, kvs redigo.Conn) repository.UserRepository {
	return &userRepository{
		MySQL: rdb,
		Redis: kvs,
	}
}

// ExistsUser returns whether the user id has been used or not.
func (repo *userRepository) ExistsUser(ctx context.Context, userID string) (bool, error) {
	existsInRedis, err := redis.ExistUser(repo.Redis, userID)
	if err != nil {
		return false, err
	}
	tx, err := repo.MySQL.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	existsInMySQL, err := mysql.ExistUser(ctx, tx, userID)
	if err != nil {
		return false, err
	}

	return existsInRedis || existsInMySQL, nil
}

// CreateUser creates a new user into Redis and returns the session id.
// The user is temporary user.
func (repo *userRepository) CreateUser(ctx context.Context, userID, password string) (string, error) {
	secret := generator.NewSecret()
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
	return redis.CreateSession(repo.Redis, userID)
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
	return redigo.String(repo.Redis.Do("GET", "session:"+sessid))
}

// FindUserByID finds user using given user id.
func (repo *userRepository) FindUserByID(ctx context.Context, userID string) (entity.User, error) {
	var u entity.User
	var err error
	u, err = redis.FindUser(repo.Redis, userID)
	if err != nil && err != redis.ErrUserNotFound {
		return nilUser, err
	}
	if u != nilUser {
		return u, nil
	}
	tx, err := repo.MySQL.BeginTx(ctx, nil)
	if err != nil {
		return u, nil
	}
	return mysql.FindUser(ctx, tx, userID)
}

// UpdateUser updates user information.
func (repo *userRepository) UpdateUser(ctx context.Context, u entity.User) error {
	inRedis, err := redis.ExistUser(repo.Redis, u.ID)
	if err != nil {
		return err
	}
	if inRedis {
		return redis.UpdateUser(repo.Redis, u)
	}
	tx, err := repo.MySQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	inMySQL, err := mysql.ExistUser(ctx, tx, u.ID)
	if err != nil {
		return err
	}
	if !inMySQL {
		return ErrUserNotFound
	}
	if err := mysql.UpdateUser(ctx, tx, u); err != nil {
		return err
	}

	return tx.Commit()
}

// Verify makes user non-temporary.
func (repo *userRepository) Verify(ctx context.Context, u entity.User) error {
	exists, err := redis.ExistUser(repo.Redis, u.ID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserNotFound
	}
	tx, err := repo.MySQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := mysql.CreateUser(ctx, tx, u); err != nil {
		return err
	}
	if err := redis.DeleteUser(repo.Redis, u); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteUser deletes user from Redis and MySQL.
func (repo *userRepository) DeleteUser(ctx context.Context, u entity.User) error {
	if err := redis.DeleteUser(repo.Redis, u); err != nil {
		return err
	}
	tx, err := repo.MySQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := mysql.DeleteUser(ctx, tx, u); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *userRepository) CreateSession(u entity.User) (string, error) {
	sessid := uuid.New().String()
	_, err := repo.Redis.Do("SET", "session:"+sessid, u.ID)
	if err != nil {
		return "", err
	}
	return sessid, nil
}
