.PHONY: build run test lint clean docker-build docker-up docker-down

build:
	go build -o bin/pr-review-service ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./... -v

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

clean:
	rm -rf bin/
	go clean

docker-build:
	docker-compose build

docker-up:
	docker-compose up

docker-down:
	docker-compose down

loadtest:
	go run ./tools/loadtest

check: lint test

