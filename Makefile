.PHONY: fmt vet build run test test-e2e docker-up docker-down clean

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build -o bin/api ./cmd/api

run: build docker-up
	./bin/api

test:
	go test ./internal/...

test-e2e:
	go test -tags e2e -v .

docker-up:
	docker-compose up -d postgres

docker-down:
	docker-compose down

clean:
	rm -rf bin/
