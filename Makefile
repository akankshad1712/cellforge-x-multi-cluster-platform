.PHONY: run test build fmt vet
run:
	go run ./cmd/controlplane
test:
	go test ./... -race -cover
build:
	mkdir -p bin && go build -o bin/cellforge ./cmd/controlplane
fmt:
	gofmt -w ./cmd ./internal
vet:
	go vet ./...
