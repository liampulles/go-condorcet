# Init variables
GOBIN := $(shell go env GOBIN)

# Keep test at the top so that it is default when `make` is called.
# This is used by Travis CI.
coverage.txt:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
view-cover: clean coverage.txt
	go tool cover -html=coverage.txt
test: build
	go test ./...
build:
	go build ./...
install: build
	go install ./...
inspect: build $(GOBIN)/golangci-lint
	golangci-lint run
update:
	go get -u ./...
pre-commit: update clean coverage.txt inspect
	go mod tidy
clean:
	rm -f coverage.txt $(GOBIN)/condorcli

# Needed tools
$(GOBIN)/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN)/bin v1.41.1