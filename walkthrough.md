source code: https://github.com/techschool/simplebank

### 1. Database setup

- Generated ERD: https://dbdiagram.io/d/63cb8d49296d97641d7b24f9, generate and export postgres queries
- Run docker for postgres: `docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest`
- Instead, use docker compose: `docker compose up`, db container name should be _postgres-1_
- To connect to the db in docker: `docker exec -it postgres psql -U root`
- Inside db, run `select now();` to show date time, run `\q` to quit pg container
- To show logs: `docker logs postgres`
- Use TablePlus to open database as user `root`
- Run the exported sql query from _dbdiagram_ to create tables

### 2. Database migration

- Install golang-migrate: `brew install golang-migrate`
- Validate installation: `migrate -version`
- Add folder: `mkdir -p db/migration`
- Create first migration "init_schema": `migrate create -ext sql -dir db/migration -seq init_schema`
- Copy the exported SQL queries into the generated up file _blah.up.sql_
- In the down file _blah.down.sql_, add sql queries to drop tables

#### 2.1 Create db from inside the shell

- To access shell: `docker exec -it postgres /bin/sh`
- Create new db "simple_bank": `createdb --username=root --owner=root simple_bank`
- Connect to db "simple_bank“: `psql simple_bank`, then run `\l` to list databases
- Drop the db "simple_bank": run `\q` to disconnect db then run `drop simple_bank`, then run `exit` to quit shell

#### 2.2 Create db from terminal

- To create db "simple_bank" from terminal: `docker exec -it postgres createdb --username=root --owner=root simple_bank`
- To connect to db "simple_bank" from terminal: `docker exec -it postgres psql -U root simple_bank`
- Disconnect db and return to terminal: `\q`
- To drop db "simple_bank" from terminal: `docker exec -it postgres dropdb simple_bank`
- The above commands are added in the _Makefile_, run `make createdb` or `make dropdb`

#### 2.3 Run migration using Makefile

- To run migration up: `migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up`
- To run migration down: `migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down`
- Add the commands into Makefile, then run `make migrateup` or `make migratedown`

### 3 Generate CRUD Golang code from SQLC

- Install sqlc: `brew install sqlc`, validate installation by running `sqlc version`
- In project root folder, run `sqlc init`, and update the generated _sqlc.yaml_, created corresponding folders inside _./db_
- Update Makefile to run `sqlc generate`

#### 3.1 Generate CreateAccount method of Query object

- Create _db/query/account.sql_, add a `INSERT` query to one create account
- Run `make sqlc`, go files will be generated in _db/sqlc_
- SQLC will read schema from migration sql files, and generate models in go
- Do NOT modify the generated go files, because they will be regenerated everytime we run `make sqlc`

#### 3.2 Fix missing dependencies in the generated files

- Initialize go mod: `go mod init github.com/XiaozhouCui/go-bank`
- Run `go mod tidy` to automatically fix the missing dependencies

#### 3.3 Generate GetAccount and ListAccounts methods

- Add 2 `SELECT` queries in _db/query/account.sql_
- Run `make sqlc` will only update _account.sql.go_ to generate `GetAccount` and `ListAccounts` methods

#### 3.4 Generate UpdateAccount method

- Add a `UPDATE` query in _db/query/account.sql_, to only update the account `balance`
- Run `make sqlc` will update _account.sql.go_ to generate `UpdateAccount` method

#### 3.5 Generate DeleteAccount method

- Add a `DELETE` query in _db/query/account.sql_, to delete an account
- Run `make sqlc` will update _account.sql.go_ to generate `DeleteAccount` method

### 4 Add unit tests for database CRUD

#### 4.1 Setup unit test entry point and dependencies

- Create _db/sqlc/main_test.go_ as the main entry point
- Create test file db/sqlc/account_test.go
- Install lib/pq: `go get github.com/lib/pq`, this will update _go.mod_
- Import `github.com/lib/pq` in the main_test.go, so that unit tests can connect to database in docker
- Run `go mod tidy` to clean up _go.mod_
- Run `go test -v -cover ./...` should pass all tests
- Install testify: `go get github.com/stretchr/testify`

### 5 Load config from file and environment variables

- Install viper: `go get github.com/spf13/viper`
- Create _app.env_ in root folder to config environment variables
- Create _db/util/config.go_ to read configuration from file or environment vairables
- Update _main.go_ and _main_test.go_ to load db configs with viper

### 6 Use gomock to mock db for testing

#### 6.1 Setup Store interface for the mock

- Install mockgen: `go install github.com/golang/mock/mockgen@v1.6.0`
- Update store.go to convert `Store` from struct to an interface
- Update sqlc.yaml `emit_interface: true`
- Run `make sqlc` to generate _go/sqlc/querier.go_
- Embed the `Querier` into the `Store` interface in store.go

#### 6.2 Generate mock file

- Create folder _db/mock_
- Run `mockgen -build_flags=--mod=mod -package mockdb -destination db/mock/store.go github.com/XiaozhouCui/go-bank/db/sqlc Store`
- This will generate _db/mock/store.go_
- Add the command to makefile

### 7 Create users table for auth

#### 7.1 Add migration for users table

- Create migration: `migrate create -ext sql -dir db/migration -seq add_users`
- Update the generated sql migration files
- Add one-step migrations in makefile, and ran the migrations

#### 7.2 Add sql query for users table

- Create slq file _db/query/user.sql_
- Run `make sqlc` to generate functions `CreateUser` and `GetUser` in _db/sqlc/user.sql.go_
- Run `make mock` to re-run mockgen, which will add the above functions into _db/mock/store.go_
- Add tests for the above functions `user_test.go` in _db/sqlc_
- Need to update `account_test.go` for foreign key constraintsk, adding `user.Username` as `account.Owner`
- Handle foreign key errors in CreateAccount by using pq error code.

#### 7.3 Hash the password

- Use `bcrypt` to hash the password
- In _db/util/password.go_, add `HashPassword` and `CheckPassword`

### 8 Deployment

#### 8.1 Upgrade golang-migrate

- Create a new branch `git checkout -b ft/docker`
- Run `brew upgrade golang-migrate`
- Verify the upgrade `migrate -version`, and run `make migrateup`
- Update the version in ci.yml
