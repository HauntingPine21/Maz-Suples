.PHONY: run test vet build fmt db-init db-seed bootstrap-admin

run:
	go run ./cmd/server

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -o bin/maz-suplementos ./cmd/server

fmt:
	gofmt -w cmd internal

db-init:
	mysql --ssl-mode=VERIFY_IDENTITY -h "$(DB_HOST)" -P "$(DB_PORT)" -u "$(DB_USER)" -p < database/schema.sql

db-seed:
	mysql --ssl-mode=VERIFY_IDENTITY -h "$(DB_HOST)" -P "$(DB_PORT)" -u "$(DB_USER)" -p "$(DB_NAME)" < database/seeds.sql

bootstrap-admin:
	go run ./cmd/bootstrap-admin
