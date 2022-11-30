BASE_IMAGE_NAME := "antonjah/static"
DEV_IMAGE_TAG := "dev"
LATEST_IMAGE_TAG := "latest"
LATEST_GIT_TAG := $(shell git describe --tags `git rev-list --tags --max-count=1`)

.PHONY: build
build: ## Build static
	@CGO_ENABLED=0 go build -ldflags="-w -s" -o static cmd/static/static.go

.PHONY: test
test: lint unittest ## Run linting and unittest

.PHONY: unittest
unittest: ## Run unittest
	go test -cover ./...

.PHONY: lint
lint: ## Run linting
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@LOG_LEVEL=error golangci-lint run

.PHONY: buildimage
buildimage: ## Build docker image
	@docker build --quiet -t "${BASE_IMAGE_NAME}:${DEV_IMAGE_TAG}" .
	@echo built "${BASE_IMAGE_NAME}:${DEV_IMAGE_TAG}"

.PHONY: tagimage
tagimage: buildimage ## Tag the docker image
	@docker tag "${BASE_IMAGE_NAME}:${DEV_IMAGE_TAG}" "${BASE_IMAGE_NAME}:${LATEST_GIT_TAG}"
	@docker tag "${BASE_IMAGE_NAME}:${DEV_IMAGE_TAG}" "${BASE_IMAGE_NAME}:${LATEST_IMAGE_TAG}"
	@echo tagged "${BASE_IMAGE_NAME}:${LATEST_GIT_TAG}"
	@echo tagged "${BASE_IMAGE_NAME}:${LATEST_IMAGE_TAG}"

.PHONY: help
help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
