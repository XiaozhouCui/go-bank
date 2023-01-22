postgres:
	docker compose start

createdb:
	docker exec -it go-bank-db-1 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it go-bank-db-1 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:54321/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:54321/simple_bank?sslmode=disable" -verbose down

.PHONY: postgres createdb dropdb migrateup migratedown
