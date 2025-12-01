.PHONY: run fmt tidy

run:
	go run ./cmd/server

fmt:
	go fmt ./...

tidy:
	go mod tidy