run:
	@go run $(GO_ARGS) ./cmd/comet $(APP_ARGS)

build:
	@go build -o main $(GO_ARGS) ./cmd/comet $(APP_ARGS)