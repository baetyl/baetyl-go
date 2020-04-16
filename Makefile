GO_TEST_FLAGS?=-race -short -covermode=atomic -coverprofile=coverage.txt
GO_TEST_PKGS?=$(shell go list ./...)

.PHONY: test
test:
	env GO111MODULE=on GOPROXY=https://goproxy.cn go test ${GO_TEST_FLAGS} ${GO_TEST_PKGS}

.PHONY: fmt format
fmt: format
format:
	go fmt ${GO_TEST_PKGS}