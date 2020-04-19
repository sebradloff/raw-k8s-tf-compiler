.PHONY: help

default: help


help: ## Show this help
	@echo "rawk8stfc"
	@echo "======================"
	@echo
	@echo "A cli tool to create tf resources for all k8s objects inputed"
	@echo
	@fgrep -h " ## " $(MAKEFILE_LIST) | fgrep -v fgrep | sed -Ee 's/([a-z.]*):[^#]*##(.*)/\1##\2/' | column -t -s "##"

build: ## build the binary
	go build -o rawk8stfc main.go

