GO_TEST_FLAGS?=
GO_TEST_PKGS?=$(shell go list ./...)

.PHONY: test
test:
	@go test ${GO_TEST_FLAGS} -coverprofile=coverage.out ${GO_TEST_PKGS}
	@go tool cover -func=coverage.out | grep total

.PHONY: fmt
fmt:
	go fmt  ./...