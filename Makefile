keygen:
	@rm -fr key
	@mkdir key
	@go run internal/cmd/keygen/keygen.go

generate:
	@echo "generate handlers..."

initdb:
	@echo "drop database"
	@mysql -uroot -e 'DROP DATABASE IF EXISTS ident;'
	@echo "create database"
	@mysql -uroot -e 'CREATE DATABASE ident;'
	@mysql -uroot ident < sql/ident.sql
