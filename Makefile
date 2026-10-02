GO_PACKAGES := ./api ./app ./cmd/... ./database ./internal/...

.PHONY: run test vet build fmt fmt-check lint coverage verify test-e2e db-init db-seed bootstrap-admin

run:
	go run ./cmd/server

test:
	go test $(GO_PACKAGES)

vet:
	go vet $(GO_PACKAGES)

build:
	go build -o bin/maz-suplementos ./cmd/server

fmt:
	gofmt -w api app cmd database internal

fmt-check:
	@test -z "$$(gofmt -l api app cmd database internal)" || (gofmt -l api app cmd database internal && exit 1)

lint: fmt-check vet
	npm run lint

coverage:
	go test -coverprofile=coverage.out $(GO_PACKAGES)
	npm run check:coverage

test-e2e:
	npm run test:e2e
	npm run test:a11y

verify: lint test coverage build
	npm run check:size
	npm run check:licenses

db-init:
	mysql --ssl-mode=VERIFY_IDENTITY -h "$(DB_HOST)" -P "$(DB_PORT)" -u "$(DB_USER)" -p < database/schema.sql

db-seed:
	mysql --ssl-mode=VERIFY_IDENTITY -h "$(DB_HOST)" -P "$(DB_PORT)" -u "$(DB_USER)" -p "$(DB_NAME)" < database/seeds.sql

bootstrap-admin:
	go run ./cmd/bootstrap-admin
