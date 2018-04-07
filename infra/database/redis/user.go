package redis

import (
	"log"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/nasa9084/ident/domain/entity"
)

var nilUser = entity.User{}

// Error is package-specific error type.
type Error string

// error constants
const (
	ErrUserNotFound Error = "user not found"
)

// Error implements error interface.
func (e Error) Error() string { return string(e) }

// ExistUser returns given user exists in Redis or not.
func ExistUser(conn redis.Conn, userID string) (bool, error) {
	const exist = 1 // redis::EXISTS returns 1 if the key exists

	resp, err := redis.Int(conn.Do("EXISTS", "user:"+userID))
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return false, err
	}
	return resp == exist, nil
}

// CreateSession creates a new session.
func CreateSession(conn redis.Conn, userID string) (string, error) {
	sessid := uuid.New().String()
	if _, err := conn.Do("SET", "session:"+sessid, userID, "EX", 60*10); err != nil {
		return "", err
	}
	return sessid, nil
}

// FindUser finds by given user id from Redis.
func FindUser(conn redis.Conn, userID string) (entity.User, error) {
	userMap, err := redis.StringMap(conn.Do("HGETALL", "user:"+userID))
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

// UpdateUser updates the user on Redis.
func UpdateUser(conn redis.Conn, u entity.User) error {
	_, err := conn.Do("HSET", "user:"+u.ID,
		"password", u.Password,
		"email", u.Email,
		"totp_verified", u.TOTPVerified,
	)
	return err
}

// DeleteUser deletes from Redis.
func DeleteUser(conn redis.Conn, u entity.User) error {
	_, err := conn.Do("DEL", "user:"+u.ID)
	return err
}
