BASE_IMAGE_NAME := "antonjah/static"
DEV_IMAGE_TAG := "dev"
LATEST_IMAGE_TAG := "latest"
LATEST_GIT_TAG := $(shell git describe --tags `git rev-list --tags --max-count=1`)

.PHONY: build
build: ## Build static service binary
	@CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/static cmd/static/static.go

.PHONY: build-operator
build-operator: ## Build operator binary
	@CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/operator cmd/operator/main.go

.PHONY: build-all
build-all: build build-operator ## Build all binaries

.PHONY: test
test: lint unittest ## Run linting and unittest

.PHONY: unittest
unittest: ## Run unittest
	go test -cover ./...

.PHONY: lint
lint: ## Run linting
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@LOG_LEVEL=error golangci-lint run

.PHONY: generate
generate: ## Generate CRD manifests
	controller-gen crd paths=./pkg/apis/... output:crd:dir=./deployments/crds

.PHONY: docker-static
docker-static: ## Build static service Docker image
	@docker build --quiet -t "static:${LATEST_IMAGE_TAG}" -f deployments/docker/Dockerfile.static .
	@echo built "static:${LATEST_IMAGE_TAG}"

.PHONY: docker-operator
docker-operator: ## Build operator Docker image
	@docker build --quiet -t "static-operator:${LATEST_IMAGE_TAG}" -f deployments/docker/Dockerfile.operator .
	@echo built "static-operator:${LATEST_IMAGE_TAG}"

.PHONY: docker-all
docker-all: docker-static docker-operator ## Build all Docker images

.PHONY: deploy-crds
deploy-crds: ## Deploy CRDs to Kubernetes
	kubectl apply -f deployments/crds/

.PHONY: deploy
deploy: deploy-crds ## Deploy operator and static service to Kubernetes
	kubectl apply -f deployments/namespace.yaml
	kubectl apply -f deployments/rbac.yaml
	kubectl apply -f deployments/configmap.yaml
	kubectl apply -f deployments/operator-deployment.yaml
	kubectl apply -f deployments/examples/static-example.yaml

.PHONY: undeploy
undeploy: ## Remove operator and static service from Kubernetes
	kubectl delete -f deployments/examples/static-example.yaml --ignore-not-found
	kubectl delete -f deployments/operator-deployment.yaml --ignore-not-found
	kubectl delete -f deployments/configmap.yaml --ignore-not-found
	kubectl delete -f deployments/rbac.yaml --ignore-not-found
	kubectl delete -f deployments/namespace.yaml --ignore-not-found
	kubectl delete -f deployments/crds/ --ignore-not-found

.PHONY: buildimage
buildimage: docker-static ## Build docker image (legacy target)

.PHONY: tagimage
tagimage: buildimage ## Tag the docker image
	@docker tag "${BASE_IMAGE_NAME}:${DEV_IMAGE_TAG}" "${BASE_IMAGE_NAME}:${LATEST_GIT_TAG}"
	@docker tag "${BASE_IMAGE_NAME}:${DEV_IMAGE_TAG}" "${BASE_IMAGE_NAME}:${LATEST_IMAGE_TAG}"
	@echo tagged "${BASE_IMAGE_NAME}:${LATEST_GIT_TAG}"
	@echo tagged "${BASE_IMAGE_NAME}:${LATEST_IMAGE_TAG}"

.PHONY: help
help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
