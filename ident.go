package ident

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/nasa9084/ident/infra"
	"github.com/nasa9084/ident/infra/mail"
	"github.com/nasa9084/syg"
)

// Server is a main application object.
type Server struct {
	server *http.Server
	closed chan struct{}
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

// NewServer returns a new server.
func NewServer(addr string, privKeyPath string, mysqlCfg MySQLConfig, redisCfg RedisConfig, mailCfg MailConfig) (*Server, error) {
	rdb, err := openMySQL(mysqlCfg)
	if err != nil {
		return nil, err
	}
	kvs, err := redis.Dial("tcp", redisCfg.Addr)
	if err != nil {
		return nil, err
	}
	mailer := mail.NewSendGrid(mailCfg.APIKey, mailCfg.FromAddr)
	key, err := infra.LoadPrivateKey(privKeyPath)
	if err != nil {
		return nil, err
	}
	env := &infra.Environment{
		RDB:        rdb,
		KVS:        kvs,
		Mail:       mailer,
		PrivateKey: key,
	}
	router := mux.NewRouter()
	bindRoutes(router, env)

	s := &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
		closed: make(chan struct{}),
	}
	return s, nil
}

// Run the server.
func (s *Server) Run() error {
	cancel := syg.Listen(s.Shutdown, os.Interrupt)
	defer cancel()

	log.Printf("server is listening on: %s", s.server.Addr)
	err := s.server.ListenAndServe()
	<-s.closed
	return err
}

// Shutdown shuts down the server gracefully.
func (s *Server) Shutdown(os.Signal) {
	defer close(s.closed)
	log.Print("server shutdown")

	s.server.Shutdown(context.Background())
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
