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

build-test-image: ## build the docker test image
	docker build -f Dockerfile.test . -t integration-tests

integration-tests: build-test-image ## validate golden files work with terraform-provider-k8s
	docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock integration-tests

shell: build-test-image ## run the integration-tests container and interact with it
	docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock --entrypoint bash integration-tests

coverage: ## show test coverage in browser
	go test -coverprofile=cover.out ./... && go tool cover -html=cover.out