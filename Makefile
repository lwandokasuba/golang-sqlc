DB_URL=postgresql://root:secret@localhost:5432/simplebank?sslmode=disable

network:
	docker network create bank-network

postgres:
	docker-compose up -d

createdb:
	docker exec -it golang-sqlc-postgres-1 createdb --username=root --owner=root simplebank

dropdb:
	docker exec -it golang-sqlc-postgres-1 dropdb simplebank

migrateup:
	goose -dir internal/db/migrations postgres "$(DB_URL)" up

migratedown:
	goose -dir internal/db/migrations postgres "$(DB_URL)" down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run cmd/api/main.go

.PHONY: network postgres createdb dropdb migrateup migratedown sqlc test server
