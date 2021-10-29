export GOLDFLAGS=-s -w -extldflags '-zrelro -znow'
export GOFLAGS=-trimpath
export CGO_ENABLED=0

.PHONY: all
all: dist

.PHONY: dist
dist: dist/amd64/healthcheck dist/arm64/healthcheck

.PHONY: dist/amd64/healthcheck
dist/amd64/healthcheck:
	GOARCH=amd64 go build -ldflags "$(GOLDFLAGS)" -o $@ ./cmd/healthcheck

.PHONY: dist/arm64/healthcheck
dist/arm64/healthcheck:
	GOARCH=arm64 go build -ldflags "$(GOLDFLAGS)" -o $@ ./cmd/healthcheck

.PHONY: generate
generate:
	go generate ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	gofumpt -s -w .

.PHONY: update
update:
	go get -t -u ./...
