NAME			:= fs-store
VERSION			:= v0.1-beta
REVISION		:= $(shell git rev-parse --short HEAD)
LDFLAGS 		:= "-s -w -X main.Version=$(VERSION) -X main.Revision=$(REVISION)"
OSARCH			:= "linux/arm64 linux/amd64 darwin/arm64 darwin/amd64 windows/amd64"


ifndef GOBIN
GOBIN := $(shell echo "$${GOPATH%%:*}/bin")
endif

LINT := $(GOBIN)/golint
GOX := $(GOBIN)/gox

$(LINT): ; @go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.44.0
$(GOX): ; @go get github.com/mitchellh/gox

.DEFAULT_GOAL := build

.PHONY: docker-build
docker-build:
	docker build -t $(NAME):$(VERSION) .

.PHONY: build
build:
	go build -ldflags $(LDFLAGS) -o bin/$(NAME)

.PHONY: install
install:
	go install -ldflags $(LDFLAGS)

.PHONY: release
release: $(GOX)
	rm -rf ./out && \
	gox -ldflags $(LDFLAGS) -osarch $(OSARCH) -parallel=2 \
	-output "./dist/${NAME}_${VERSION}_{{.OS}}_{{.Arch}}"

.PHONY: lint
lint: $(LINT)
	@golangci-lint run ./...

.PHONY: vet
vet:
	@go vet ./...

.PHONY: test
test:
	@go test ./... -v

.PHONY: check
check: lint vet test build