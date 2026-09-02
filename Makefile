.PHONY: all build test test-race vet benchmark clean run

BINARY_NAME=porta

all: test vet build

build:
	go build -o $(BINARY_NAME) ./cmd/porta

build-windows:
	GOOS=windows GOARCH=amd64 go build -o dist/porta-windows-amd64.exe ./cmd/porta

build-linux:
	GOOS=linux GOARCH=amd64 go build -o dist/porta-linux-amd64 ./cmd/porta

build-darwin:
	GOOS=darwin GOARCH=arm64 go build -o dist/porta-darwin-arm64 ./cmd/porta

test:
	go test ./... -v

test-race:
	go test ./... -v -race

vet:
	go vet ./...

benchmark:
	go test ./internal/proxy -bench=BenchmarkProxyRoutingLatency -run=^$ -benchmem

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	rm -rf dist/ .porta/
