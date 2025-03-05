all: build

init:
	@echo "Initializing..."
	@$(MAKE) sqlc_download
	@$(MAKE) buf_download

build:
	@echo "Building..."
	@go mod tidy
	@go mod download
	@$(MAKE) proto_gen
	@$(MAKE) sqlc_gen
	@go build -o bin/$(shell basename $(PWD)) ./cmd

build_alone:
	@go build -o bin/$(shell basename $(PWD)) ./cmd

proto_gen:
	@echo "Generating proto..."
	@cd proto && \
	buf dep update && \
	buf generate

sqlc_download:
	@echo "Downloading sqlc..."
	@go install ithub.com/sqlc-dev/sqlc/cmd/sqlc@latest

buf_download:
	@echo "Downloading buf..."
	@go install github.com/bufbuild/buf/cmd/buf@latest

sqlc_gen:
	@echo "Generating sqlc..."
	@sqlc generate

run:
	@echo "Running..."
	@./bin/$(shell basename $(PWD))

linter-golangci: ### check by golangci linter
	golangci-lint run
.PHONY: linter-golangci