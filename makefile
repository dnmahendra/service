SHELL := /bin/bash

# ==============================================================================
# Building containers

# $(shell git rev-parse --short HEAD)
VERSION := 1.0
# VERSION := `git rev-parse --short HEAD`

all: sales-api

sales-api:
	docker build \
		-f zarf/docker/dockerfile.sales-api \
		-t sales-api-amd64:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ===============================================================================
# Running from within k8s/dev

KIND_CLUSTER := falcon

kind-up:
	kind create cluster \
		--image kindest/node:v1.21.1 \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=sales-system

kind-down:
	kind delete cluster --name $(KIND_CLUSTER)

kind-load:
	cd zarf/k8s/kind/sales-pod; kustomize edit set image sales-api-image=sales-api-amd64:$(VERSION)
	kind load docker-image sales-api-amd64:$(VERSION) --name $(KIND_CLUSTER)

kind-apply:
	kustomize build zarf/k8s/kind/sales-pod | kubectl apply -f -

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods --watch --all-namespaces

kind-status-sales:
	kubectl get pods -o wide --watch

kind-logs:
	kubectl logs -lapp=sales --all-containers=true -f --tail=100 | go run app/tooling/logfmt/main.go

kind-sales-api: sales-api
	kind load docker-image sales-api-amd64:1.0 --name falcon
	kubectl delete pods -lapp=sales-api

kind-update: all kind-load kind-apply

kind-update-apply: all kind-load kind-apply

kind-describe:
	kubectl describe pod -l app=sales

# =============================================================================
run:
	go run app/services/sales-api/main.go | go run app/tooling/logfmt/main.go

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
