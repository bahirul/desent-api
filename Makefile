.PHONY: dev build test test-cover vet fmt tidy air-install

dev:
	go tool air -c .air.toml

air-install:
	go install github.com/air-verse/air@v1.64.5

build:
	go build -o bin/api ./cmd/api

test:
	go test ./...

test-cover:
	go test -cover ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy
