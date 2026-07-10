.PHONY: build test generate

build:
	go build -o build/watchue ./cmd/web

test:
	go test ./cmd/... ./internal/... -v

generate:
	sqlc generate
