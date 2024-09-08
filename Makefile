postgres:
	docker run --name postgres16.2 -p 5432:5432 -e POSTGRES_USER=backend_stuff -e POSTGRES_PASSWORD=robinrobin -d postgres:16.2-alpine

createdb:
	docker exec -it postgres16.2 createdb --username=backend_stuff --owner=backend_stuff simple_bank

dropdb:
	docker exec -it postgres16.2 dropdb simple_bank

migrateup:
	migrate -path internal/db/migration -database "postgresql://backend_stuff:robinrobin@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path internal/db/migration -database "postgresql://backend_stuff:robinrobin@localhost:5432/simple_bank?sslmode=disable" -verbose down

test:
	go test -v -cover ./...

server:
	go run ./cmd/simplebank/main.go

sqlc:
	sqlc generate

.PHONY: postgres createdb dropdb migrateup migratedown test sqlc server