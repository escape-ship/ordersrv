all: build

build:
	@echo "Building..."
	@go mod tidy
	@go mod download
	@$(MAKE) proto_gen
	@go build -o bin/$(shell basename $(PWD)) ./cmd

proto_gen:
	@echo "Generating proto..."
	@cd proto && \
	buf dep update && \
	buf generate

run:
	@echo "Running..."
	@./bin/$(shell basename $(PWD))


linter-golangci: ### check by golangci linter
	golangci-lint run
.PHONY: linter-golangci