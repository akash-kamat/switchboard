.PHONY: test build build-arm64

test:
	go test ./...
	go vet ./...

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/switchboard ./cmd/switchboard

build-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/switchboard-linux-arm64 ./cmd/switchboard
