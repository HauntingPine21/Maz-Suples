.PHONY: run test vet build fmt bootstrap-admin

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

bootstrap-admin:
	go run ./cmd/bootstrap-admin
