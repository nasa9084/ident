keygen:
	@rm -fr key
	@mkdir key
	@go run internal/cmd/keygen/keygen.go

initdb:
	@echo "flush redis"
	@redis-cli $(REDIS_OPTS) flushdb
	@echo "drop database"
	@mysql $(MYSQL_OPTS) -uroot -e 'DROP DATABASE IF EXISTS ident;'
	@echo "create database"
	@mysql $(MYSQL_OPTS) -uroot -e 'CREATE DATABASE ident;'
	@mysql $(MYSQL_OPTS) -uroot ident < sql/ident.sql

generate:
	@go run internal/cmd/genHandler/genHandler.go -f spec/ident.yml

test:
	@TEST_KEYPATH=$(PWD)/key/id_ecdsa go test -v ./...
