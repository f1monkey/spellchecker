.PHONY: lint test

COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)

-include .env
export CGO_CFLAGS
export CGO_LDFLAGS

# Support shell-style quoted values in .env.
CGO_CFLAGS := $(subst ",,$(CGO_CFLAGS))
CGO_LDFLAGS := $(subst ",,$(CGO_LDFLAGS))

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2 run

test:
	go test ./... --race