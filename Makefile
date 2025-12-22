# Makefile for simple_bank migrations

# Variables
DB_CONTAINER=postgres12
DB_NAME=simple_bank
DB_USER=root
DB_PASSWORD=secret
DB_PORT=5433
MIGRATE_PATH=db/migration
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Phony targets
.PHONY: postgres createdb dropdb migrateup migratedown sqlc test

# Start Postgres container
postgres:
	sudo docker run --name $(DB_CONTAINER) -p $(DB_PORT):5432 \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-v ~/postgres_data:/var/lib/postgresql/data \
		-d postgres:12-alpine

# Create database (if container already running)
createdb:
	sudo docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -c "CREATE DATABASE $(DB_NAME);"

# Drop database
dropdb:
	sudo docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -c "DROP DATABASE IF EXISTS $(DB_NAME);"

# Run migrations up
migrateup:
	migrate -path $(MIGRATE_PATH) -database "$(DB_URL)" -verbose up

# Run migrations down (rollback last migration)
migratedown:
	migrate -path $(MIGRATE_PATH) -database "$(DB_URL)" -verbose down

#Run sqlc
sqlc:
	~/go/bin/sqlc generate

#Run test
test:
	go test -v -cover ./...