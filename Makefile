DB_URL=postgresql://backend_stuff:robinrobin@localhost:5432/simple_bank?sslmode=disable

postgres:
	docker run --name postgres16.2 --network bank-network -p 5432:5432 -e POSTGRES_USER=backend_stuff -e POSTGRES_PASSWORD=robinrobin -d postgres:16.2-alpine

createdb:
	docker exec -it postgres16.2 createdb --username=backend_stuff --owner=backend_stuff simple_bank

dropdb:
	docker exec -it postgres16.2 dropdb simple_bank

migrateup:
	migrate -path internal/db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path internal/db/migration -database "$(DB_URL)" -verbose down

test:
	go test -v -cover -short ./...

server:  
	go run ./cmd/simplebank/main.go

new_migration:
	migrate create -ext sql -dir internal/db/migration -seq $(name)

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

sqlc:
	sqlc generate

proto:
	rm -f internal/pb/*.go
	rm -rf doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=internal/pb --go_opt=paths=source_relative \
	--go-grpc_out=internal/pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=internal/pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
	proto/*.proto 
	statik -src=./doc/swagger -dest=./doc

evans:
	evans --host localhost --port 9090 -r repl --package pb --service SimpleBank

redis:
	docker run --name redis7.4.1 -p 6379:6379 -d redis:7.4.1-alpine

.PHONY: postgres createdb dropdb migrateup migratedown test sqlc server db_docs db_schema proto evans redis new_migration