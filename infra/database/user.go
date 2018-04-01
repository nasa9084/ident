package database

import (
	"context"
	"database/sql"
	"log"
	"strconv"

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

func (repo *userRepository) existsInRedis(userID string) (bool, error) {
	const exist = 1 // redis::EXISTS returns 1 if the key exists

	resp, err := redis.Int(repo.Redis.Do("EXISTS", "user:"+userID))
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return false, err
	}
	return resp == exist, nil
}

func (repo *userRepository) existsInMySQL(ctx context.Context, userID string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM users WHERE user_id = ?)`
	const exist = 1

	row := repo.MySQL.QueryRowContext(ctx, query, userID)
	var resp int
	if err := row.Scan(&resp); err != nil {
		return false, err
	}
	return resp == exist, nil
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

func (repo *userRepository) createSession(userID string) (string, error) {
	sessid := uuid.New().String()
	if _, err := repo.Redis.Do("SET", "session:"+sessid, userID, "EX", 60*10); err != nil {
		return "", err
	}
	return sessid, nil
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

func (repo *userRepository) findFromRedis(userID string) (entity.User, error) {
	userMap, err := redis.StringMap(repo.Redis.Do("HGETALL", "user:"+userID))
	if err != nil {
		return nilUser, err
	}
	if len(userMap) == 0 {
		return nilUser, ErrUserNotFound
	}

	u := entity.User{
		ID:         userID,
		Password:   userMap["password"],
		TOTPSecret: userMap["totp_secret"],
	}

	if b, ok := userMap["totp_verified"]; ok {
		totpVerified, err := strconv.ParseBool(b)
		if err != nil {
			return nilUser, err
		}
		u.TOTPVerified = totpVerified
	}

	return u, nil
}

func (repo *userRepository) findFromMySQL(ctx context.Context, userID string) (entity.User, error) {
	const query = `SELECT user_id, password, totp_secret, email FROM users WHERE user_id = ?`
	row := repo.MySQL.QueryRowContext(ctx, query, userID)
	var u entity.User
	if err := row.Scan(&u.ID, &u.Password, &u.TOTPSecret, &u.Email); err != nil {
		return nilUser, err
	}
	u.TOTPVerified = true
	return u, nil
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

func (repo *userRepository) updateRedis(u entity.User) error {
	_, err := repo.Redis.Do("HSET", "user:"+u.ID,
		"password", u.Password,
		"email", u.Email,
		"totp_verified", u.TOTPVerified,
	)
	return err
}

func (repo *userRepository) updateMySQL(ctx context.Context, u entity.User) error {
	const query = `UPDATE users SET password=?, email=? WHERE user_id=?`
	tx, err := repo.MySQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(u.Password, u.Email, u.ID); err != nil {
		return err
	}
	return tx.Commit()
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

func (repo *userRepository) createInMySQL(ctx context.Context, u entity.User) error {
	const query = `INSERT INTO users(user_id, password, totp_secret, email) VALUES(?, ?, ?, ?)`
	tx, err := repo.MySQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(u.ID, u.Password, u.TOTPSecret, u.Email); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *userRepository) deleteFromRedis(u entity.User) error {
	_, err := repo.Redis.Do("DEL", "user:"+u.ID)
	return err
}

// DeleteUser deletes user from Redis and MySQL.
func (repo *userRepository) DeleteUser(ctx context.Context, u entity.User) error {
	if err := repo.deleteFromRedis(u); err != nil {
		return err
	}
	return repo.deleteFromMySQL(ctx, u)
}

func (repo *userRepository) deleteFromMySQL(ctx context.Context, u entity.User) error {
	const query = `DELETE FROM users WHERE user_id = ?`
	tx, err := repo.MySQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(u.ID); err != nil {
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
