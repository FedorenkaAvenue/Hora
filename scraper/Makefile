OK_RESULT= echo "Done😌!"

clear_log:
	@echo "Clearing log files..."
	@rm -f *.log && ${OK_RESULT}

help: # Show Makefile commands
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m  %-30s\033[0m %s\n", $$1, $$2}'
