package infra

import (
	"crypto/ecdsa"
	"database/sql"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/nasa9084/ident/domain/repository"
	"github.com/nasa9084/ident/domain/service"
	"github.com/nasa9084/ident/infra/database"
)

// Environment holds RDB Connection, KVS Connection, and Private KEY.
type Environment struct {
	RDB        *sql.DB
	KVS        redis.Conn
	MailFrom   string
	Mail       service.Mail
	PrivateKey *ecdsa.PrivateKey
}

// GetUserRepository generates UserRepository instance fron env itself.
func (env Environment) GetUserRepository() repository.UserRepo {
	return database.NewUserRepository(env.RDB, env.KVS)
}

// SendVerifyMail sends address verification mail using sendgrid.
func (env Environment) SendVerifyMail(from, to, sessid string) error {
	const body = `access below to verify your e-mail address.
http://localhost:8080/v1/user/email/%s
`
	return env.Mail.Send(from, to, fmt.Sprintf(body, sessid))
}
