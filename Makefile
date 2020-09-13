.PHONY: deps clean build lint changelog snapshot release signer signer-proto

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH")))

GO ?= latest

BINARY = ghub
MAIN = $(shell pwd)/cmd/main.go

BUILDDIR = $(shell pwd)/build
VERSION ?= 0.0.1
GITREV = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +'%FT%TZ%z')
LDFLAGS=-ldflags '-X main.version=${VERSION} -X main.commit=${GITREV} -X main.date=${BUILDTIME}'
GO_BUILDER_VERSION=v1.15.1

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
	@cat assets/footer.txt >> CHANGELOG.md

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
		goreng/golang-cross:$(GO_BUILDER_VERSION) --rm-dist --debug --release-notes=CHANGELOG.md

signer-proto:
	protoc -I$(GOPATH)/src -I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I=$(shell pwd)/grpc/proto/ --go_out=plugins=grpc:$(shell pwd)/grpc/proto $(shell pwd)/grpc/proto/signer.proto

signer:
	go build ${LDFLAGS} -o $(BUILDDIR)/signer -i $(shell pwd)/cmd/signer/main.go
	@echo 'Build signer done.'
