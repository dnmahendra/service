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
		--build-arg VCS_REF=`git rev-parse --short HEAD` \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ===============================================================================
# Running from within k8s/dev

kind-up:
	kind create cluster --image kindest/node:v1.19.1 --name falcon --config zarf/k8s/dev/kind-config.yaml

kind-down:
	kind delete cluster --name falcon

kind-load:
	kind load docker-image sales-api-amd64:1.0 --name falcon

kind-services:
	kustomize build zarf/k8s/dev | kubectl apply -f -

kind-status:
	kubectl get nodes
	kubectl get pods --watch

kind-status-full:
	kubectl describe pod -lapp=sales-api

kind-logs:
	kubectl logs -lapp=sales-api --all-containers=true -f

kind-sales-api: sales-api
	kind load docker-image sales-api-amd64:1.0 --name falcon
	kubectl delete pods -lapp=sales-api

# =============================================================================
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
