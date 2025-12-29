# Makefile for simple_bank migrations (Windows-friendly)

# Variables
DB_CONTAINER=postgres12
DB_NAME=simple_bank
DB_USER=root
DB_PASSWORD=secret
DB_PORT=5433
MIGRATE_PATH=db/migration
# Use localhost because Docker for Windows maps ports to host
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable
# Update this path to your Windows user directory
POSTGRES_DATA=/c/Users/itsco/postgres_data

# Phony targets
.PHONY: postgres start-postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc test server mock

# Start Postgres container (fresh)
postgres:
	docker run --name $(DB_CONTAINER) -p $(DB_PORT):5432 \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-v $(POSTGRES_DATA):/var/lib/postgresql/data \
		-d postgres:12-alpine

# Start existing Postgres container
start-postgres:
	docker start -ai $(DB_CONTAINER)

# Create database (if container already running)
createdb:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -c "CREATE DATABASE $(DB_NAME);"

# Drop database
dropdb:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -c "DROP DATABASE IF EXISTS $(DB_NAME);"

# Run migrations up
migrateup:
	migrate -path $(MIGRATE_PATH) -database "$(DB_URL)" -verbose up

# Run migrations down (rollback last migration)
migratedown:
	migrate -path $(MIGRATE_PATH) -database "$(DB_URL)" -verbose down

# Run migration up (one-up only)
migrateup1:
	migrate -path $(MIGRATE_PATH) -database "$(DB_URL)" -verbose up 1

# Run migration down (one-down only)
migratedown1:
	migrate -path $(MIGRATE_PATH) -database "$(DB_URL)" -verbose down 1

# Run sqlc
sqlc:
	sqlc generate

# Run tests
test:
	go test -v -cover ./...

# Run server
server:
	go run main.go

# Run mock
mock:
	mockgen -destination=db/mock/store.go -package=mock github.com/codercollo/simple_bank/db/sqlc Store