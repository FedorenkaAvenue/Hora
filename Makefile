OK_RESULT= echo "Done😌!"
CMD_ROOT_FILE_PATH=./main.go
BUILD_FILE_PATH=./dist/hora

install: ## Install all dev dependencies
	@echo "Install all dev dependencies..."
	@sh ./scripts/install.sh && ${OK_RESULT}

build: ## Build application
	@echo "Building application..."
	@go build -o ${BUILD_FILE_PATH} $(CMD_ROOT_FILE_PATH) && ${OK_RESULT}
	@cp ./config.yaml ./dist/config.yaml

run: ## Run application without build
	@go run ${BUILD_FILE_PATH}

dev: ## Run build during any file changing (dev live mode)
	@echo "Starting dev demon server..."
	ulimit -n 1000
	reflex --start-service -r '\.go$$' make build && make run

help: # Show Makefile commands
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m  %-30s\033[0m %s\n", $$1, $$2}'
