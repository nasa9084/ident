keygen:
	@rm -fr key
	@mkdir key
	@go run internal/cmd/keygen/keygen.go

initdb:
	@echo "flush redis"
	@redis-cli flushdb
	@echo "drop database"
	@mysql -uroot -e 'DROP DATABASE IF EXISTS ident;'
	@echo "create database"
	@mysql -uroot -e 'CREATE DATABASE ident;'
	@mysql -uroot ident < sql/ident.sql

generate:
	@go run internal/cmd/genHandler/genHandler.go -f spec/ident.yml

test:
	@TEST_KEYPATH=$(PWD)/key/id_ecdsa go test -v ./...
