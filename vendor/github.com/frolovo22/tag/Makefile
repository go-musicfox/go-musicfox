CURRENT_DIR = $(shell pwd)
GREEN = \033[0;32m
YELLOW = \033[0;33m
NC = \033[0m

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'

test: ## Run unit-tests
	@echo "\n${GREEN}Running unit-tests${NC}"
	go mod download
	go test -race -v -covermode=atomic -coverpkg=./${SRC_PATH}/... -coverprofile=coverage.out ./${SRC_PATH}/...
	go tool cover -func coverage.out | grep total | awk '{print "coverage: " $$3}'

fmt: ## Auto formatting Golang code
	@echo "\n${GREEN}Auto formatting golang code with golangci-lint${NC}"
	golangci-lint run --fix
	@echo "\n${GREEN}Auto formatting golang code with gofmt${NC}"
	gofmt

lint: golangci-lint ## Linting Golang code

# ============= Other project specific commands =============
golangci-lint: ## Linting Golang code with golangci
	@echo "\n${GREEN}Linting Golang code with golangci${NC}"
	golangci-lint --version
	golangci-lint cache clean
	golangci-lint run ./... -v --timeout 240s
