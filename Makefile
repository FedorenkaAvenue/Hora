OK_RESULT= echo "Done😌!"
CMD_ROOT_FILE_PATH=./main.go

build: ## Build application
	@echo "Building application..."
	@go build -o dist/hora $(CMD_ROOT_FILE_PATH) && ${OK_RESULT}
	@cp ./configs/bot.yaml ./dist/bot.yaml

run: ## Run application without build
	@go run $(CMD_ROOT_FILE_PATH)

deploy: ## Deploy compiled application to your `/home/$USER/.local/bin` folder
	@echo "Moving application to /home/USER/.local/bin/ folder..."
	@sudo cp ./dist/andy /home/${USER}/.local/bin/ && ${OK_RESULT}

dev: ## Run build during any file changing (dev live mode)
	@echo "Restarting..."
	ulimit -n 1000
	${GOPATH}/bin/reflex --start-service -r '\.go$$' make build

help: # Show Makefile commands
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m  %-30s\033[0m %s\n", $$1, $$2}'
