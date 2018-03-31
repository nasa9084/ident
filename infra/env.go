package infra

import (
	"crypto/ecdsa"
	"database/sql"

	"github.com/gomodule/redigo/redis"
	"github.com/nasa9084/ident/domain/repository"
	sendgrid "github.com/sendgrid/sendgrid-go"
)

// Environment holds RDB Connection, KVS Connection, and Private KEY.
type Environment struct {
	RDB        *sql.DB
	KVS        redis.Conn
	Mail       *sendgrid.Client
	PrivateKey *ecdsa.PrivateKey
}

// GetUserRepository generates UserRepository instance fron env itself.
func (env Environment) GetUserRepository() repository.UserRepository {
	return repository.UserRepository{
		RDB: env.RDB,
		KVS: env.KVS,
	}
}

// SendVerifyMail sends address verification mail using sendgrid.
func (env Environment) SendVerifyMail(to, sessid string) error {
	_, err := env.Mail.Send(NewVerificationMail(to, sessid))
	return err
}
