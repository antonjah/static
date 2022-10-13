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
	docker build -t antonjah/static:dev .

.PHONY: help
help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
