.PHONY: build run test clean docker-build docker-up docker-down docker-logs deps lint fmt swagger test-unit test-cache test-coverage

build:
	go build -o bin/address-validator ./cmd/api

run:
	go run ./cmd/api/main.go

test:
	go test -v ./...

test-unit:
	go test -v -short ./...

test-cache:
	go test -v ./internal/services/ -run TestCache

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean:
	rm -rf bin/
	rm -f coverage.out

deps:
	go mod download
	go mod tidy

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down


lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	gofmt -s -w .

swagger:
	swag init -g cmd/api/main.go -o docs
