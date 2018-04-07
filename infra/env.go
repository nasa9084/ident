package infra

import (
	"crypto/ecdsa"
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/nasa9084/ident/domain/repository"
	"github.com/nasa9084/ident/domain/service"
	"github.com/nasa9084/ident/infra/database"
	"github.com/nasa9084/ident/infra/mail"
)

// Config is wrapper for all configurations.
type Config struct {
	MySQL MySQLConfig
	Redis RedisConfig
	Mail  MailConfig
}

// MySQLConfig holds configurations for connect to MySQL server.
// This struct can also be used for go-flags.
type MySQLConfig struct {
	Addr     string `long:"mysql-addr" env:"MYSQL_ADDR" value-name:"MYSQL_ADDR" default:"127.0.0.1:3306"`
	User     string `long:"mysql-user" env:"MYSQL_USER" value-name:"MYSQL_USER" default:"root"`
	Password string `long:"mysql-password" env:"MYSQL_PASSWORD" value-name:"MYSQL_PASSWORD" default:""`
	DBName   string `long:"mysql-db" env:"MYSQL_DB" value-name:"MYSQL_DB" default:"ident"`
}

// RedisConfig holds configurations for connect to Redis server.
// This struct can also be used for go-flags.
type RedisConfig struct {
	Addr string `long:"redis-addr" env:"REDIS_ADDR" value-name:"REDIS_ADDR" default:"127.0.0.1:6379"`
}

// MailConfig holds configuration to use to sendgrid API.
// This struct can also be used for go-flags.
type MailConfig struct {
	FromAddr string `long:"email" env:"EMAIL_ADDR"`
	APIKey   string `long:"sg-apikey" env:"SENDGRID_APIKEY" value-name:"SENDGRID_APIKEY" required:"yes"`
}

// Environment holds RDB Connection, KVS Connection, and Private KEY.
type Environment struct {
	RDB        *sql.DB
	KVS        redis.Conn
	MailFrom   string
	Mail       service.Mail
	PrivateKey *ecdsa.PrivateKey
}

// NewEnvironment returns a new Environment object.
func NewEnvironment(cfg Config, keyPath string) (*Environment, error) {
	rdb, err := openMySQL(cfg.MySQL)
	if err != nil {
		return nil, err
	}
	kvs, err := redis.Dial("tcp", cfg.Redis.Addr)
	if err != nil {
		return nil, err
	}
	key, err := LoadPrivateKey(keyPath)
	if err != nil {
		return nil, err
	}
	env := &Environment{
		RDB:        rdb,
		KVS:        kvs,
		Mail:       mail.NewSendGrid(cfg.Mail.APIKey, cfg.Mail.FromAddr),
		PrivateKey: key,
	}
	return env, nil
}

func openMySQL(opts MySQLConfig) (*sql.DB, error) {
	cfg := mysql.Config{
		Net:    "tcp",
		Addr:   opts.Addr,
		User:   opts.User,
		Passwd: opts.Password,
		DBName: opts.DBName,

		ParseTime: true,
	}
	return sql.Open("mysql", cfg.FormatDSN())
}

// GetUserRepository generates UserRepository instance fron env itself.
func (env Environment) GetUserRepository() repository.UserRepository {
	return database.NewUserRepository(env.RDB, env.KVS)
}

// SendVerifyMail sends address verification mail using sendgrid.
func (env Environment) SendVerifyMail(from, to, sessid string) error {
	const body = `access below to verify your e-mail address.
http://localhost:8080/v1/user/email/%s
`
	return env.Mail.Send(from, to, fmt.Sprintf(body, sessid))
}
