.PHONY: deps clean build lint changelog snapshot release

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

GO ?= latest

BINARY = ghub
MAIN = $(shell pwd)/cmd/main.go

BUILDDIR = $(shell pwd)/build
VERSION ?= 0.0.1
GITREV = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +'%FT%TZ%z')
LDFLAGS=-ldflags '-X main.version=${VERSION} -X main.commit=${GITREV} -X main.date=${BUILDTIME}'
GO_BUILDER_VERSION=v1.15

default: build

deps:
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	go get -u github.com/goreleaser/goreleaser
	go get -u github.com/git-chglog/git-chglog/cmd/git-chglog
	go get -u golang.org/x/tools/cmd/goimports

build:
	go build ${LDFLAGS} -o $(BUILDDIR)/${BINARY} -i $(MAIN)
	@echo 'Build $(BINARY) done.'

changelog:
	git-chglog $(VERSION) > CHANGELOG.md

clean:
	rm -rf $(BUILDDIR)/

lint:
	golangci-lint run --fix

style:
	gofmt -w .
	goimports -local github.com/qlcchain/qlc-hub -w .

snapshot:
	docker run --rm --privileged \
		-e PRIVATE_KEY=$(PRIVATE_KEY) \
		-v $(CURDIR):/qlc-hub \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(GOPATH)/src:/go/src \
		-w /qlc-hub \
		goreng/golang-cross:$(GO_BUILDER_VERSION) --snapshot --rm-dist

release: changelog
	docker run --rm --privileged \
		-e GITHUB_TOKEN=$(GITHUB_TOKEN) \
		-e PRIVATE_KEY=$(PRIVATE_KEY) \
		-v $(CURDIR):/qlc-hub \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(GOPATH)/src:/go/src \
		-w /qlc-hub \
		goreng/golang-cross:$(GO_BUILDER_VERSION) --rm-dist --release-notes=CHANGELOG.md
