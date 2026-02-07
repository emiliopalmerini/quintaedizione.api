.PHONY: fmt vet build run test test-e2e bench bench-save bench-compare loadtest docker-up docker-down clean

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

bench:
	go test -bench=. -benchmem ./internal/...

bench-save:
	go test -bench=. -benchmem -count=6 ./internal/... > bench.txt

bench-compare:
	go test -bench=. -benchmem -count=6 ./internal/... > bench_new.txt
	benchstat bench.txt bench_new.txt

loadtest:
	./scripts/loadtest.sh

docker-up:
	docker-compose up -d postgres

docker-down:
	docker-compose down

clean:
	rm -rf bin/
