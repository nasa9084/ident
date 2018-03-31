package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/nasa9084/ident/domain/entity"
	"github.com/nasa9084/ident/util"
)

// UserRepository privides some operations related to user entity.
type UserRepository struct {
	RDB *sql.DB
	KVS redis.Conn
}

// IsUserExists returns whether given user ID has been used or not.
// When the user ID has been used, this function returns true, otherwise false.
// This function returns error when some error is occurred on connect to databases.
func (repo *UserRepository) IsUserExists(ctx context.Context, userid string) (bool, error) {
	exists, err := repo.isUserExistsInKVS(userid)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}
	exists, err = repo.isUserExistsInRDB(ctx, userid)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}
	return false, nil
}

func (repo *UserRepository) isUserExistsInKVS(userid string) (bool, error) {
	const exist = 1
	// redis::EXISTS returns 1 if the key exists, otherwise returns 0
	res, err := redis.Int(repo.KVS.Do("EXISTS", "user:"+userid))
	if err != nil {
		return false, err
	}
	if res == exist {
		return true, nil
	}
	return false, nil
}

func (repo *UserRepository) isUserExistsInRDB(ctx context.Context, userid string) (bool, error) {
	// if not exists on redis, search on mysql
	tx, err := repo.RDB.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	stmt, err := tx.Prepare(`SELECT EXISTS (SELECT 1 FROM users WHERE user_id = ?)`)
	if err != nil {
		return false, err
	}
	row := stmt.QueryRow(userid)
	var exists int
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	if exists == 1 {
		return true, nil
	}
	return false, nil
}

// CreateUser creates a new temporaly user into KVS.
// The user is temporaly user, must verify TOTP, must verify email.
func (repo *UserRepository) CreateUser(ctx context.Context, userid, password string) (string, error) {
	exists, err := repo.IsUserExists(ctx, userid)
	if err != nil {
		return "", err
	}
	if exists {
		return "", ErrUserExists
	}

	secret := util.SHA512Digest(uuid.New().String())

	repo.KVS.Send("MULTI")
	repo.KVS.Send("HMSET", "user:"+userid,
		"password", util.Hash(password, userid),
		"totp_secret", secret,
	)
	repo.KVS.Send("EXPIRE", "user:"+userid, 60*10)
	sessid := uuid.New().String()
	repo.KVS.Send("SET", "session:"+sessid, userid, "EX", 60*10)
	if _, err := repo.KVS.Do("EXEC"); err != nil {
		return "", err
	}
	return sessid, nil
}

// LookupUserBySessionID finds a user associated with given session ID.
// This function searches both of RDB and KVS.
func (repo *UserRepository) LookupUserBySessionID(ctx context.Context, sessid string) (entity.User, error) {
	uid, err := repo.findSession(sessid)
	if err != nil {
		return entity.User{}, err
	}
	udb, err := repo.LookupUserByUserID(ctx, uid)
	if err == nil {
		return udb, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return entity.User{}, err
	}
	u, err := repo.findUserFromKVS(uid)
	if err != nil {
		return entity.User{}, err
	}
	return u, nil
}

func (repo *UserRepository) findSession(sessid string) (string, error) {
	uid, err := redis.String(repo.KVS.Do("GET", "session:"+sessid))
	if err != nil {
		if err == redis.ErrNil {
			return "", errors.New("no sessiond id found")
		}
		return "", err
	}
	return uid, nil
}

func (repo *UserRepository) findUserFromKVS(uid string) (entity.User, error) {
	uar, err := redis.Strings(repo.KVS.Do("HGETALL", "user:"+uid))
	if err != nil {
		return entity.User{}, err
	}
	ump := map[string]string{}
	for i := 0; i < len(uar)/2; i++ {
		// uar is []string{KEY, VALUE, KEY, VALUE...}
		ump[uar[i*2]] = uar[i*2+1]
	}
	var u entity.User
	u.ID = uid
	u.Password = ump["password"]
	u.TOTPSecret = ump["totp_secret"]
	if sbool, ok := ump["totp_verified"]; ok {
		b, err := strconv.ParseBool(sbool)
		if err != nil {
			return entity.User{}, err
		}
		u.TOTPVerified = b
	}
	return u, nil
}

// LookupUserByUserID finds a user using given user ID.
// This function searches only RDB, not searches KVS.
func (repo *UserRepository) LookupUserByUserID(ctx context.Context, uid string) (entity.User, error) {
	tx, err := repo.RDB.BeginTx(ctx, nil)
	if err != nil {
		return entity.User{}, err
	}
	stmt, err := tx.Prepare(`SELECT user_id, password, totp_secret FROM users WHERE user_id = ?`)
	if err != nil {
		return entity.User{}, err
	}
	row := stmt.QueryRow(uid)
	var u entity.User
	if err := row.Scan(&u.ID, &u.Password, &u.TOTPSecret); err != nil {
		return entity.User{}, err
	}
	return u, nil
}

// VerifyTOTP updates the user has been verified TOTP.
func (repo *UserRepository) VerifyTOTP(ctx context.Context, u entity.User) error {
	u.TOTPVerified = true

	exists, err := repo.IsUserExists(ctx, u.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("user not found")
	}

	if _, err := repo.KVS.Do("HSET", "user:"+u.ID, "totp_verified", true); err != nil {
		return err
	}
	return nil
}

// UpdateEmail updates user email in KVS.
func (repo *UserRepository) UpdateEmail(ctx context.Context, u entity.User) error {
	exists, err := repo.IsUserExists(ctx, u.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("user not found")
	}
	if _, err := repo.KVS.Do("HSET", "user:"+u.ID, "email", u.Email); err != nil {
		return err
	}
	return nil
}

// VerifyEmail move temporaly user into RDB.
func (repo *UserRepository) VerifyEmail(ctx context.Context, u entity.User) error {
	exists, err := repo.IsUserExists(ctx, u.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("user not found")
	}
	if _, err := repo.KVS.Do("DEL", "user:"+u.ID); err != nil {
		return err
	}
	tx, err := repo.RDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO users(user_id, password, totp_secret, email) VALUES(?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(u.ID, u.Password, u.TOTPSecret, u.Email); err != nil {
		return err
	}
	return tx.Commit()
}

// CreateSession creates a new web session and returns its ID.
// The session expires in 10 minutes.
func (repo *UserRepository) CreateSession(ctx context.Context, u entity.User) (string, error) {
	sessid := uuid.New().String()
	if _, err := repo.KVS.Do("SET", "session:"+sessid, u.ID, "EX", 60*10); err != nil {
		return "", err
	}
	return sessid, nil
}

// RenewSession recreates a web session and retursn its ID.
// The session expires in 10 minutes.
func (repo *UserRepository) RenewSession(ctx context.Context, u entity.User, oldSessid string) (string, error) {
	newSessid := uuid.New().String()
	repo.KVS.Send("MULTI")
	repo.KVS.Send("DEL", "session:"+oldSessid)
	repo.KVS.Send("SET", "session:"+newSessid, u.ID, "EX", 60*10)
	repo.KVS.Send("EXPIRE", "user:"+u.ID, 60*10)
	if _, err := repo.KVS.Do("EXEC"); err != nil {
		return "", err
	}
	return newSessid, nil
}
