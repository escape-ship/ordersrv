all: build

init:
	@echo "Initializing..."
	@$(MAKE) tool_download
	@$(MAKE) build

build:
	@echo "Building..."
	@go mod tidy
	@go mod download
	@$(MAKE) proto_gen
	@$(MAKE) sqlc_gen
	@go build -o bin/$(shell basename $(PWD)) ./cmd

build_alone:
	@go build -tags migrate -o bin/$(shell basename $(PWD)) ./cmd

proto_gen:
	@echo "Generating proto..."
	@cd proto && \
	buf dep update && \
	buf generate

sqlc_gen:
	@echo "Generating sqlc..."
	@cd internal/infra/sqlc && \
	sqlc generate

tool_download:
	$(MAKE) protoc_download
	$(MAKE) buf_download
	$(MAKE) sqlc_download

protoc_download:
	@echo "Downloading protoc..."
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

sqlc_download:
	@echo "Downloading sqlc..."
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

buf_download:
	@echo "Downloading buf..."
	@go install github.com/bufbuild/buf/cmd/buf@latest

run:
	@echo "Running..."
	@./bin/$(shell basename $(PWD))

linter-golangci: ### check by golangci linter
	golangci-lint run
.PHONY: linter-golangci

clean:
	rm -f bin/$(shell basename $(PWD))