package database

import (
	"log"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/nasa9084/ident/domain/entity"
)

func (repo *userRepository) existsInRedis(userID string) (bool, error) {
	const exist = 1 // redis::EXISTS returns 1 if the key exists

	resp, err := redis.Int(repo.Redis.Do("EXISTS", "user:"+userID))
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return false, err
	}
	return resp == exist, nil
}

func (repo *userRepository) createSession(userID string) (string, error) {
	sessid := uuid.New().String()
	if _, err := repo.Redis.Do("SET", "session:"+sessid, userID, "EX", 60*10); err != nil {
		return "", err
	}
	return sessid, nil
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

func (repo *userRepository) updateRedis(u entity.User) error {
	_, err := repo.Redis.Do("HSET", "user:"+u.ID,
		"password", u.Password,
		"email", u.Email,
		"totp_verified", u.TOTPVerified,
	)
	return err
}

func (repo *userRepository) deleteFromRedis(u entity.User) error {
	_, err := repo.Redis.Do("DEL", "user:"+u.ID)
	return err
}
