package infra

import "github.com/gomodule/redigo/redis"

// OpenRedis connection.
func OpenRedis(addr string) (redis.Conn, error) {
	return redis.Dial("tcp", addr)
}
