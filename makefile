SHELL := /bin/bash


# Access metrics directly (4000) or through the sidecar (3001)
# go install github.com/divan/expvarmon@latest
# expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"
# expvarmon -ports=":3001" -endpoint="/metrics" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

####################################################################################################
run:
	go run app/services/go_service/main.go | go run app/tool/logfmt/main.go

tidy:
	go mod tidy
	go mod vendor

# build:
# 	go build -ldflags "-X main.build=local"

VERSION := 1.0

all: service

service:
	docker build \
	-f lego/docker/dockerfile.go_service \
	-t go_service-amd64:${VERSION} \
	--build-arg BUILD_REF=${VERSION} \
	--build-arg BUILD_DATE='date -u +"%d-%m-%Y  %H:%M:%S"' \
	.

#=====================================================================================================
#Kind(k8s running localy on docker) run

KIND_KLUSTER := go-service-deployment

kind-up:
	kind create cluster \
		--image kindest/node:v1.24.0@sha256:0866296e693efe1fed79d5e6c7af8df71fc73ae45e3679af05342239cdc5bc8e \
		--name ${KIND_KLUSTER} \
		--config ./lego/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=go-service

kind-down:
	kind delete cluster --name ${KIND_KLUSTER}

kind-load:
	cd lego/k8s/kind/service-pod; kustomize edit set image go_service-image=go_service-amd64:${VERSION}
	kind load docker-image go_service-amd64:${VERSION} --name ${KIND_KLUSTER}

kind-apply:
	kustomize build lego/k8s/kind/service-pod/ | kubectl apply -f -

kind-status-all:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

kind-status:
	kubectl get pods -o wide --watch 

kind-logs:
	kubectl logs -l app=go-service-app --all-containers=true -f --tail=100 | go run app/tool/logfmt/main.go

kind-describe:
	kubectl describe pod -l app=go-service-app

kind-restart:
	kubectl rollout restart deployment go-service-deployment

kind-update: all kind-load kind-restart

kind-update-apply: all kind-load kind-apply kind-restart