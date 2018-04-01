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
	@$(eval DBNAME := $(shell python -c "import uuid; print(str(uuid.uuid4()).replace('-', ''))"))
	@mysql $(MYSQL_OPTS) -uroot -e "CREATE DATABASE $(DBNAME);"
	@mysql $(MYSQL_OPTS) -uroot $(DBNAME) < sql/ident.sql
	@-TEST_KEYPATH=$(PWD)/key/id_ecdsa MYSQL_DB=$(DBNAME) go test -v ./...
	@mysql $(MYSQL_OPTS) -uroot -e "DROP DATABASE $(DBNAME);"

check:
	@echo "go vet"
	@go vet -v ./...
	@echo "golint"
	@go list ./... | xargs -L1 golint
	@echo "staticcheck"
	@staticcheck ./...
	@echo "gosimple"
	@gosimple ./...
