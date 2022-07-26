SHELL := /bin/bash

run:
	go run main.go

tidy:
	go mod tidy
	go mod vendor

# build:
# 	go build -ldflags "-X main.build=local"

VERSION := 1.0

all: service

service:
	docker build \
	-f lego/docker/dockerfile \
	-t service-amd64:${VERSION} \
	--build-arg BUILD_REF=${VERSION} \
	--build-arg BUILD_DATE='date -u +"%d-%m-%Y  %H:%M:%S"' \
	.

#=====================================================================================================
#Kind(k8s running localy on docker) run

KIND_KLUSTER := go-service-deployment

kind-up:
	kind create cluster \
		--image kindest/node:v1.24.0 \
		--name ${KIND_KLUSTER} \
		--config ./lego/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=go-service

kind-down:
	kind delete cluster --name ${KIND_KLUSTER}

kind-load:
	kind load docker-image service-amd64:${VERSION} --name ${KIND_KLUSTER}

kind-apply:
	kustomize build lego/k8s/kind/service-pod/ | kubectl apply -f -

kind-status-all:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

kind-status:
	kubectl get pods -o wide --watch 

kind-logs:
	kubectl logs -l app=go_service --all-containers=true -f --tail=100 

kind-describe:
	kubectl describe pod -l app=go_service

kind-restart:
	kubectl rollout restart deployment go-service-deployment

kind-update: all kind-load kind-restart

kind-update-apply: all kind-load kind-apply