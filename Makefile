BINARY_NAME := xbe
PKG := ./...
VERSION ?= dev

.PHONY: build
build:
	go build -ldflags "-X github.com/xbe-inc/xbe-cli/internal/version.Version=$(VERSION)" -o $(BINARY_NAME) ./cmd/xbe

.PHONY: test
test:
	go test $(PKG)

.PHONY: fmt
fmt:
	go fmt $(PKG)

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	@echo "lint not configured (intentionally)."
