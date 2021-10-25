SHELL := /bin/bash

# ==============================================================================
# Building containers

# $(shell git rev-parse --short HEAD)
VERSION := 1.0

all: sales-api

sales-api:
	docker build \
		-f zarf/docker/dockerfile.sales-api \
		-t sales-api-amd64:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

run:
	go run app/sales-api/main.go

runa:
	go run app/admin/main.go

tidy:
	go mod tidy
	go mod vendor

# ==============================================================================
# Running tests within the local computer

test:
	go test -v ./... -count=1
	staticcheck ./...

# ==============================================================================
