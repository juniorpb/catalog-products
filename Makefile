.PHONY: setup run test

setup:
	go mod download

run:
	go run cmd/main.go

test:
	go test ./... -v
