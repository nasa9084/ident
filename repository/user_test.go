package repository_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/nasa9084/ident/repository"
)

var (
	userRDB *sql.DB
	userKVS redis.Conn
)

const (
	aliceID = "alice"
)

func TestUserRepository(t *testing.T) {
	var err error
	userRDB, err = sql.Open("mysql", "root@tcp(127.0.0.1:3306)/ident")
	if err != nil {
		t.Fatal(err)
	}
	userKVS, err = redis.Dial("tcp", "localhost:6379")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("is user exists", testIsUserExists)
}

func testIsUserExists(t *testing.T) {
	repo := &repository.UserRepository{
		RDB: userRDB,
		KVS: userKVS,
	}
	exists, err := repo.IsUserExists(context.Background(), aliceID)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(exists)
}
