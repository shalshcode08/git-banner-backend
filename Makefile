APP_NAME  = git-banner-backend
CMD_PATH  = ./cmd/server
BIN_DIR   = bin
BIN       = $(BIN_DIR)/$(APP_NAME)

.PHONY: all build fmt run dev clean test tidy lint help

all: build

## build: compile the binary into bin/
build: fmt
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN) $(CMD_PATH)
	@echo "built -> $(BIN)"

## fmt: format all Go source files
fmt:
	gofmt -w .

## run: build and run the server
run: build
	./$(BIN)

## dev: run without building to a bin (uses go run)
dev:
	go run $(CMD_PATH)/main.go

## test: run all tests
test:
	go test ./... -v

## tidy: tidy and verify go modules
tidy:
	go mod tidy
	go mod verify

## lint: run go vet
lint:
	go vet ./...

## clean: remove compiled binaries
clean:
	rm -rf $(BIN_DIR)
	@echo "cleaned"

## help: show this help message
help:
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## /  /'
