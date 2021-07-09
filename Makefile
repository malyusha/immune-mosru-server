.PHONY: generate download test run-local

generate:
	go generate ./...

download:
	go mod download

test:
	CGO_ENABLED=0 go test ./...

build:
	CGO_ENABLED=0 go build \
		-ldflags "-s -X 'main.Version=$(VERSION)' -X 'main.Name=$(NAME)'" \
		-v -o $(NAME) ./cmd/$(NAME)

run-local:
	go run ./cmd/immune -values=configs/values-local.json
