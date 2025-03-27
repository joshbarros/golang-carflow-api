.PHONY: build run test clean lint fmt help build-cli run-cli build-ui run-ui

BINARY_NAME=carflow
MAIN_FILE=cmd/main.go
CLI_BINARY_NAME=carflow-cli
CLI_MAIN_FILE=cmd/cli/main.go
UI_BINARY_NAME=carflow-ui
UI_MAIN_FILE=cmd/ui/main.go

help:
	@echo "Available commands:"
	@echo "  make build       - Build the application"
	@echo "  make run         - Run the application"
	@echo "  make build-cli   - Build the CLI tool"
	@echo "  make run-cli     - Run the CLI tool"
	@echo "  make build-ui    - Build the UI application"
	@echo "  make run-ui      - Run the UI application"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make lint        - Run linter"
	@echo "  make fmt         - Format code"

build:
	go build -o ${BINARY_NAME} ${MAIN_FILE}

run: build
	./${BINARY_NAME}

build-cli:
	go build -o ${CLI_BINARY_NAME} ${CLI_MAIN_FILE}

run-cli: build-cli
	./${CLI_BINARY_NAME} help

build-ui:
	go build -o ${UI_BINARY_NAME} ${UI_MAIN_FILE}

run-ui: build-ui
	./${UI_BINARY_NAME}

test:
	go test ./... -v

clean:
	go clean
	rm -f ${BINARY_NAME}
	rm -f ${CLI_BINARY_NAME}
	rm -f ${UI_BINARY_NAME}

lint:
	go vet ./...
	go run golang.org/x/lint/golint@latest ./...

fmt:
	go fmt ./... 