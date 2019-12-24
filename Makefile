GO_TEST_FLAGS?=-race

.PHONY: test
test:
	@go test ./... ${GO_TEST_FLAGS} -coverprofile=coverage.txt covermode=atomic
	@go tool cover -func=coverage.txt | grep total

.PHONY: fmt
fmt:
	go fmt  ./...