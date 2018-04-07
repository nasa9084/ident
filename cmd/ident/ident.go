package main

import (
	"log"
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/nasa9084/ident"
)

type options struct {
	Addr           string `short:"a" long:"addr" env:"IDENT_ADDR" value-name:"ADDR" default:":8080"`
	PrivateKeyPath string `long:"private-key-path" env:"PRIVATE_KEY_PATH" value-name:"PRIVATE_KEY_PATH" default:"key/id_ecdsa"`
	ident.MySQLConfig
	ident.RedisConfig
	ident.MailConfig
}

func main() { os.Exit(exec()) }

func exec() int {
	var opts options
	if _, err := flags.Parse(&opts); err != nil {
		if fe, ok := err.(*flags.Error); ok && fe.Type == flags.ErrHelp {
			return 0
		}
		log.Print(err)
		return 1
	}
	s, err := ident.NewServer(
		opts.Addr,
		opts.PrivateKeyPath,
		ident.ServerConfig{
			opts.MySQLConfig,
			opts.RedisConfig,
			opts.MailConfig,
		},
	)
	if err != nil {
		log.Print(err)
		return 1
	}
	if err := s.Run(); err != nil {
		log.Print(err)
		return 1
	}
	return 0
}
