HOMEDIR := $(shell pwd)
OUTDIR  := $(HOMEDIR)/output

GIT_TAG:=$(shell git tag --contains HEAD)
GIT_REV:=git-$(shell git rev-parse --short HEAD)
VERSION:=$(if $(GIT_TAG),$(GIT_TAG),$(GIT_REV))

GO       = go
GO_MOD   = $(GO) mod
GO_ENV   = env CGO_ENABLED=0
GO_BUILD = $(GO_ENV) $(GO) build
GOTEST   = $(GO) test
GOPKGS   = $$($(GO) list ./...)

all: test

prepare: prepare-dep
prepare-dep:
	git config --global http.sslVerify false

set-env:
	$(GO) env -w GOPROXY=https://goproxy.cn
	$(GO) env -w GONOSUMDB=\*

compile:build
build: set-env
	$(GO_MOD) tidy
	$(GO_BUILD) ./...

test: fmt test-case
test-case: set-env
	$(GOTEST) -race -cover -coverprofile=coverage.out $(GOPKGS)

fmt:
	go fmt ./...

.PHONY: all prepare compile test build
