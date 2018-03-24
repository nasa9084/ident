package infra

import (
	"database/sql"

	"github.com/go-sql-driver/mysql"
)

// OpenMySQL connection.
func OpenMySQL(addr, user, password, dbname string) (*sql.DB, error) {
	cfg := mysql.Config{
		Net:    "tcp",
		Addr:   addr,
		User:   user,
		Passwd: password,
		DBName: dbname,

		ParseTime: true,
	}
	return sql.Open("mysql", cfg.FormatDSN())
}
